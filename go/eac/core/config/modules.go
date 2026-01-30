package config

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"gopkg.in/yaml.v3"
)

// Module represents a single module definition.
type Module struct {
	Moniker       string                   `yaml:"moniker"`
	Name          string                   `yaml:"name"`
	Description   string                   `yaml:"description"`
	Template      string                   `yaml:"template,omitempty"`       // Reference to a module template name
	AutoDiscover  bool                     `yaml:"auto_discover,omitempty"`  // Enable automatic component discovery
	Parameters    map[string]string        `yaml:"parameters,omitempty"`     // Template parameter values
	DependsOn     []string                 `yaml:"depends_on"`
	DependsOnCI   []string                 `yaml:"depends_on_ci"`            // CI artifact dependencies (merged into DependsOn)
	CIDeps        []string                 `yaml:"-"`                        // Computed: CI artifact deps for dispatch layering
	EvidenceBooks []string                 `yaml:"evidence_books,omitempty"` // Evidence book names, built via 'update evidence' command
	ReleaseBundle *ReleaseBundle           `yaml:"release_bundle,omitempty"` // Release bundle configuration (for release modules)
	Metadata      map[string]string        `yaml:"metadata,omitempty"`       // Generic key-value store for module-specific data
	Versioning    *ModuleVersioning        `yaml:"versioning,omitempty"`
	Components    ModuleComponents         `yaml:"components"`        // Component types for this module (required)
	Linting       *contracts.ModuleLinting `yaml:"linting,omitempty"` // Linting configuration overrides
}

// HasComponent returns true if a component with the given name exists for this module.
func (m *Module) HasComponent(name string) bool {
	return m.Components.HasComponent(name)
}

// GetComponentAmp returns the resource amplifier for a component and operation.
// Returns 1.0 (no amplification) if the component doesn't exist or has no amp config.
func (m *Module) GetComponentAmp(componentName, operation string) float64 {
	if m == nil || m.Components == nil {
		return 1.0
	}
	entry, exists := m.Components[componentName]
	if !exists || entry == nil {
		return 1.0
	}
	return entry.GetAmpForOperation(operation)
}

// GetEnabledComponents returns all component names for this module.
func (m *Module) GetEnabledComponents() []string {
	return m.Components.GetEnabled()
}

// GetComponentTypesDisplay returns a comma-separated list of enabled component types.
// This is useful for logging and display purposes.
func (m *Module) GetComponentTypesDisplay() string {
	comps := m.GetEnabledComponents()
	if len(comps) == 0 {
		return "(no components)"
	}
	result := comps[0]
	for i := 1; i < len(comps); i++ {
		result += ", " + comps[i]
	}
	return result
}

// GetBooks returns the list of book component names (components with type 'book').
// Names are sorted for consistent ordering.
func (m *Module) GetBooks() []string {
	books := m.Components.GetComponentsByType("book")
	if len(books) == 0 {
		return nil
	}
	names := make([]string, 0, len(books))
	for name := range books {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetChangelog returns the changelog path, defaulting to release/<moniker>/CHANGELOG.md.
func (m *Module) GetChangelog() string {
	if m.Versioning != nil && m.Versioning.Changelog != "" {
		return m.Versioning.Changelog
	}
	return "release/" + m.Moniker + "/CHANGELOG.md"
}

// ReleaseBundle configures how the release module creates GitHub releases.
type ReleaseBundle struct {
	TitleFormat string                  `yaml:"title_format"` // Title template, e.g., "{r2r} ({r2r_version}) + {eac} ({eac_version})"
	Headline    map[string]string       `yaml:"headline"`     // Map of label -> moniker for title, e.g., {"r2r": "r2r-cli", "eac": "ext-eac"}
	Categories  []ReleaseBundleCategory `yaml:"categories"`   // Grouped modules for release notes
}

// ReleaseBundleCategory groups modules in release notes.
type ReleaseBundleCategory struct {
	Name        string   `yaml:"name"`        // Category name, e.g., "Core Tools"
	Description string   `yaml:"description"` // Category description
	Modules     []string `yaml:"modules"`     // Module monikers in this category
}

// ModuleVersioning holds module versioning configuration.
// Modules are either "supporting" (Implicit scheme) or "releasable" (SemVer/CalVer scheme).
// Supporting modules have no explicit version and are released only as part of releasable modules.
// Releasable modules have explicit versions and their own CI/release pipelines.
type ModuleVersioning struct {
	Scheme      string `yaml:"scheme"`                 // Implicit (supporting), SemVer, or CalVer (releasable)
	Current     string `yaml:"current,omitempty"`      // Current version (releasable modules only)
	Changelog   string `yaml:"changelog,omitempty"`    // Path to changelog (defaults to release/<moniker>/CHANGELOG.md)
	ReleaseType string `yaml:"release_type,omitempty"` // published | internal | bundle | none
}

// ModuleBuild contains per-module build configuration.
type ModuleBuild struct {
	Handler   string           `yaml:"handler,omitempty"`    // Explicit build handler override
	Artifacts []ModuleArtifact `yaml:"artifacts,omitempty"`  // Artifacts to produce
	Options   *BuildOptions    `yaml:"options,omitempty"`    // Build behavior options
	PostBuild *PostBuildConfig `yaml:"post_build,omitempty"` // Post-build actions
}

// PostBuildConfig contains post-build actions for a component.
// Configured at the component level within a module's build configuration.
type PostBuildConfig struct {
	// CopyTo specifies the target path to copy build output.
	// Path is relative to workspace root.
	// The target directory is cleaned before copying to avoid stale files.
	CopyTo string `yaml:"copy_to,omitempty" json:"copy_to,omitempty"`

	// CopyFiles specifies additional files to copy after build.
	// Each entry maps a source path (relative to component root) to a
	// target path (relative to workspace root).
	CopyFiles []CopyFileEntry `yaml:"copy_files,omitempty" json:"copy_files,omitempty"`
}

// CopyFileEntry specifies a single file copy operation.
type CopyFileEntry struct {
	// From is the source path relative to component root.
	From string `yaml:"from" json:"from"`
	// To is the target path relative to workspace root.
	To string `yaml:"to" json:"to"`
}

// ModuleArtifact defines an artifact to be produced by a module build.
type ModuleArtifact struct {
	ID          string `yaml:"id"`                    // Unique artifact identifier
	Type        string `yaml:"type"`                  // executable, file, directory, test
	Pattern     string `yaml:"pattern"`               // Output path pattern
	Compression string `yaml:"compression,omitempty"` // none, strip, upx
	DeriveFrom  string `yaml:"derive_from,omitempty"` // Source artifact to derive from
}

// BuildOptions contains optional build behavior flags.
type BuildOptions struct {
	// Reserved for future build options
}

// Clone creates a deep copy of ModuleBuild.
func (b *ModuleBuild) Clone() *ModuleBuild {
	if b == nil {
		return nil
	}
	clone := &ModuleBuild{
		Handler: b.Handler,
	}
	if b.Artifacts != nil {
		clone.Artifacts = make([]ModuleArtifact, len(b.Artifacts))
		copy(clone.Artifacts, b.Artifacts)
	}
	if b.Options != nil {
		clone.Options = &BuildOptions{}
	}
	if b.PostBuild != nil {
		clone.PostBuild = &PostBuildConfig{
			CopyTo: b.PostBuild.CopyTo,
		}
	}
	return clone
}

// ModuleComponents is a map of component type name to its configuration.
// The first component in the map becomes the default for build operations.
// Each entry can be:
//   - A string: root path for the component
//   - nil/empty: use default_root from component-types.yml
//   - ComponentEntry: full configuration with root and optional pattern overrides
type ModuleComponents map[string]*ComponentEntry

// ComponentEntry represents a component's configuration within a module.
// It can be parsed from either a simple string (root path) or full object.
// AmpConfig contains per-operation resource amplifiers for a component.
// Each value is a multiplier applied to the tool's base weight.
// Values < 1.0 reduce resources, > 1.0 increase resources.
type AmpConfig struct {
	Build float64 `yaml:"build,omitempty" json:"build,omitempty"`
	Lint  float64 `yaml:"lint,omitempty" json:"lint,omitempty"`
	Test  float64 `yaml:"test,omitempty" json:"test,omitempty"`
	Scan  float64 `yaml:"scan,omitempty" json:"scan,omitempty"`
}

// GetAmp returns the amplifier for the given operation type.
// Returns 1.0 (no amplification) if not specified.
func (a *AmpConfig) GetAmp(op string) float64 {
	if a == nil {
		return 1.0
	}
	switch op {
	case "build":
		if a.Build > 0 {
			return a.Build
		}
	case "lint":
		if a.Lint > 0 {
			return a.Lint
		}
	case "test":
		if a.Test > 0 {
			return a.Test
		}
	case "scan":
		if a.Scan > 0 {
			return a.Scan
		}
	}
	return 1.0
}

type ComponentEntry struct {
	// Type is the component type (e.g., "go", "book"). If omitted, the map key (name) is used as the type.
	// This allows multiple components of the same type with different names.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Root is the root path for this component (relative to repo root)
	Root string `yaml:"root,omitempty" json:"root,omitempty"`

	// Patterns contains optional pattern overrides for this component
	Patterns *ComponentPatterns `yaml:"patterns,omitempty" json:"patterns,omitempty"`

	// Build contains build configuration for this component (artifacts, handler override)
	Build *ModuleBuild `yaml:"build,omitempty" json:"build,omitempty"`

	// DockerBuild contains Docker build configuration (for dockerfile components)
	DockerBuild *DockerBuildConfig `yaml:"docker_build,omitempty" json:"docker_build,omitempty"`

	// Amp contains per-operation resource amplifiers
	Amp *AmpConfig `yaml:"amp,omitempty" json:"amp,omitempty"`

	// Resolved indicates if this entry has been resolved with defaults
	Resolved bool `yaml:"-" json:"-"`
}

// GetAmpForOperation returns the resource amplifier for an operation.
// Returns 1.0 (no amplification) if not specified.
func (ce *ComponentEntry) GetAmpForOperation(op string) float64 {
	if ce == nil {
		return 1.0
	}
	return ce.Amp.GetAmp(op)
}

// ComponentPatterns contains optional file pattern overrides for a component.
type ComponentPatterns struct {
	Source []string `yaml:"source,omitempty" json:"source,omitempty"`
	Tests  []string `yaml:"tests,omitempty" json:"tests,omitempty"`
	Config []string `yaml:"config,omitempty" json:"config,omitempty"`
}

// UnmarshalYAML implements custom unmarshaling for ComponentEntry.
// Handles both string (root path) and object formats.
func (ce *ComponentEntry) UnmarshalYAML(node *yaml.Node) error {
	// Handle null/empty - use default root
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!null" || node.Value == "") {
		return nil
	}

	// Handle string - just the root path
	if node.Kind == yaml.ScalarNode {
		ce.Root = node.Value
		return nil
	}

	// Handle object - full configuration
	if node.Kind == yaml.MappingNode {
		type rawComponentEntry ComponentEntry
		return node.Decode((*rawComponentEntry)(ce))
	}

	return fmt.Errorf("invalid component entry: expected string or object")
}

// Clone creates a copy of ModuleComponents.
func (mc ModuleComponents) Clone() ModuleComponents {
	if mc == nil {
		return nil
	}
	clone := make(ModuleComponents)
	for name, entry := range mc {
		if entry != nil {
			clonedEntry := &ComponentEntry{
				Type:     entry.Type,
				Root:     entry.Root,
				Resolved: entry.Resolved,
			}
			if entry.Patterns != nil {
				clonedEntry.Patterns = &ComponentPatterns{}
				if entry.Patterns.Source != nil {
					clonedEntry.Patterns.Source = make([]string, len(entry.Patterns.Source))
					copy(clonedEntry.Patterns.Source, entry.Patterns.Source)
				}
				if entry.Patterns.Tests != nil {
					clonedEntry.Patterns.Tests = make([]string, len(entry.Patterns.Tests))
					copy(clonedEntry.Patterns.Tests, entry.Patterns.Tests)
				}
				if entry.Patterns.Config != nil {
					clonedEntry.Patterns.Config = make([]string, len(entry.Patterns.Config))
					copy(clonedEntry.Patterns.Config, entry.Patterns.Config)
				}
			}
			if entry.Build != nil {
				clonedEntry.Build = entry.Build.Clone()
			}
			if entry.DockerBuild != nil {
				clonedEntry.DockerBuild = entry.DockerBuild.Clone()
			}
			clone[name] = clonedEntry
		} else {
			clone[name] = nil
		}
	}
	return clone
}

// HasComponent returns true if a component with the given name exists.
func (mc ModuleComponents) HasComponent(name string) bool {
	if mc == nil {
		return false
	}
	_, ok := mc[name]
	return ok
}

// GetDefault returns the default component name (first in map), or empty string if not set.
// Note: Go maps don't preserve order, so this returns the first encountered.
// For deterministic behavior, use GetDefaultComponent which checks for common defaults.
func (mc ModuleComponents) GetDefault() string {
	if mc == nil {
		return ""
	}
	// Return first key found (order not guaranteed in Go maps)
	for name := range mc {
		return name
	}
	return ""
}

// GetEnabled returns a list of all component names.
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

// GetComponentRoot returns the root path for a specific component by name.
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

// GetBuildHandler returns the first explicit build handler found in components.
func (m *Module) GetBuildHandler() string {
	for _, comp := range m.Components {
		if comp != nil && comp.Build != nil && comp.Build.Handler != "" {
			return comp.Build.Handler
		}
	}
	return ""
}

// GetDockerBuildConfig returns the docker_build configuration from the dockerfile component.
// Returns nil if no dockerfile component exists or no docker_build is configured.
func (m *Module) GetDockerBuildConfig() *DockerBuildConfig {
	dockerfileEntry := m.Components["dockerfile"]
	if dockerfileEntry == nil || dockerfileEntry.DockerBuild == nil {
		return nil
	}
	return dockerfileEntry.DockerBuild
}

// applyModuleDefaults applies default values to all modules (generic defaults only).
// Call ApplyComponentDefaults after loading ComponentTypes for component-specific defaults.
func (c *RepositoryConfig) applyModuleDefaults() {
	for i := range c.Modules {
		m := &c.Modules[i]

		if m.Description == "" {
			m.Description = m.Name
		}
		if m.DependsOn == nil {
			m.DependsOn = []string{}
		}

		// Merge depends_on_ci into depends_on and track separately as CIDeps
		if len(m.DependsOnCI) > 0 {
			m.CIDeps = make([]string, len(m.DependsOnCI))
			copy(m.CIDeps, m.DependsOnCI)

			// Merge into DependsOn (avoiding duplicates)
			existing := make(map[string]bool)
			for _, dep := range m.DependsOn {
				existing[dep] = true
			}
			for _, ciDep := range m.DependsOnCI {
				if !existing[ciDep] {
					m.DependsOn = append(m.DependsOn, ciDep)
					existing[ciDep] = true
				}
			}
		}

		// Modules without versioning config default to Implicit scheme
		// Implicit modules derive their version from parent modules and don't have standalone releases.
		if m.Versioning == nil {
			m.Versioning = &ModuleVersioning{
				Scheme:      "Implicit",
				ReleaseType: "internal",
			}
		} else if m.Versioning.Scheme == "Implicit" {
			// Implicit versioned modules can only have release_type "internal" or "none"
			// Default to "internal" if not specified or if an invalid value was given
			if m.Versioning.ReleaseType == "" {
				m.Versioning.ReleaseType = "internal"
			} else if m.Versioning.ReleaseType != "internal" && m.Versioning.ReleaseType != "none" {
				// Silently correct invalid release types - implicit modules can't be published
				m.Versioning.ReleaseType = "internal"
			}
		}

		// Note: Other defaults (Source, Config, Changelog, Specs, etc.) are now
		// applied by ApplyComponentDefaults using component-specific defaults with fallback.
	}
}

// ApplyComponentDefaults resolves component roots and patterns from component-types.yml.
// This should be called after ComponentTypes are loaded.
func (c *RepositoryConfig) ApplyComponentDefaults(compTypes *ComponentTypesConfig) {
	for i := range c.Modules {
		m := &c.Modules[i]
		m.resolveComponentRoots(compTypes)
		m.resolveDerivedPaths()
	}
}

// resolveComponentRoots resolves default roots and patterns for all components in the module.
func (m *Module) resolveComponentRoots(compTypes *ComponentTypesConfig) {
	if compTypes == nil {
		return
	}

	for compName, entry := range m.Components {
		// Get type (explicit or name as fallback)
		compType := compName
		if entry != nil && entry.Type != "" {
			compType = entry.Type
		}

		// Look up in component-types.yml by TYPE, not name
		ct := compTypes.Get(compType)
		if ct == nil {
			continue
		}

		// Ensure entry exists
		if entry == nil {
			entry = &ComponentEntry{}
			m.Components[compName] = entry
		}

		// Resolve root from component-type default if not set
		if entry.Root == "" && ct.DefaultRoot != "" {
			entry.Root = ct.GetRoot(m.Moniker, "")
		}

		// Merge patterns from component-type if not overridden
		if entry.Patterns == nil {
			entry.Patterns = &ComponentPatterns{}
		}
		if entry.Patterns.Source == nil && ct.Files != nil {
			entry.Patterns.Source = substituteMoniker(ct.Files.Source, m.Moniker)
		}
		if entry.Patterns.Tests == nil && ct.Files != nil {
			entry.Patterns.Tests = substituteMoniker(ct.Files.Tests, m.Moniker)
		}
		if entry.Patterns.Config == nil && ct.Files != nil {
			entry.Patterns.Config = substituteMoniker(ct.Files.Config, m.Moniker)
		}

		entry.Resolved = true
	}
}

// substituteMoniker replaces {moniker} in a list of patterns.
func substituteMoniker(patterns []string, moniker string) []string {
	if patterns == nil {
		return nil
	}
	result := make([]string, len(patterns))
	for i, p := range patterns {
		result[i] = replaceMoniker(p, moniker)
	}
	return result
}

// replaceMoniker replaces {moniker} placeholder in a string.
func replaceMoniker(s, moniker string) string {
	return filepath.ToSlash(filepath.Clean(
		strings.ReplaceAll(s, "{moniker}", moniker),
	))
}

// resolveDerivedPaths is a no-op since changelog derivation is now done in contracts.
func (m *Module) resolveDerivedPaths() {
	// Changelog is derived from components in contracts.ModuleContract.GetChangelogPath()
}
