package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/resource"
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

	// ToolChain is an ordered list of tools to run in sequence within a single UOW.
	// Each tool depends on the previous tool completing. When set, this takes
	// precedence over the legacy Builder field.
	// Example: ["preprocess", "mkdocs-pdf"] runs preprocess first, then mkdocs-pdf.
	ToolChain []string `yaml:"tool_chain,omitempty" json:"tool_chain,omitempty"`

	// Builder is the build handler to use (e.g., "go", "mkdocs", "buildx")
	// Deprecated: Use Tools for new component types. Kept for backwards compatibility.
	Builder string `yaml:"builder,omitempty" json:"builder,omitempty"`

	// Scanners are the default security scanners for this component type (e.g., ["sbom", "vuln", "secrets", "sast"])
	// Empty or omitted means the component type is not scannable.
	Scanners []string `yaml:"scanners,omitempty" json:"scanners,omitempty"`

	// Testable indicates this component type supports testing (e.g., go test, npm test)
	Testable bool `yaml:"testable,omitempty" json:"testable,omitempty"`

	// Lintable indicates this component type supports linting
	// When true, lint providers can be configured for this type
	Lintable bool `yaml:"lintable,omitempty" json:"lintable,omitempty"`

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

	// Resources defines resource requirements for this component type.
	// Used for scheduling weight and container resource limits.
	Resources *ComponentTypeResources `yaml:"resources,omitempty" json:"resources,omitempty"`

	// Defaults contains default values for component instances of this type.
	// Enables convention-over-configuration patterns.
	Defaults *ComponentTypeDefaults `yaml:"defaults,omitempty" json:"defaults,omitempty"`

	// Pool specifies which resource pool this component uses.
	// Valid values: "host" (default), "docker".
	// "docker" means the component runs in a container and uses BOTH pools.
	// If not specified, inferred from Requirements (docker requirement -> docker pool).
	Pool string `yaml:"pool,omitempty" json:"pool,omitempty"`

	// Amp is the weight amplifier for this component type.
	// Multiplies the tool's base resource weight (from tool-config.yml resources.cpus).
	// Default: 1.0 (no amplification). Values < 1.0 reduce weight, > 1.0 increase weight.
	Amp float64 `yaml:"amp,omitempty" json:"amp,omitempty"`
}

// ComponentTypeResources defines resource requirements for a component type.
// CPUs is used as the scheduling weight. Memory is used for container limits.
type ComponentTypeResources struct {
	// CPUs is the number of CPUs required. Used as scheduling weight (host pool).
	// Default: 1
	CPUs int `yaml:"cpus,omitempty" json:"cpus,omitempty"`

	// Memory is the memory requirement (e.g., "8g", "512m").
	// Used for container memory limits.
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`

	// DockerWeight is the weight for docker pool allocation.
	// If not set, defaults to CPUs value for container components.
	// Only used when component pool is "docker".
	DockerWeight int `yaml:"docker_weight,omitempty" json:"docker_weight,omitempty"`
}

// ComponentTypeDefaults contains default values for component instances.
// These enable convention-over-configuration patterns for simpler declarations.
type ComponentTypeDefaults struct {
	// BookFromName if true, derives book name from component name.
	// For "tutorials-pdf", book="tutorials".
	BookFromName bool `yaml:"book_from_name,omitempty" json:"book_from_name,omitempty"`

	// Theme is the default theme for PDF rendering ("dark" or "light").
	Theme string `yaml:"theme,omitempty" json:"theme,omitempty"`

	// ArtifactPattern is the default artifact naming pattern for built outputs.
	// Supports placeholders: {moniker} for module name, {ext} for platform extension.
	// Example: "{moniker}{ext}" produces "r2r.exe" on Windows, "r2r" on Unix.
	ArtifactPattern string `yaml:"artifact_pattern,omitempty" json:"artifact_pattern,omitempty"`
}

// GetArtifactName returns the artifact name by substituting placeholders in the pattern.
// Returns empty string if defaults is nil or pattern is empty.
func (d *ComponentTypeDefaults) GetArtifactName(moniker, ext string) string {
	if d == nil || d.ArtifactPattern == "" {
		return ""
	}
	name := strings.ReplaceAll(d.ArtifactPattern, "{moniker}", moniker)
	name = strings.ReplaceAll(name, "{ext}", ext)
	return name
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
	return c.Builder != "" || len(c.ToolChain) > 0
}

// GetTools returns the list of tools to run for this component type.
// If ToolChain is set, returns it directly. Otherwise, falls back to a single-element
// slice containing Builder for backwards compatibility. Returns nil if neither is set.
func (c *ComponentType) GetTools() []string {
	if c == nil {
		return nil
	}
	if len(c.ToolChain) > 0 {
		return c.ToolChain
	}
	if c.Builder != "" {
		return []string{c.Builder}
	}
	return nil
}

// HasToolChain returns true if this component type has multiple tools configured.
// A tool chain means tools are run in sequence with dependencies between them.
func (c *ComponentType) HasToolChain() bool {
	if c == nil {
		return false
	}
	return len(c.ToolChain) > 1
}

// IsBuildable returns true if this component type has a builder configured.
func (c *ComponentType) IsBuildable() bool {
	return c.HasBuilder()
}

// IsTestable returns true if this component type supports testing.
func (c *ComponentType) IsTestable() bool {
	return c.Testable
}

// IsLintable returns true if this component type supports linting.
func (c *ComponentType) IsLintable() bool {
	return c.Lintable
}

// IsScannable returns true if this component type has scanners configured.
// Component types without scanners (or with empty scanners) are not scannable.
func (c *ComponentType) IsScannable() bool {
	return len(c.Scanners) > 0
}

// GetScanners returns the list of default scanners for this component type.
func (c *ComponentType) GetScanners() []string {
	return c.Scanners
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

// GetWeight returns the scheduling weight for this component type.
// Returns the CPUs value from resources, or 1 if not specified.
func (c *ComponentType) GetWeight() int {
	if c.Resources != nil && c.Resources.CPUs > 0 {
		return c.Resources.CPUs
	}
	return 1
}

// GetAmp returns the weight amplifier for this component type.
// Returns 1.0 if not specified or if the value is 0.
func (c *ComponentType) GetAmp() float64 {
	if c == nil || c.Amp == 0 {
		return 1.0
	}
	return c.Amp
}

// GetMemory returns the memory requirement for this component type.
// Returns empty string if not specified.
func (c *ComponentType) GetMemory() string {
	if c.Resources != nil {
		return c.Resources.Memory
	}
	return ""
}

// GetPool returns the resource pool type for this component.
// Returns "host" if not specified and no docker requirement.
// Returns "docker" if Pool field is "docker" OR if Requirements includes "docker".
func (c *ComponentType) GetPool() string {
	// Explicit pool setting takes precedence
	if c.Pool == "host" || c.Pool == "docker" {
		return c.Pool
	}
	// Infer from requirements
	if c.RequiresDocker() {
		return "docker"
	}
	return "host"
}

// RequiresDocker returns true if this component requires docker.
func (c *ComponentType) RequiresDocker() bool {
	for _, req := range c.Requirements {
		if req == "docker" {
			return true
		}
	}
	return false
}

// GetDockerWeight returns the docker pool weight.
// Returns 0 for host-only components.
func (c *ComponentType) GetDockerWeight() int {
	if c.GetPool() != "docker" {
		return 0
	}
	if c.Resources != nil && c.Resources.DockerWeight > 0 {
		return c.Resources.DockerWeight
	}
	// Default to CPU/host weight
	return c.GetWeight()
}

// GetPoolAllocation returns the complete pool allocation for this component.
func (c *ComponentType) GetPoolAllocation() resource.PoolAllocation {
	hostWeight := c.GetWeight()
	dockerWeight := c.GetDockerWeight()

	return resource.PoolAllocation{
		HostWeight:   hostWeight,
		DockerWeight: dockerWeight,
	}
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
