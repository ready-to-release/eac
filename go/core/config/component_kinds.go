package config

import (
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/resource"
)

// Component type name constants for well-known component types.
// These correspond to keys in the blueprints.yml component-kinds configuration.
const (
	ComponentTypeSiteRender = "site-render"
	ComponentTypeDocsSite   = "docs-site"
	ComponentTypePdfRender  = "pdf-render"
	ComponentTypeDocsPdf    = "docs-pdf"
)

// ComponentKindsConfig holds component kind definitions loaded from blueprints.yml.
// After loading and evaluation, each kind becomes a ComponentType used throughout
// the domain layer.
type ComponentKindsConfig struct {
	// Kinds maps component kind name to its evaluated ComponentType.
	Kinds map[string]*ComponentType

	// extensionToKind maps file extension to component kind name for O(1) lookup.
	// Built once via buildExtensionIndex after all kinds are loaded.
	extensionToKind map[string]string `yaml:"-"`

	// extensionToType maps file extension to *ComponentType for O(1) lookup.
	// Built once via buildExtensionIndex after all kinds are loaded.
	extensionToType map[string]*ComponentType `yaml:"-"`
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

	// Deployers are the deploy tools for this component type (tool IDs).
	// Empty or omitted = not deployable.
	Deployers []string `yaml:"deployers,omitempty" json:"deployers,omitempty"`

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

	// ComponentGroup is the default component_group for components of this type.
	// Applied during component defaults resolution if the component doesn't already have one.
	ComponentGroup string `yaml:"component_group,omitempty" json:"component_group,omitempty"`

	// DependsOn is the default component depends_on for components of this type.
	// Applied during component defaults resolution if the component doesn't already have depends_on.
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`

	// SourcePrefixes are directory prefixes stripped during name inference.
	// When a component is declared in list format without an explicit name,
	// the name is derived from the root path by stripping these prefixes
	// and the module moniker. For example, go kind with source_prefixes [go, src]
	// maps "go/adapters/godog" → "godog" (for module moniker "adapters").
	SourcePrefixes []string `yaml:"source_prefixes,omitempty" json:"source_prefixes,omitempty"`

	// Pool specifies which resource pool this component uses.
	// Valid values: "host" (default), "docker".
	// "docker" means the component runs in a container and uses BOTH pools.
	// Pool is determined at tool bind time based on tool bindings.
	Pool string `yaml:"pool,omitempty" json:"pool,omitempty"`

	// Amp is the weight amplifier for this component type.
	// Multiplies the tool's base resource weight (from tool-config.yml resources.cpus).
	// Default: 1.0 (no amplification). Values < 1.0 reduce weight, > 1.0 increase weight.
	Amp float64 `yaml:"amp,omitempty" json:"amp,omitempty"`

	// BDDRunner is the component kind to auto-create when this component declares
	// a specs: facet. For example, "go" kind has BDDRunner "godog", meaning a go
	// component with specs: auto-creates a godog component for BDD step implementation.
	// Empty string means no auto-BDD runner.
	BDDRunner string `yaml:"bdd_runner,omitempty" json:"bdd_runner,omitempty"`

	// RunnerSearchDirs lists subdirectories to search for the BDD test runner file
	// (e.g., godog_test.go) relative to the base test implementation path.
	// Configured in blueprints.yml per BDD runner component kind.
	// "." means check the base path directly; additional entries like "specs"
	// search subdirectories of the base path.
	RunnerSearchDirs []string `yaml:"runner_search_dirs,omitempty" json:"runner_search_dirs,omitempty"`

	// DockerBuildDefaults provides default docker_build configuration for components of this kind.
	// Merged as defaults into components during resolution; component values take precedence.
	DockerBuildDefaults *DockerBuildConfig `yaml:"docker_build_defaults,omitempty" json:"docker_build_defaults,omitempty"`
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
	Data   []string `yaml:"data,omitempty" json:"data,omitempty"` // Owned files not processed by tools
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

// DeriveName infers a component name from a root path and module moniker.
// It strips source prefixes and the moniker from the path, normalizes separators,
// strips leading dots, and truncates to 16 characters.
//
// Algorithm:
//  1. Strip the first matching source prefix (e.g., "go/" from "go/adapters/godog")
//  2. If remaining path starts with moniker+"/", strip that prefix
//  3. If remaining path equals moniker exactly, use moniker as name
//  4. Replace "/" and "_" with "-"
//  5. Strip leading dots
//  6. Truncate to 16 characters
//  7. If result is empty, fall back to DefaultDeriveName (last path segment)
func (c *ComponentType) DeriveName(rootPath, moniker string) string {
	if rootPath == "" {
		return ""
	}
	p := filepath.ToSlash(rootPath)

	// Strip source prefixes
	if c != nil {
		for _, prefix := range c.SourcePrefixes {
			trimmed := strings.TrimPrefix(p, prefix+"/")
			if trimmed != p {
				p = trimmed
				break
			}
		}
	}

	// Strip moniker prefix or match exact
	if moniker != "" {
		if strings.HasPrefix(p, moniker+"/") {
			p = strings.TrimPrefix(p, moniker+"/")
		} else if p == moniker {
			return moniker
		}
	}

	// Normalize separators
	p = strings.ReplaceAll(p, "/", "-")
	p = strings.ReplaceAll(p, "_", "-")

	// Strip leading dots
	p = strings.TrimLeft(p, ".")

	// Truncate
	if len(p) > 16 {
		p = p[:16]
	}

	// Fallback
	if p == "" {
		return DefaultDeriveName(rootPath)
	}
	return p
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

// IsDeployable returns true if this component type has deployers configured.
func (c *ComponentType) IsDeployable() bool {
	return len(c.GetDeployers()) > 0
}

// GetDeployers returns the deployer tool IDs for this component type.
func (c *ComponentType) GetDeployers() []string {
	if c == nil {
		return nil
	}
	return c.Deployers
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
	case core.ActionDeploy:
		return c.Deployers
	default:
		return nil
	}
}

// Get returns a component type definition by name.
func (c *ComponentKindsConfig) Get(name string) *ComponentType {
	if c == nil || c.Kinds == nil {
		return nil
	}
	return c.Kinds[name]
}

// GetBDDRunnerTypes returns all component kind names that have a bdd_runner set.
// These are the BDD runner component types (e.g., "godog", "cucumberjs", "behave").
func (c *ComponentKindsConfig) GetBDDRunnerTypes() []string {
	if c == nil || c.Kinds == nil {
		return nil
	}
	var types []string
	for _, kind := range c.Kinds {
		if kind.BDDRunner != "" {
			types = append(types, kind.BDDRunner)
		}
	}
	return types
}

// GetByExtension returns the component type for a given file extension.
// Returns nil if no component type matches the extension.
// Uses a pre-built index for O(1) lookup when available.
func (c *ComponentKindsConfig) GetByExtension(ext string) *ComponentType {
	if c == nil || c.Kinds == nil {
		return nil
	}
	if c.extensionToType != nil {
		return c.extensionToType[ext]
	}
	// Fallback to linear scan if index not yet built
	for _, ct := range c.Kinds {
		for _, e := range ct.Extensions {
			if e == ext {
				return ct
			}
		}
	}
	return nil
}

// GetComponentKindNameByExtension returns the name of the component kind for a given file extension.
// Returns empty string if no component kind matches the extension.
// Uses a pre-built index for O(1) lookup when available.
func (c *ComponentKindsConfig) GetComponentKindNameByExtension(ext string) string {
	if c == nil || c.Kinds == nil {
		return ""
	}
	if c.extensionToKind != nil {
		return c.extensionToKind[ext]
	}
	// Fallback to linear scan if index not yet built
	for name, ct := range c.Kinds {
		for _, e := range ct.Extensions {
			if e == ext {
				return name
			}
		}
	}
	return ""
}

// buildExtensionIndex builds extension-to-kind and extension-to-type maps for O(1) lookup.
// Must be called after all component kinds are fully loaded.
func (c *ComponentKindsConfig) buildExtensionIndex() {
	if c == nil || c.Kinds == nil {
		return
	}
	c.extensionToKind = make(map[string]string)
	c.extensionToType = make(map[string]*ComponentType)
	for name, ct := range c.Kinds {
		for _, ext := range ct.Extensions {
			c.extensionToKind[ext] = name
			c.extensionToType[ext] = ct
		}
	}
}

// EnsureComponentKinds ensures the component kinds configuration is initialized.
// Component kinds are sourced exclusively from blueprints.yml component-kinds
// (loaded during LoadBlueprints). This method ensures a non-nil map exists
// as a safety net; it does not load from any file.
func (c *EACConfig) EnsureComponentKinds(validate bool) error {
	if c.ComponentKinds == nil {
		c.ComponentKinds = &ComponentKindsConfig{
			Kinds: make(map[string]*ComponentType),
		}
	}
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
func (c *ComponentKindsConfig) ResolveComponentsForFiles(files, enabledComponents []string) map[string]*ResolvedComponent {
	result := make(map[string]*ResolvedComponent)

	// Build enabled set for quick lookup
	enabledSet := make(map[string]bool)
	for _, comp := range enabledComponents {
		enabledSet[comp] = true
	}

	// Group files by component type
	for _, file := range files {
		ext := getFileExtension(file)
		compName := c.GetComponentKindNameByExtension(ext)

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
func (c *ComponentKindsConfig) GetBuilder(componentName string) string {
	ct := c.Get(componentName)
	if ct == nil || len(ct.Builders) == 0 {
		return ""
	}
	return ct.Builders[0]
}

