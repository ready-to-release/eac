//go:build L1
// +build L1

package definitions

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// L1 Tests - Logic tests with minimal dependencies, under 2s

func TestMergeDefinitions_L1(t *testing.T) {
	t.Run("merge root level content", func(t *testing.T) {
		// Create yaml.Node content for testing
		rootContent := &yaml.Node{Kind: yaml.MappingNode}
		rootContent.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "root"},
			{Kind: yaml.ScalarNode, Value: "value"},
			{Kind: yaml.ScalarNode, Value: "shared"},
			{Kind: yaml.ScalarNode, Value: "root"},
		}

		definitions := []DefinitionFile{
			{
				Path:     "/root/definitions.yml",
				Content:  rootContent,
				YAMLPath: "",
			},
		}

		result, err := MergeDefinitions(definitions)
		if err != nil {
			t.Fatalf("MergeDefinitions failed: %v", err)
		}

		// Convert result to map for easier testing
		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		if resultMap["root"] != "value" {
			t.Errorf("Expected root value, got %v", resultMap["root"])
		}

		if resultMap["shared"] != "root" {
			t.Errorf("Expected shared value, got %v", resultMap["shared"])
		}
	})

	t.Run("merge nested content", func(t *testing.T) {
		nestedContent := &yaml.Node{Kind: yaml.MappingNode}
		nestedContent.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "nested"},
			{Kind: yaml.ScalarNode, Value: "value"},
		}

		definitions := []DefinitionFile{
			{
				Path:     "/root/test/definitions.yml",
				Content:  nestedContent,
				YAMLPath: "test",
			},
		}

		result, err := MergeDefinitions(definitions)
		if err != nil {
			t.Fatalf("MergeDefinitions failed: %v", err)
		}

		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		testNode, ok := resultMap["test"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected test node to be map[string]interface{}, got %T", resultMap["test"])
		}

		if testNode["nested"] != "value" {
			t.Errorf("Expected nested value, got %v", testNode["nested"])
		}
	})

	t.Run("merge complex nested structure", func(t *testing.T) {
		rootContent := &yaml.Node{Kind: yaml.MappingNode}
		rootContent.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "root"},
			{Kind: yaml.ScalarNode, Value: "value"},
		}

		level1Content := &yaml.Node{Kind: yaml.MappingNode}
		level1Content.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "level1"},
			{Kind: yaml.ScalarNode, Value: "value"},
		}

		level2Content := &yaml.Node{Kind: yaml.MappingNode}
		level2Content.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "level2"},
			{Kind: yaml.ScalarNode, Value: "deepvalue"},
		}

		definitions := []DefinitionFile{
			{Path: "/root/definitions.yml", Content: rootContent, YAMLPath: ""},
			{Path: "/root/a/definitions.yml", Content: level1Content, YAMLPath: "a"},
			{Path: "/root/a/b/definitions.yml", Content: level2Content, YAMLPath: "a.b"},
		}

		result, err := MergeDefinitions(definitions)
		if err != nil {
			t.Fatalf("MergeDefinitions failed: %v", err)
		}

		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		// Verify root level
		if resultMap["root"] != "value" {
			t.Errorf("Expected root value, got %v", resultMap["root"])
		}

		// Verify level 1
		aNode, ok := resultMap["a"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected 'a' node to be map[string]interface{}, got %T", resultMap["a"])
		}

		if aNode["level1"] != "value" {
			t.Errorf("Expected level1 value, got %v", aNode["level1"])
		}

		// Verify level 2
		bNode, ok := aNode["b"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected 'b' node to be map[string]interface{}, got %T", aNode["b"])
		}

		if bNode["level2"] != "deepvalue" {
			t.Errorf("Expected level2 deepvalue, got %v", bNode["level2"])
		}
	})

	t.Run("merge document node content", func(t *testing.T) {
		// Test merging content wrapped in document nodes
		innerContent := &yaml.Node{Kind: yaml.MappingNode}
		innerContent.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "doc"},
			{Kind: yaml.ScalarNode, Value: "value"},
		}

		documentContent := &yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{innerContent},
		}

		definitions := []DefinitionFile{
			{
				Path:     "/root/definitions.yml",
				Content:  documentContent,
				YAMLPath: "",
			},
		}

		result, err := MergeDefinitions(definitions)
		if err != nil {
			t.Fatalf("MergeDefinitions failed: %v", err)
		}

		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		if resultMap["doc"] != "value" {
			t.Errorf("Expected doc value, got %v", resultMap["doc"])
		}
	})

	t.Run("empty definitions", func(t *testing.T) {
		result, err := MergeDefinitions([]DefinitionFile{})
		if err != nil {
			t.Fatalf("MergeDefinitions failed: %v", err)
		}

		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		if len(resultMap) != 0 {
			t.Errorf("Expected empty result map, got %v", resultMap)
		}
	})
}

func TestMergeDefinitions_EdgeCases_L1(t *testing.T) {
	t.Run("conflicting keys at different paths", func(t *testing.T) {
		content1 := &yaml.Node{Kind: yaml.MappingNode}
		content1.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "key"},
			{Kind: yaml.ScalarNode, Value: "first"},
		}

		content2 := &yaml.Node{Kind: yaml.MappingNode}
		content2.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "key"},
			{Kind: yaml.ScalarNode, Value: "second"},
		}

		definitions := []DefinitionFile{
			{Path: "/root/a/definitions.yml", Content: content1, YAMLPath: "app1"},
			{Path: "/root/b/definitions.yml", Content: content2, YAMLPath: "app2"},
		}

		result, err := MergeDefinitions(definitions)
		if err != nil {
			t.Fatalf("MergeDefinitions failed: %v", err)
		}

		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		// Both app1 and app2 should exist with their respective keys
		app1Node, ok1 := resultMap["app1"].(map[string]interface{})
		app2Node, ok2 := resultMap["app2"].(map[string]interface{})

		if !ok1 {
			t.Fatalf("Expected app1 node to be map[string]interface{}, got %T", resultMap["app1"])
		}
		if !ok2 {
			t.Fatalf("Expected app2 node to be map[string]interface{}, got %T", resultMap["app2"])
		}

		if app1Node["key"] != "first" {
			t.Errorf("Expected app1 key to be 'first', got %v", app1Node["key"])
		}
		if app2Node["key"] != "second" {
			t.Errorf("Expected app2 key to be 'second', got %v", app2Node["key"])
		}
	})

	t.Run("mixed nested and root content", func(t *testing.T) {
		rootContent := &yaml.Node{Kind: yaml.MappingNode}
		rootContent.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "root"},
			{Kind: yaml.ScalarNode, Value: "value"},
		}

		nestedContent := &yaml.Node{Kind: yaml.MappingNode}
		nestedContent.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "nested"},
			{Kind: yaml.ScalarNode, Value: "value"},
		}

		definitions := []DefinitionFile{
			{Path: "/root/definitions.yml", Content: rootContent, YAMLPath: ""},
			{Path: "/root/app/definitions.yml", Content: nestedContent, YAMLPath: "app"},
		}

		result, err := MergeDefinitions(definitions)
		if err != nil {
			t.Fatalf("MergeDefinitions failed: %v", err)
		}

		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		// Should have both root and nested content
		if resultMap["root"] != "value" {
			t.Errorf("Expected root value, got %v", resultMap["root"])
		}

		appNode, ok := resultMap["app"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected app node to be map[string]interface{}, got %T", resultMap["app"])
		}

		if appNode["nested"] != "value" {
			t.Errorf("Expected nested value, got %v", appNode["nested"])
		}
	})
}
