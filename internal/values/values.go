package values

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(filename string) (*yaml.Node, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal values: %w", err)
	}

	return &node, nil
}

func Write(filename string, node *yaml.Node) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	err = encoder.Encode(node)
	if err != nil {
		return fmt.Errorf("failed to encode values: %w", err)
	}

	return nil
}

func GetValue(node *yaml.Node, keys ...string) (string, error) {
	if node.Kind != yaml.DocumentNode {
		return "", fmt.Errorf("expected document node")
	}

	current := node.Content[0]
	for _, key := range keys {
		found := false
		for i := 0; i < len(current.Content); i += 2 {
			if current.Content[i].Value == key {
				current = current.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("key not found: %s", key)
		}
	}

	return current.Value, nil
}

func SetValue(node *yaml.Node, newValue string, keys ...string) error {
	if node.Kind != yaml.DocumentNode {
		return fmt.Errorf("expected document node")
	}

	current := node.Content[0]
	for _, key := range keys[:len(keys)-1] {
		found := false
		for i := 0; i < len(current.Content); i += 2 {
			if current.Content[i].Value == key {
				current = current.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("key not found: %s", key)
		}
	}

	lastKey := keys[len(keys)-1]
	for i := 0; i < len(current.Content); i += 2 {
		if current.Content[i].Value == lastKey {
			current.Content[i+1].Value = newValue
			return nil
		}
	}

	return fmt.Errorf("key not found: %s", lastKey)
}

func DeleteValue(node *yaml.Node, keys ...string) error {
	if node.Kind != yaml.DocumentNode {
		return fmt.Errorf("expected document node")
	}

	current := node.Content[0]
	for _, key := range keys[:len(keys)-1] {
		found := false
		for i := 0; i < len(current.Content); i += 2 {
			if current.Content[i].Value == key {
				current = current.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("key not found: %s", key)
		}
	}

	lastKey := keys[len(keys)-1]
	for i := 0; i < len(current.Content); i += 2 {
		if current.Content[i].Value == lastKey {
			current.Content = append(current.Content[:i], current.Content[i+2:]...)
			return nil
		}
	}

	return fmt.Errorf("key not found: %s", lastKey)
}
