package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// OrderedMap represents a map that preserves insertion order when marshaled to JSON
type OrderedMap struct {
	keys   []string
	values map[string]interface{}
}

func newOrderedMap() *OrderedMap {
	return &OrderedMap{
		keys:   make([]string, 0),
		values: make(map[string]interface{}),
	}
}

func (om *OrderedMap) Set(key string, value interface{}) {
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
	}
	om.values[key] = value
}

// MarshalJSON implements json.Marshaler to preserve key order
func (om *OrderedMap) MarshalJSON() ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("{")
	for i, key := range om.keys {
		if i > 0 {
			buf.WriteString(",")
		}
		// Marshal key
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteString(":")
		// Marshal value
		valueBytes, err := json.Marshal(om.values[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteString("}")
	return []byte(buf.String()), nil
}

// RenderAsJSON converts a Go struct to JSON by first marshaling to YAML then to JSON
// This ensures YAML is the single source of truth for serialization
// Order is preserved by using yaml.Node to maintain field ordering
//
// Example:
//
//	type Person struct {
//	    Name  string `yaml:"name"`
//	    Age   int    `yaml:"age"`
//	}
//	person := Person{Name: "Alice", Age: 30}
//	jsonStr, err := RenderAsJSON(person)
func RenderAsJSON(v interface{}) (string, error) {
	// First marshal to YAML
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Unmarshal into yaml.Node to preserve order
	var node yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &node); err != nil {
		return "", fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Convert yaml.Node to ordered map/slice structure
	intermediate := yamlNodeToOrderedInterface(&node)

	// Finally marshal to JSON with indentation
	jsonBytes, err := marshalIndentOrdered(intermediate, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// marshalIndentOrdered marshals with proper indentation, handling OrderedMap
func marshalIndentOrdered(v interface{}, prefix, indent string) ([]byte, error) {
	// For OrderedMap, we need custom indentation
	if om, ok := v.(*OrderedMap); ok {
		return marshalOrderedMapIndent(om, prefix, indent, 0)
	}
	// For everything else, use standard json.MarshalIndent
	return json.MarshalIndent(v, prefix, indent)
}

// marshalOrderedMapIndent recursively marshals OrderedMap with indentation
func marshalOrderedMapIndent(om *OrderedMap, prefix, indent string, depth int) ([]byte, error) {
	if len(om.keys) == 0 {
		return []byte("{}"), nil
	}

	var buf strings.Builder
	currentIndent := prefix + strings.Repeat(indent, depth)
	nextIndent := prefix + strings.Repeat(indent, depth+1)

	buf.WriteString("{\n")
	for i, key := range om.keys {
		if i > 0 {
			buf.WriteString(",\n")
		}
		buf.WriteString(nextIndent)
		// Marshal key
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteString(": ")

		// Marshal value with proper indentation
		value := om.values[key]
		var valueBytes []byte
		if childOM, ok := value.(*OrderedMap); ok {
			valueBytes, err = marshalOrderedMapIndent(childOM, prefix, indent, depth+1)
		} else if childSlice, ok := value.([]interface{}); ok {
			valueBytes, err = marshalSliceIndent(childSlice, prefix, indent, depth+1)
		} else {
			valueBytes, err = json.Marshal(value)
		}
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteString("\n")
	buf.WriteString(currentIndent)
	buf.WriteString("}")
	return []byte(buf.String()), nil
}

// marshalSliceIndent marshals a slice with proper indentation
func marshalSliceIndent(slice []interface{}, prefix, indent string, depth int) ([]byte, error) {
	if len(slice) == 0 {
		return []byte("[]"), nil
	}

	var buf strings.Builder
	currentIndent := prefix + strings.Repeat(indent, depth)
	nextIndent := prefix + strings.Repeat(indent, depth+1)

	buf.WriteString("[\n")
	for i, item := range slice {
		if i > 0 {
			buf.WriteString(",\n")
		}
		buf.WriteString(nextIndent)

		var itemBytes []byte
		var err error
		if om, ok := item.(*OrderedMap); ok {
			itemBytes, err = marshalOrderedMapIndent(om, prefix, indent, depth+1)
		} else if childSlice, ok := item.([]interface{}); ok {
			itemBytes, err = marshalSliceIndent(childSlice, prefix, indent, depth+1)
		} else {
			itemBytes, err = json.Marshal(item)
		}
		if err != nil {
			return nil, err
		}
		buf.Write(itemBytes)
	}
	buf.WriteString("\n")
	buf.WriteString(currentIndent)
	buf.WriteString("]")
	return []byte(buf.String()), nil
}

// yamlNodeToOrderedInterface converts yaml.Node to interface{} while preserving order
func yamlNodeToOrderedInterface(node *yaml.Node) interface{} {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			return yamlNodeToOrderedInterface(node.Content[0])
		}
		return nil
	case yaml.MappingNode:
		// Use OrderedMap to maintain field order
		result := newOrderedMap()
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := yamlNodeToOrderedInterface(node.Content[i+1])
			result.Set(key, value)
		}
		return result
	case yaml.SequenceNode:
		result := make([]interface{}, len(node.Content))
		for i, item := range node.Content {
			result[i] = yamlNodeToOrderedInterface(item)
		}
		return result
	case yaml.ScalarNode:
		return decodeScalar(node)
	case yaml.AliasNode:
		return yamlNodeToOrderedInterface(node.Alias)
	default:
		return nil
	}
}

// decodeScalar converts YAML scalar values to appropriate Go types
func decodeScalar(node *yaml.Node) interface{} {
	var result interface{}
	// Let yaml unmarshal handle type detection
	if err := node.Decode(&result); err != nil {
		return node.Value
	}
	return result
}

// RenderAsJSONOrPanic is a convenience wrapper that panics on error
// Useful for cases where marshaling is guaranteed to succeed
func RenderAsJSONOrPanic(v interface{}) string {
	result, err := RenderAsJSON(v)
	if err != nil {
		panic(err)
	}
	return result
}
