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

func GetValue(node *yaml.Node, keys ...string) (*yaml.Node, error) {
	if node.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("expected document node")
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
			return nil, fmt.Errorf("key not found: %s", key)
		}
	}

	return current, nil
}

func SetValue(node *yaml.Node, newValue interface{}, keys ...string) error {
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
			current.Content[i+1] = interfaceToNode(newValue)
			return nil
		}
	}

	current.Content = append(current.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: lastKey}, interfaceToNode(newValue))
	return nil
}

func interfaceToNode(v interface{}) *yaml.Node {
	switch val := v.(type) {
	case map[string]interface{}:
		n := &yaml.Node{Kind: yaml.MappingNode}
		for k, v := range val {
			n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k}, interfaceToNode(v))
		}
		return n
	case []interface{}:
		n := &yaml.Node{Kind: yaml.SequenceNode}
		for _, v := range val {
			n.Content = append(n.Content, interfaceToNode(v))
		}
		return n
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", v)}
	}
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
