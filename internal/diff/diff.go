package diff

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cstanislawski/helm-valgrade/internal/chart"
)

type Result struct {
	Added    map[string]interface{}
	Removed  map[string]interface{}
	Modified map[string]interface{}
}

func Compare(base, target *chart.Chart, userValues map[string]interface{}, keepValues []string, ignoreMissing bool) (*Result, error) {
	result := &Result{
		Added:    make(map[string]interface{}),
		Removed:  make(map[string]interface{}),
		Modified: make(map[string]interface{}),
	}

	baseValues := base.GetDefaultValues()
	targetValues := target.GetDefaultValues()

	userChanges := identifyUserChanges("", baseValues, userValues)

	err := compareValues("", baseValues, targetValues, userChanges, keepValues, ignoreMissing, result)
	if err != nil {
		return nil, fmt.Errorf("failed to compare values: %w", err)
	}

	cleanupEmptyMaps(result)

	return result, nil
}

func identifyUserChanges(prefix string, base, user map[string]interface{}) map[string]interface{} {
	changes := make(map[string]interface{})

	for k, v := range user {
		path := joinPath(prefix, k)
		baseVal, baseExists := base[k]

		if !baseExists {
			changes[path] = v
			continue
		}

		if reflect.TypeOf(v) != reflect.TypeOf(baseVal) {
			changes[path] = v
			continue
		}

		switch typedV := v.(type) {
		case map[string]interface{}:
			subChanges := identifyUserChanges(path, baseVal.(map[string]interface{}), typedV)
			for subK, subV := range subChanges {
				changes[subK] = subV
			}
		default:
			if !reflect.DeepEqual(v, baseVal) {
				changes[path] = v
			}
		}
	}

	return changes
}

func compareValues(prefix string, base, target, userChanges map[string]interface{}, keepValues []string, ignoreMissing bool, result *Result) error {
	for k, v := range target {
		path := joinPath(prefix, k)

		if shouldKeep(path, keepValues) {
			continue
		}

		baseVal, baseExists := base[k]
		userVal, userChanged := userChanges[path]

		if !baseExists {
			if userChanged {
				result.Added[path] = userVal
			} else {
				result.Added[path] = v
			}
			continue
		}

		if reflect.TypeOf(v) != reflect.TypeOf(baseVal) {
			if userChanged {
				result.Modified[path] = userVal
			} else {
				result.Modified[path] = v
			}
			continue
		}

		switch typedV := v.(type) {
		case map[string]interface{}:
			if baseMap, ok := baseVal.(map[string]interface{}); ok {
				err := compareValues(path, baseMap, typedV, userChanges, keepValues, ignoreMissing, result)
				if err != nil {
					return err
				}
			} else {
				result.Modified[path] = v
			}
		case []interface{}:
			if baseSlice, ok := baseVal.([]interface{}); ok {
				if !reflect.DeepEqual(typedV, baseSlice) {
					result.Modified[path] = v
				}
			} else {
				result.Modified[path] = v
			}
		default:
			if !reflect.DeepEqual(v, baseVal) {
				if userChanged {
					result.Modified[path] = userVal
				} else {
					result.Modified[path] = v
				}
			}
		}
	}

	if !ignoreMissing {
		for k, v := range base {
			path := joinPath(prefix, k)

			if shouldKeep(path, keepValues) {
				continue
			}

			if _, exists := target[k]; !exists {
				result.Removed[path] = v
			}
		}
	}

	return nil
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

func cleanupEmptyMaps(result *Result) {
	if len(result.Added) == 0 {
		result.Added = nil
	}
	if len(result.Removed) == 0 {
		result.Removed = nil
	}
	if len(result.Modified) == 0 {
		result.Modified = nil
	}
}
