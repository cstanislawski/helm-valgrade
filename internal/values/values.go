package values

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var values map[string]interface{}
	switch filepath.Ext(filename) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &values)
	case ".json":
		err = json.Unmarshal(data, &values)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filepath.Ext(filename))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal values: %w", err)
	}

	return values, nil
}

func Write(filename string, values map[string]interface{}) error {
	var data []byte
	var err error

	switch filepath.Ext(filename) {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(values)
	case ".json":
		data, err = json.MarshalIndent(values, "", "  ")
	default:
		return fmt.Errorf("unsupported file format: %s", filepath.Ext(filename))
	}

	if err != nil {
		return fmt.Errorf("failed to marshal values: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
