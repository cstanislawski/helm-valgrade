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

	if len(node.Content) == 0 {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.MappingNode})
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
			newMap := &yaml.Node{Kind: yaml.MappingNode}
			current.Content = append(current.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, newMap)
			current = newMap
		}
	}

	lastKey := keys[len(keys)-1]
	for i := 0; i < len(current.Content); i += 2 {
		if current.Content[i].Value == lastKey {
			current.Content[i+1].Value = newValue
			return nil
		}
	}

	current.Content = append(current.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: lastKey}, &yaml.Node{Kind: yaml.ScalarNode, Value: newValue})
	return nil
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

func SetNestedValue(m map[string]interface{}, value interface{}, keys ...string) error {
	if len(keys) == 0 {
		return fmt.Errorf("no keys provided")
	}

	if len(keys) == 1 {
		m[keys[0]] = value
		return nil
	}

	key := keys[0]
	restKeys := keys[1:]

	if _, exists := m[key]; !exists {
		m[key] = make(map[string]interface{})
	}

	subMap, ok := m[key].(map[string]interface{})
	if !ok {
		return fmt.Errorf("value at key %s is not a map", key)
	}

	return SetNestedValue(subMap, value, restKeys...)
}
