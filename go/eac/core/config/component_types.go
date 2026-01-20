package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ComponentTypesConfig represents the component-types.yml configuration.
type ComponentTypesConfig struct {
	// ComponentTypes maps component type name to its configuration
	ComponentTypes map[string]*ComponentType `yaml:"component-types"`
}

// ComponentType defines how to process files of a certain type for building.
// Linting configuration is separate - see LintProvider in lint_providers.go.
type ComponentType struct {
	// Extensions are the file extensions belonging to this component type (e.g., [".go"], [".md", ".markdown"])
	// Empty for non-file-based components like "book"
	Extensions []string `yaml:"extensions" json:"extensions"`

	// Builder is the build handler to use (e.g., "go", "mkdocs", "buildx")
	Builder string `yaml:"builder,omitempty" json:"builder,omitempty"`

	// BuildWeight is the resource weight for parallel build scheduling.
	// Higher weight = more resource pressure. Default is 1.
	// Examples: go=1, npm=1, mkdocs=4 (uses Docker), buildx=4
	BuildWeight int `yaml:"build_weight,omitempty" json:"build_weight,omitempty"`

	// TestWeight is the resource weight for parallel test scheduling.
	// Higher weight = more resource pressure. Default is 1.
	// Examples: go=1, typescript=2 (npm install overhead)
	TestWeight int `yaml:"test_weight,omitempty" json:"test_weight,omitempty"`

	// BuildAfter specifies component types that must complete before this one
	// within the same module. Used for intra-module dependency ordering.
	// Example: ["go"] means this component waits for the "go" component to finish.
	BuildAfter []string `yaml:"build_after,omitempty" json:"build_after,omitempty"`

	// Requirements are system dependencies needed for building (e.g., ["go"], ["docker"])
	Requirements []string `yaml:"requirements,omitempty" json:"requirements,omitempty"`

	// DefaultRoot is the default root path pattern for this component type.
	// Supports {moniker} variable for module-specific paths (e.g., "specs/{moniker}").
	DefaultRoot string `yaml:"default_root,omitempty" json:"default_root,omitempty"`

	// Files contains default file patterns for this component type
	Files *ComponentTypeFiles `yaml:"files,omitempty" json:"files,omitempty"`
}

// ComponentTypeFiles contains default file patterns for a component type.
type ComponentTypeFiles struct {
	Source []string `yaml:"source,omitempty" json:"source,omitempty"`
	Tests  []string `yaml:"tests,omitempty" json:"tests,omitempty"`
	Config []string `yaml:"config,omitempty" json:"config,omitempty"`
}


// GetRoot returns the root path for this component type, using the provided explicit root
// if set, otherwise falling back to DefaultRoot with {moniker} substitution.
// Returns empty string if neither explicit nor default root is available.
func (c *ComponentType) GetRoot(moniker, explicitRoot string) string {
	if explicitRoot != "" {
		return explicitRoot
	}
	if c.DefaultRoot != "" {
		return strings.ReplaceAll(c.DefaultRoot, "{moniker}", moniker)
	}
	return ""
}


// HasBuilder returns true if this component type has a builder configured.
func (c *ComponentType) HasBuilder() bool {
	return c.Builder != ""
}

// GetBuildWeight returns the build weight for parallel scheduling.
// Returns 1 as default if not specified.
func (c *ComponentType) GetBuildWeight() int {
	if c.BuildWeight <= 0 {
		return 1
	}
	return c.BuildWeight
}

// GetTestWeight returns the test weight for parallel scheduling.
// Returns 1 as default if not specified.
func (c *ComponentType) GetTestWeight() int {
	if c.TestWeight <= 0 {
		return 1
	}
	return c.TestWeight
}

// GetBuildAfter returns the list of component types that must complete
// before this one within the same module.
func (c *ComponentType) GetBuildAfter() []string {
	return c.BuildAfter
}

// GetRequirements returns the system dependencies needed for this component type.
func (c *ComponentType) GetRequirements() []string {
	return c.Requirements
}

// Get returns a component type definition by name.
func (c *ComponentTypesConfig) Get(name string) *ComponentType {
	if c == nil || c.ComponentTypes == nil {
		return nil
	}
	return c.ComponentTypes[name]
}

// GetByExtension returns the component type for a given file extension.
// Returns nil if no component type matches the extension.
func (c *ComponentTypesConfig) GetByExtension(ext string) *ComponentType {
	if c == nil || c.ComponentTypes == nil {
		return nil
	}
	for _, ct := range c.ComponentTypes {
		for _, e := range ct.Extensions {
			if e == ext {
				return ct
			}
		}
	}
	return nil
}

// GetComponentTypeNameByExtension returns the name of the component type for a given file extension.
// Returns empty string if no component type matches the extension.
func (c *ComponentTypesConfig) GetComponentTypeNameByExtension(ext string) string {
	if c == nil || c.ComponentTypes == nil {
		return ""
	}
	for name, ct := range c.ComponentTypes {
		for _, e := range ct.Extensions {
			if e == ext {
				return name
			}
		}
	}
	return ""
}

// LoadComponentTypes loads the component-types.yml configuration.
// It merges the contract defaults with any project-level overrides.
func (c *EACConfig) LoadComponentTypes(validate bool) error {
	// Load defaults first
	ctConfig, err := LoadComponentTypesDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return err
	}

	// Ensure we have a valid config even if no defaults
	if ctConfig == nil {
		ctConfig = &ComponentTypesConfig{
			ComponentTypes: make(map[string]*ComponentType),
		}
	}

	// Load project-level overrides (optional)
	overridePath := filepath.Join(c.ConfigRoot, ComponentTypesFileName)
	if data, err := os.ReadFile(overridePath); err == nil {
		var override ComponentTypesConfig
		if err := yaml.Unmarshal(data, &override); err != nil {
			return fmt.Errorf("parsing component-types override: %w", err)
		}
		// Merge overrides into defaults
		for name, ct := range override.ComponentTypes {
			ctConfig.ComponentTypes[name] = ct
		}
	}

	c.ComponentTypes = ctConfig
	return nil
}

// ResolvedComponent is a component with its actual files computed at runtime.
// This represents the files within a module that belong to a specific component type.
type ResolvedComponent struct {
	// Type is the component type name (e.g., "go", "markdown")
	Type string `json:"type"`

	// Files is the list of file paths belonging to this component (relative to workspace root)
	Files []string `json:"files"`

	// Count is the number of files in this component
	Count int `json:"count"`
}

// ResolveComponentsForFiles groups a list of files by their component type.
// Only files matching enabled component types are included.
// Returns a map of component type name to resolved component.
func (c *ComponentTypesConfig) ResolveComponentsForFiles(files, enabledComponents []string) map[string]*ResolvedComponent {
	result := make(map[string]*ResolvedComponent)

	// Build enabled set for quick lookup
	enabledSet := make(map[string]bool)
	for _, comp := range enabledComponents {
		enabledSet[comp] = true
	}

	// Group files by component type
	for _, file := range files {
		ext := getFileExtension(file)
		compName := c.GetComponentTypeNameByExtension(ext)

		// Skip files not matching any component type or not enabled
		if compName == "" || !enabledSet[compName] {
			continue
		}

		if result[compName] == nil {
			result[compName] = &ResolvedComponent{
				Type:  compName,
				Files: []string{},
			}
		}
		result[compName].Files = append(result[compName].Files, file)
	}

	// Update counts
	for _, comp := range result {
		comp.Count = len(comp.Files)
	}

	return result
}

// getFileExtension returns the file extension including the dot (e.g., ".go").
// Returns empty string if no extension.
func getFileExtension(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ""
}

// GetBuilder returns the builder for a component type, or empty string if none.
func (c *ComponentTypesConfig) GetBuilder(componentName string) string {
	ct := c.Get(componentName)
	if ct == nil {
		return ""
	}
	return ct.Builder
}

// GetBuildRequirements returns all system dependencies needed for a list of components.
// Deduplicates requirements across components.
func (c *ComponentTypesConfig) GetBuildRequirements(components []string) []string {
	if c == nil {
		return nil
	}

	depsMap := make(map[string]bool)
	for _, compName := range components {
		ct := c.Get(compName)
		if ct == nil {
			continue
		}
		for _, req := range ct.Requirements {
			depsMap[req] = true
		}
	}

	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	return deps
}
