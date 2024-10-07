package diff

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cstanislawski/helm-valgrade/internal/chart"
)

type Diff struct {
	Added    map[string]interface{}
	Removed  map[string]interface{}
	Modified map[string]interface{}
}

func Compare(base, target *chart.Chart, userValues map[string]interface{}, keepValues []string, ignoreMissing bool) (*Diff, error) {
	userValuesDiff, err := compareValues("", base.NormalizedValues, userValues, true)
	if err != nil {
		return nil, fmt.Errorf("failed to compare user values: %w", err)
	}

	if keepValues != nil {
		mergeUserChangesKeepValues(userValuesDiff, base.NormalizedValues, keepValues)
	}

	baseValuesDiff, err := compareValues("", base.NormalizedValues, target.NormalizedValues, false)
	if err != nil {
		return nil, fmt.Errorf("failed to compare base values: %w", err)
	}

	if ignoreMissing && len(baseValuesDiff.Removed) > 0 {
		removeMissingKeys(baseValuesDiff.Removed, userValuesDiff)
	}

	return userValuesDiff, nil
}

func compareValues(prefix string, base, compare map[string]interface{}, isUserValues bool) (*Diff, error) {
	diff := &Diff{
		Added:    make(map[string]interface{}),
		Removed:  make(map[string]interface{}),
		Modified: make(map[string]interface{}),
	}

	for key, baseValue := range base {
		path := joinPath(prefix, key)
		compareValue, exists := compare[key]

		if !exists {
			if !isUserValues {
				diff.Removed[path] = baseValue
			}
			continue
		}

		if !reflect.DeepEqual(baseValue, compareValue) {
			switch baseTyped := baseValue.(type) {
			case map[string]interface{}:
				if compareTyped, ok := compareValue.(map[string]interface{}); ok {
					subDiff, err := compareValues(path, baseTyped, compareTyped, isUserValues)
					if err != nil {
						return nil, fmt.Errorf("error comparing nested map at %s: %w", path, err)
					}
					mergeDiffs(diff, subDiff)
				} else {
					diff.Modified[path] = compareValue
				}
			default:
				diff.Modified[path] = compareValue
			}
		}
	}

	for key, compareValue := range compare {
		if _, exists := base[key]; !exists {
			path := joinPath(prefix, key)
			diff.Added[path] = compareValue
		}
	}

	return diff, nil
}

func mergeUserChangesKeepValues(userDiff *Diff, baseValues map[string]interface{}, keepValues []string) {
	addKeepValuesToModified(userDiff, baseValues, keepValues, "")

	for k := range userDiff.Removed {
		if shouldKeep(k, keepValues) {
			delete(userDiff.Removed, k)
		}
	}
}

func addKeepValuesToModified(diff *Diff, values map[string]interface{}, keepValues []string, prefix string) {
	for k, v := range values {
		path := joinPath(prefix, k)
		if shouldKeep(path, keepValues) {
			if _, exists := diff.Modified[path]; !exists {
				diff.Modified[path] = v
			}
		}
		if nestedMap, ok := v.(map[string]interface{}); ok {
			addKeepValuesToModified(diff, nestedMap, keepValues, path)
		}
	}
}

func mergeDiffs(main, sub *Diff) {
	for k, v := range sub.Added {
		main.Added[k] = v
	}
	for k, v := range sub.Removed {
		main.Removed[k] = v
	}
	for k, v := range sub.Modified {
		main.Modified[k] = v
	}
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func shouldKeep(path string, keepValues []string) bool {
	for _, keep := range keepValues {
		if path == keep || strings.HasPrefix(path, keep+".") {
			return true
		}
	}
	return false
}

func PrintDiff(baseValuesDiff, userValuesDiff *Diff) {
	fmt.Println("Differences between base old version and target version:")
	printDiffSection(baseValuesDiff)

	fmt.Println("\nDifferences between base old version and user values:")
	printDiffSection(userValuesDiff)
}

func printDiffSection(diff *Diff) {
	if len(diff.Added) > 0 {
		fmt.Println("  Added:")
		for k, v := range diff.Added {
			fmt.Printf("    %s: %v\n", k, v)
		}
	}

	if len(diff.Removed) > 0 {
		fmt.Println("  Removed:")
		for k := range diff.Removed {
			fmt.Printf("    %s\n", k)
		}
	}

	if len(diff.Modified) > 0 {
		fmt.Println("  Modified:")
		for k, v := range diff.Modified {
			fmt.Printf("    %s: %v\n", k, v)
		}
	}

	if len(diff.Added) == 0 && len(diff.Removed) == 0 && len(diff.Modified) == 0 {
		fmt.Println("  No differences found.")
	}
}

func removeMissingKeys(removedKeys map[string]interface{}, userValuesDiff *Diff) {
	for key := range removedKeys {
		removeKeyAndChildren(key, userValuesDiff.Added)
		removeKeyAndChildren(key, userValuesDiff.Modified)
	}
}

func removeKeyAndChildren(key string, diffMap map[string]interface{}) {
	for diffKey := range diffMap {
		if strings.HasPrefix(diffKey, key) || diffKey == key {
			delete(diffMap, diffKey)
		}
	}
}
