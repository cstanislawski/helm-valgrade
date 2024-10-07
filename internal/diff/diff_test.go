package diff

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cstanislawski/helm-valgrade/internal/chart"
	helmchart "helm.sh/helm/v3/pkg/chart"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name          string
		base          map[string]interface{}
		target        map[string]interface{}
		user          map[string]interface{}
		keepValues    []string
		ignoreMissing bool
		expected      Diff
	}{
		{
			name:     "Simple addition",
			base:     map[string]interface{}{"a": 1},
			target:   map[string]interface{}{"a": 1, "b": 2},
			user:     map[string]interface{}{},
			expected: Diff{Added: map[string]interface{}{"b": 2}},
		},
		{
			name:     "Simple removal",
			base:     map[string]interface{}{"a": 1, "b": 2},
			target:   map[string]interface{}{"a": 1},
			user:     map[string]interface{}{},
			expected: Diff{Removed: map[string]interface{}{"b": 2}},
		},
		{
			name:     "Simple modification",
			base:     map[string]interface{}{"a": 1},
			target:   map[string]interface{}{"a": 2},
			user:     map[string]interface{}{},
			expected: Diff{Modified: map[string]interface{}{"a": 2}},
		},
		{
			name:   "Nested changes",
			base:   map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 2}},
			target: map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 3, "d": 4}},
			user:   map[string]interface{}{},
			expected: Diff{
				Added:    map[string]interface{}{"a.d": 4},
				Modified: map[string]interface{}{"a.c": 3},
			},
		},
		{
			name:       "Keep values",
			base:       map[string]interface{}{"a": 1, "b": 2},
			target:     map[string]interface{}{"a": 2, "b": 3},
			user:       map[string]interface{}{},
			keepValues: []string{"a"},
			expected:   Diff{Modified: map[string]interface{}{"b": 3}},
		},
		{
			name:          "Ignore missing",
			base:          map[string]interface{}{"a": 1, "b": 2},
			target:        map[string]interface{}{"a": 1},
			user:          map[string]interface{}{},
			ignoreMissing: true,
			expected:      Diff{},
		},
		{
			name:     "User values",
			base:     map[string]interface{}{"a": 1},
			target:   map[string]interface{}{"a": 2},
			user:     map[string]interface{}{"a": 3},
			expected: Diff{Modified: map[string]interface{}{"a": 3}},
		},
		{
			name:   "Complex nested changes with ignore missing",
			base:   map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 2, "d": 3}, "e": 4},
			target: map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 3, "f": 5}, "g": 6},
			user:   map[string]interface{}{"a": map[string]interface{}{"c": 4}},
			expected: Diff{
				Added:    map[string]interface{}{"a.f": 5, "g": 6},
				Modified: map[string]interface{}{"a.c": 4},
			},
			ignoreMissing: true,
		},
		{
			name:       "Keep nested values",
			base:       map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 2}, "d": 3},
			target:     map[string]interface{}{"a": map[string]interface{}{"b": 2, "c": 3}, "d": 4},
			user:       map[string]interface{}{},
			keepValues: []string{"a.b"},
			expected: Diff{
				Modified: map[string]interface{}{"a.c": 3, "d": 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseChart := createMockChart(tt.base)
			targetChart := createMockChart(tt.target)

			diff, err := Compare(baseChart, targetChart, tt.user, tt.keepValues, tt.ignoreMissing)
			if err != nil {
				t.Fatalf("Compare returned an error: %v", err)
			}

			if !diffEqual(*diff, tt.expected) {
				t.Errorf("Compare diff mismatch for test case %s.\nExpected: %+v\nGot: %+v\nDetailed comparison:\n%s",
					tt.name, tt.expected, *diff, detailedComparison(*diff, tt.expected))
			}
		})
	}
}

func createMockChart(values map[string]interface{}) *chart.Chart {
	return &chart.Chart{
		Chart: &helmchart.Chart{
			Values: values,
		},
	}
}

func diffEqual(a, b Diff) bool {
	return reflect.DeepEqual(a.Added, b.Added) &&
		reflect.DeepEqual(a.Removed, b.Removed) &&
		reflect.DeepEqual(a.Modified, b.Modified)
}

func detailedComparison(got, expected Diff) string {
	details := "Detailed comparison:\n"
	details += compareMap("Added", got.Added, expected.Added)
	details += compareMap("Removed", got.Removed, expected.Removed)
	details += compareMap("Modified", got.Modified, expected.Modified)

	return details
}

func compareMap(name string, got, expected map[string]interface{}) string {
	var details string

	if (got == nil) != (expected == nil) {
		details += fmt.Sprintf("%s map nil mismatch:\n", name)
		details += fmt.Sprintf("  Got is nil: %v\n", got == nil)
		details += fmt.Sprintf("  Expected is nil: %v\n", expected == nil)
		return details
	}

	if !reflect.DeepEqual(got, expected) {
		details += fmt.Sprintf("%s map mismatch:\n", name)
		details += fmt.Sprintf("  Got:      %#v\n", got)
		details += fmt.Sprintf("  Expected: %#v\n", expected)

		for k := range got {
			if _, exists := expected[k]; !exists {
				details += fmt.Sprintf("  Extra key in got: %s\n", k)
			}
		}
		for k := range expected {
			if _, exists := got[k]; !exists {
				details += fmt.Sprintf("  Missing key in got: %s\n", k)
			}
		}

		for k, v := range got {
			if expectedV, exists := expected[k]; exists {
				if !reflect.DeepEqual(v, expectedV) {
					details += fmt.Sprintf("  Value mismatch for key %s:\n    Got:      %#v\n    Expected: %#v\n", k, v, expectedV)
				}
			}
		}
	}

	return details
}

func TestShouldKeep(t *testing.T) {
	tests := []struct {
		path       string
		keepValues []string
		expected   bool
	}{
		{"a", []string{"a"}, true},
		{"a.b", []string{"a"}, true},
		{"a.b.c", []string{"a.b"}, true},
		{"a.b", []string{"a.c"}, false},
		{"a", []string{"b"}, false},
		{"a.b.c", []string{"a.b.c"}, true},
		{"a.b.c.d", []string{"a.b.c"}, true},
		{"a.b", []string{"a.b.c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			diff := shouldKeep(tt.path, tt.keepValues)
			if diff != tt.expected {
				t.Errorf("shouldKeep(%q, %v) = %v, want %v", tt.path, tt.keepValues, diff, tt.expected)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		prefix   string
		key      string
		expected string
	}{
		{"", "a", "a"},
		{"a", "b", "a.b"},
		{"a.b", "c", "a.b.c"},
		{"", "", ""},
		{"a.b.c", "", "a.b.c."},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s+%s", tt.prefix, tt.key), func(t *testing.T) {
			diff := joinPath(tt.prefix, tt.key)
			if diff != tt.expected {
				t.Errorf("joinPath(%q, %q) = %q, want %q", tt.prefix, tt.key, diff, tt.expected)
			}
		})
	}
}
