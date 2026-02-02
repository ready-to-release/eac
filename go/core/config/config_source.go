// Package config provides source tracking for configuration files.
// This allows tools to show users where configuration values come from.
package config

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"
)

// ConfigLayer represents the source layer of a configuration file.
// Layers are merged in priority order: contract < user < personal.
type ConfigLayer string

// Configuration layer constants.
const (
	// LayerContract represents contract defaults (contracts/core/0.1.0/defaults/).
	LayerContract ConfigLayer = "contract"

	// LayerUser represents user/team configuration (.eac/).
	LayerUser ConfigLayer = "user"

	// LayerPersonal represents personal overrides (.eac/*.personal.yml).
	LayerPersonal ConfigLayer = "personal"
)

// LoadedFile represents a single configuration file with its source information.
type LoadedFile struct {
	// Path is the absolute path to the configuration file.
	Path string

	// Layer indicates the configuration layer (contract, user, personal).
	Layer ConfigLayer

	// Exists indicates whether the file exists on disk.
	Exists bool

	// Values is the count of configuration items in this file.
	// Zero if the file does not exist.
	Values int
}

// LoadedConfig represents a configuration type with all its source files.
type LoadedConfig struct {
	// Name is the configuration name (e.g., "repository", "component-types").
	Name string

	// Files contains all source files for this configuration, ordered by layer priority.
	Files []LoadedFile
}

// configFileSpec defines a configuration file to track.
type configFileSpec struct {
	name     string // Config name (e.g., "repository")
	filename string // File name (e.g., "repository.yml")
}

// allConfigSpecs lists all configuration files to track.
var allConfigSpecs = []configFileSpec{
	{name: "repository", filename: RepositoryFileName},
	{name: "component-types", filename: ComponentTypesFileName},
	{name: "environments", filename: EnvironmentsFileName},
	{name: "testing-tags", filename: TestingTagsFileName},
	{name: "test-suites", filename: TestSuitesFileName},
	{name: "books", filename: BooksFileName},
	{name: "commands", filename: CommandsFileName},
	{name: "lint-providers", filename: LintProvidersFileName},
}

// GetLoadedFiles returns information about all configuration files that were
// loaded (or could be loaded) for this EACConfig. This includes:
// - Contract defaults from contracts/core/0.1.0/defaults/
// - User configuration from .eac/
// - Personal overrides from .eac/*.personal.yml (if applicable)
//
// Each LoadedConfig contains the config name and all its source files with:
// - Absolute path
// - Layer (contract, user, personal)
// - Whether the file exists
// - Count of configuration values
func (c *EACConfig) GetLoadedFiles() []LoadedConfig {
	var result []LoadedConfig

	for _, spec := range allConfigSpecs {
		lc := LoadedConfig{
			Name:  spec.name,
			Files: c.getFilesForConfig(spec.filename),
		}
		result = append(result, lc)
	}

	return result
}

// getFilesForConfig returns LoadedFile entries for a specific config file.
func (c *EACConfig) getFilesForConfig(filename string) []LoadedFile {
	var files []LoadedFile

	// Contract defaults path
	contractPath := filepath.Join(c.RepoRoot, "contracts", "core", paths.DefaultsVersion, "defaults", filename)
	files = append(files, makeLoadedFile(contractPath, LayerContract))

	// User config path
	userPath := filepath.Join(c.ConfigRoot, filename)
	files = append(files, makeLoadedFile(userPath, LayerUser))

	return files
}

// makeLoadedFile creates a LoadedFile entry by checking existence and counting values.
func makeLoadedFile(path string, layer ConfigLayer) LoadedFile {
	lf := LoadedFile{
		Path:   path,
		Layer:  layer,
		Exists: false,
		Values: 0,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return lf
	}

	lf.Exists = true
	lf.Values = countYAMLValues(data)
	return lf
}

// countYAMLValues counts the number of values in a YAML file.
// This provides a rough measure of configuration density.
func countYAMLValues(data []byte) int {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return 0
	}

	return countNodeValues(&root)
}

// countNodeValues recursively counts values in a YAML node tree.
func countNodeValues(node *yaml.Node) int {
	if node == nil {
		return 0
	}

	switch node.Kind {
	case yaml.DocumentNode:
		count := 0
		for _, child := range node.Content {
			count += countNodeValues(child)
		}
		return count

	case yaml.MappingNode:
		count := 0
		// Mapping nodes have pairs: key, value, key, value...
		for i := 0; i < len(node.Content); i += 2 {
			if i+1 < len(node.Content) {
				count += countNodeValues(node.Content[i+1])
			}
		}
		return count

	case yaml.SequenceNode:
		count := 0
		for _, child := range node.Content {
			count += countNodeValues(child)
		}
		return count

	case yaml.ScalarNode:
		return 1

	case yaml.AliasNode:
		return countNodeValues(node.Alias)

	default:
		return 0
	}
}
