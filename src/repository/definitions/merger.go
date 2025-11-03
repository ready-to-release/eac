package definitions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// isTemplateSkeletonPath checks if a directory path is part of a template skeleton structure
// that should be excluded from definitions processing
func isTemplateSkeletonPath(path, rootDir string) bool {
	// Normalize paths first to handle cross-platform paths
	normalizedPath := strings.ReplaceAll(path, "\\", "/")
	normalizedRootDir := strings.ReplaceAll(rootDir, "\\", "/")

	relPath, err := filepath.Rel(normalizedRootDir, normalizedPath)
	if err != nil {
		return false
	}

	// Convert to forward slashes for consistent pattern matching
	relPath = filepath.ToSlash(relPath)

	// Check if this path matches the pattern: */templates/*/skeleton
	// or is a subdirectory of such a path
	pathParts := strings.Split(relPath, "/")

	for i := 0; i < len(pathParts)-1; i++ {
		if pathParts[i] == "templates" && i < len(pathParts)-2 && pathParts[i+2] == "skeleton" {
			return true
		}
	}

	return false
}

// DefinitionFile represents a single definitions.yml file
type DefinitionFile struct {
	Path     string
	Content  *yaml.Node
	YAMLPath string
}

// EnumerateDefinitionFiles walks a directory tree and finds all definitions.yml files
func EnumerateDefinitionFiles(rootDir string) ([]DefinitionFile, error) {
	var definitions []DefinitionFile

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirName := info.Name()
			if dirName == ".git" || dirName == "node_modules" || dirName == ".vscode" || dirName == ".idea" || dirName == "out" {
				return filepath.SkipDir
			}

			// Skip template skeleton directories to avoid parsing Handlebars syntax as YAML
			if isTemplateSkeletonPath(path, rootDir) {
				return filepath.SkipDir
			}

			return nil
		}

		if info.Name() == "definitions.yml" {
			relPath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", path, err)
			}

			var yamlContent yaml.Node
			if err := yaml.Unmarshal(content, &yamlContent); err != nil {
				return fmt.Errorf("failed to parse YAML in %s: %w", path, err)
			}

			yamlPath := generateYAMLPath(relPath)

			definitions = append(definitions, DefinitionFile{
				Path:     path,
				Content:  &yamlContent,
				YAMLPath: yamlPath,
			})
		}

		return nil
	})

	return definitions, err
}

// generateYAMLPath converts a file path to a YAML path (e.g., "foo/bar/definitions.yml" -> "foo.bar")
func generateYAMLPath(filePath string) string {
	// Normalize path separators first to handle cross-platform paths
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")

	dir := filepath.Dir(normalizedPath)
	if dir == "." {
		return ""
	}

	// Convert to forward slashes for consistent cross-platform handling
	dir = filepath.ToSlash(dir)

	parts := strings.Split(dir, "/")
	return strings.Join(parts, ".")
}

// MergeDefinitions merges multiple definition files into a single YAML structure
func MergeDefinitions(definitions []DefinitionFile) (*yaml.Node, error) {
	result := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{{Kind: yaml.MappingNode}},
	}
	rootMap := result.Content[0]

	for _, def := range definitions {
		if def.YAMLPath == "" {
			// Merge root-level content
			if def.Content.Kind == yaml.DocumentNode && len(def.Content.Content) > 0 {
				if def.Content.Content[0].Kind == yaml.MappingNode {
					rootMap.Content = append(rootMap.Content, def.Content.Content[0].Content...)
				}
			} else if def.Content.Kind == yaml.MappingNode {
				rootMap.Content = append(rootMap.Content, def.Content.Content...)
			}
		} else {
			setNestedValue(rootMap, def.YAMLPath, def.Content)
		}
	}

	return result, nil
}

// setNestedValue sets a nested value in a YAML mapping
func setNestedValue(rootMap *yaml.Node, path string, value *yaml.Node) {
	parts := strings.Split(path, ".")
	current := rootMap

	for i, part := range parts {
		if i == len(parts)-1 {
			// Set the final value
			addToMapping(current, part, value)
		} else {
			// Navigate or create intermediate maps
			current = getOrCreateMapping(current, part)
		}
	}
}

// addToMapping adds a key-value pair to a YAML mapping
func addToMapping(mapping *yaml.Node, key string, value *yaml.Node) {
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}

	var valueNode *yaml.Node
	if value.Kind == yaml.DocumentNode && len(value.Content) > 0 {
		valueNode = value.Content[0]
	} else {
		valueNode = value
	}

	mapping.Content = append(mapping.Content, keyNode, valueNode)
}

// getOrCreateMapping gets or creates a nested mapping in a YAML structure
func getOrCreateMapping(mapping *yaml.Node, key string) *yaml.Node {
	// Look for existing key
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}

	// Create new mapping
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	valueNode := &yaml.Node{Kind: yaml.MappingNode}
	mapping.Content = append(mapping.Content, keyNode, valueNode)
	return valueNode
}

// ProcessDirectory is a convenience function that enumerates and merges all definitions in a directory
func ProcessDirectory(rootDir string) (*yaml.Node, error) {
	definitions, err := EnumerateDefinitionFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate definition files: %w", err)
	}

	merged, err := MergeDefinitions(definitions)
	if err != nil {
		return nil, fmt.Errorf("failed to merge definitions: %w", err)
	}

	return merged, nil
}
