package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/resource"
	"gopkg.in/yaml.v3"
)

// Component type name constants for well-known component types.
// These correspond to keys in the component-types.yml configuration.
const (
	ComponentTypeSiteRender = "site-render"
	ComponentTypeDocsSite   = "docs-site"
	ComponentTypePdfRender  = "pdf-render"
	ComponentTypeDocsPdf    = "docs-pdf"
)

// ComponentTypesConfig represents the component-types.yml configuration.
type ComponentTypesConfig struct {
	// ComponentTypes maps component type name to its configuration
	ComponentTypes map[string]*ComponentType `yaml:"component-types"`
}

// ComponentType defines how to process files of a certain type.
type ComponentType struct {
	// Extensions are the file extensions belonging to this component type (e.g., [".go"], [".md", ".markdown"])
	// Empty for non-file-based components like "book"
	Extensions []string `yaml:"extensions" json:"extensions"`

	// Builders are the build tools for this component type (tool IDs).
	// For multi-step builds, tools execute in order (e.g., ["mkdocs-preprocess", "pdf-oci"]).
	// Empty or omitted = not buildable.
	Builders []string `yaml:"builders,omitempty" json:"builders,omitempty"`

	// Linters are the lint tools for this component type (tool IDs).
	// Empty or omitted = not lintable.
	Linters []string `yaml:"linters,omitempty" json:"linters,omitempty"`

	// Testers are the test tools for this component type (tool IDs).
	// Empty or omitted = not testable.
	Testers []string `yaml:"testers,omitempty" json:"testers,omitempty"`

	// Scanners are the security scanner categories for this component type
	// (e.g., ["sbom", "vuln", "secrets", "sast"]).
	// Empty or omitted = not scannable.
	Scanners []string `yaml:"scanners,omitempty" json:"scanners,omitempty"`

	// BuildAfter specifies component types that must complete before this one
	// within the same module. Used for intra-module dependency ordering.
	// Example: ["go"] means this component waits for the "go" component to finish.
	BuildAfter []string `yaml:"build_after,omitempty" json:"build_after,omitempty"`

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
	// Pool is determined at tool bind time based on tool bindings.
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
	// Example: "{moniker}{ext}" produces "clie.exe" on Windows, "clie" on Unix.
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

// GetBuilders returns the builder tool IDs for this component type.
// For multi-step builds, tools execute in order.
func (c *ComponentType) GetBuilders() []string {
	if c == nil {
		return nil
	}
	return c.Builders
}

// IsBuildable returns true if this component type has builders configured.
func (c *ComponentType) IsBuildable() bool {
	return len(c.GetBuilders()) > 0
}

// HasToolChain returns true if this component type has multiple builders (multi-step build).
func (c *ComponentType) HasToolChain() bool {
	return len(c.GetBuilders()) > 1
}

// GetLinters returns the linter tool IDs for this component type.
func (c *ComponentType) GetLinters() []string {
	if c == nil {
		return nil
	}
	return c.Linters
}

// IsLintable returns true if this component type supports linting.
func (c *ComponentType) IsLintable() bool {
	return len(c.GetLinters()) > 0
}

// GetTesters returns the tester tool IDs for this component type.
func (c *ComponentType) GetTesters() []string {
	if c == nil {
		return nil
	}
	return c.Testers
}

// IsTestable returns true if this component type supports testing.
func (c *ComponentType) IsTestable() bool {
	return len(c.GetTesters()) > 0
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

// GetBuildAfter returns component types that must complete before this type.
// Creates intra-module dependencies: all components of this type wait for
// all components of the listed types within the same module.
func (c *ComponentType) GetBuildAfter() []string {
	if c == nil {
		return nil
	}
	return c.BuildAfter
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
// Returns "host" if not specified.
// Returns "docker" if Pool field is "docker".
func (c *ComponentType) GetPool() string {
	if c.Pool == "docker" {
		return "docker"
	}
	return "host"
}

// RequiresDocker returns true if this component uses the docker pool.
func (c *ComponentType) RequiresDocker() bool {
	return c.GetPool() == "docker"
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

// ToolIDsForAction returns the tool IDs for a given action type.
func (c *ComponentType) ToolIDsForAction(action core.ActionType) []string {
	if c == nil {
		return nil
	}
	switch action {
	case core.ActionBuild:
		return c.Builders
	case core.ActionLint:
		return c.Linters
	case core.ActionTest:
		return c.Testers
	case core.ActionScan:
		return c.Scanners
	default:
		return nil
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
// If component-kinds were already loaded from blueprints.yml, those are used as the base
// and component-types.yml entries override them (for backward compatibility).
func (c *EACConfig) LoadComponentTypes(validate bool) error {
	// Start with component-kinds already loaded from blueprints (if any)
	ctConfig := c.ComponentTypes
	if ctConfig == nil {
		ctConfig = &ComponentTypesConfig{
			ComponentTypes: make(map[string]*ComponentType),
		}
	}

	// Load contract defaults from component-types.yml (legacy, being phased out)
	defaults, err := LoadComponentTypesDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return err
	}
	if defaults != nil {
		for name, ct := range defaults.ComponentTypes {
			ctConfig.ComponentTypes[name] = ct
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
	// Type is the component type name (e.g., "go", "assets")
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

// GetBuilder returns the first builder for a component type, or empty string if none.
func (c *ComponentTypesConfig) GetBuilder(componentName string) string {
	ct := c.Get(componentName)
	if ct == nil || len(ct.Builders) == 0 {
		return ""
	}
	return ct.Builders[0]
}
