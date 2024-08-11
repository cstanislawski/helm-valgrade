package diff

import (
	"fmt"
	"reflect"

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

	err := compareValues("", baseValues, targetValues, userValues, keepValues, ignoreMissing, result)
	if err != nil {
		return nil, fmt.Errorf("failed to compare values: %w", err)
	}

	cleanupEmptyMaps(result)

	return result, nil
}

func compareValues(prefix string, base, target, user map[string]interface{}, keepValues []string, ignoreMissing bool, result *Result) error {
	for k, v := range target {
		path := joinPath(prefix, k)

		if shouldKeep(path, keepValues) {
			continue
		}

		baseVal, baseExists := base[k]
		userVal, userExists := user[k]

		if !baseExists {
			if userExists {
				result.Added[path] = userVal
			} else {
				result.Added[path] = v
			}
			continue
		}

		if reflect.TypeOf(v) != reflect.TypeOf(baseVal) {
			if userExists {
				result.Modified[path] = userVal
			} else {
				result.Modified[path] = v
			}
			continue
		}

		switch typedV := v.(type) {
		case map[string]interface{}:
			err := compareValues(path, baseVal.(map[string]interface{}), typedV, getUserSubMap(user, k), keepValues, ignoreMissing, result)
			if err != nil {
				return err
			}
		default:
			if !reflect.DeepEqual(v, baseVal) {
				if userExists {
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

func getUserSubMap(user map[string]interface{}, key string) map[string]interface{} {
	if user == nil {
		return nil
	}
	if subMap, ok := user[key].(map[string]interface{}); ok {
		return subMap
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
		if path == keep || (len(path) > len(keep) && path[:len(keep)] == keep && path[len(keep)] == '.') {
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
