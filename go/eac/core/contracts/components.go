package contracts

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ComponentType defines how to process files of a certain type for linting and other operations.
// This is loaded from component-types.yml and used by the lint orchestrator.
type ComponentType struct {
	// Extensions are the file extensions belonging to this component type (e.g., [".go"], [".md", ".markdown"])
	Extensions []string `yaml:"extensions" json:"extensions"`

	// Linter is the linter tool to use (e.g., "golangci-lint", "markdownlint-cli2")
	Linter string `yaml:"linter,omitempty" json:"linter,omitempty"`

	// LintInput specifies how files are passed to the linter:
	// - "packages": Lint by directory/package (e.g., Go modules use ./...)
	// - "files": Lint individual files (e.g., markdown files)
	LintInput string `yaml:"lint_input,omitempty" json:"lint_input,omitempty"`
}

// GetLintInput returns the lint input mode, defaulting to "files" if not specified.
func (c *ComponentType) GetLintInput() string {
	if c.LintInput == "" {
		return "files"
	}
	return c.LintInput
}

// HasLinter returns true if this component type has a linter configured.
func (c *ComponentType) HasLinter() bool {
	return c.Linter != ""
}

// ModuleComponents is a map of component type name to its configuration.
type ModuleComponents map[string]*ComponentEntry

// ComponentEntry represents a component's configuration within a module.
type ComponentEntry struct {
	Type        string                 `yaml:"type,omitempty" json:"type,omitempty"` // Component type (if different from map key name)
	Root        string                 `yaml:"root,omitempty" json:"root,omitempty"`
	Patterns    *ComponentPatterns     `yaml:"patterns,omitempty" json:"patterns,omitempty"`
	Build       *ComponentBuild        `yaml:"build,omitempty" json:"build,omitempty"`
	DockerBuild map[string]interface{} `yaml:"docker_build,omitempty" json:"docker_build,omitempty"`
	Resolved    bool                   `yaml:"-" json:"-"`
}

// ComponentBuild contains build configuration for a component.
type ComponentBuild struct {
	Handler   string              `yaml:"handler,omitempty" json:"handler,omitempty"`
	Artifacts []ComponentArtifact `yaml:"artifacts,omitempty" json:"artifacts,omitempty"`
}

// ComponentArtifact defines an artifact produced by a component build.
type ComponentArtifact struct {
	ID          string `yaml:"id" json:"id"`
	Type        string `yaml:"type" json:"type"`
	Pattern     string `yaml:"pattern" json:"pattern"`
	Compression string `yaml:"compression,omitempty" json:"compression,omitempty"`
	DeriveFrom  string `yaml:"derive_from,omitempty" json:"derive_from,omitempty"`
}

// ComponentPatterns contains file pattern overrides for a component.
type ComponentPatterns struct {
	Source []string `yaml:"source,omitempty" json:"source,omitempty"`
	Tests  []string `yaml:"tests,omitempty" json:"tests,omitempty"`
	Config []string `yaml:"config,omitempty" json:"config,omitempty"`
}

// UnmarshalYAML implements custom unmarshaling for ComponentEntry.
func (ce *ComponentEntry) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!null" || node.Value == "") {
		return nil
	}
	if node.Kind == yaml.ScalarNode {
		ce.Root = node.Value
		return nil
	}
	if node.Kind == yaml.MappingNode {
		type rawComponentEntry ComponentEntry
		return node.Decode((*rawComponentEntry)(ce))
	}
	return fmt.Errorf("invalid component entry: expected string or object")
}

// HasComponent returns true if the given component type is enabled.
func (mc ModuleComponents) HasComponent(compType string) bool {
	if mc == nil {
		return false
	}
	_, ok := mc[compType]
	return ok
}

// GetDefault returns the first component type found.
func (mc ModuleComponents) GetDefault() string {
	for name := range mc {
		return name
	}
	return ""
}

// GetEnabled returns all enabled component types.
func (mc ModuleComponents) GetEnabled() []string {
	if mc == nil {
		return nil
	}
	enabled := make([]string, 0, len(mc))
	for name := range mc {
		enabled = append(enabled, name)
	}
	return enabled
}

// GetComponentRoot returns the root path for a component by name.
func (mc ModuleComponents) GetComponentRoot(compName string) string {
	if mc == nil {
		return ""
	}
	if entry, ok := mc[compName]; ok && entry != nil {
		return entry.Root
	}
	return ""
}

// GetComponentType returns the type for a named component.
// If Type field is empty, the name IS the type (backward compatibility).
func (mc ModuleComponents) GetComponentType(name string) string {
	if entry, ok := mc[name]; ok && entry != nil && entry.Type != "" {
		return entry.Type
	}
	return name // name is the type
}

// HasComponentType returns true if any component has the given type.
func (mc ModuleComponents) HasComponentType(typeName string) bool {
	if mc == nil {
		return false
	}
	for name, entry := range mc {
		t := name
		if entry != nil && entry.Type != "" {
			t = entry.Type
		}
		if t == typeName {
			return true
		}
	}
	return false
}

// GetComponentsByType returns all components of a given type.
// Returns a map of component name to entry.
func (mc ModuleComponents) GetComponentsByType(typeName string) map[string]*ComponentEntry {
	result := make(map[string]*ComponentEntry)
	if mc == nil {
		return result
	}
	for name, entry := range mc {
		t := name
		if entry != nil && entry.Type != "" {
			t = entry.Type
		}
		if t == typeName {
			result[name] = entry
		}
	}
	return result
}

// GetComponentTypes returns unique types used across all components.
func (mc ModuleComponents) GetComponentTypes() []string {
	if mc == nil {
		return nil
	}
	types := make(map[string]bool)
	for name, entry := range mc {
		t := name
		if entry != nil && entry.Type != "" {
			t = entry.Type
		}
		types[t] = true
	}
	result := make([]string, 0, len(types))
	for t := range types {
		result = append(result, t)
	}
	return result
}

// GetBuildableRoot returns the root of the first buildable component (go or typescript).
// Returns empty string if no buildable component exists.
// Use this only for changelog derivation when no explicit changelog is set.
func (mc ModuleComponents) GetBuildableRoot() string {
	if mc == nil {
		return ""
	}

	// Only check buildable component types for changelog derivation
	buildableTypes := []string{"go", "typescript"}
	for _, compType := range buildableTypes {
		// Use type-aware lookup
		comps := mc.GetComponentsByType(compType)
		for _, entry := range comps {
			if entry != nil && entry.Root != "" {
				return entry.Root
			}
		}
	}
	return ""
}

// GetAllRoots returns all component roots as a map of component type to root path.
func (mc ModuleComponents) GetAllRoots() map[string]string {
	if mc == nil {
		return nil
	}
	roots := make(map[string]string)
	for name, entry := range mc {
		if entry != nil && entry.Root != "" {
			roots[name] = entry.Root
		}
	}
	return roots
}

// ResolvedComponent is a component with its actual files computed at runtime.
type ResolvedComponent struct {
	Type  string   `json:"type"`
	Files []string `json:"files"`
	Count int      `json:"count"`
}
