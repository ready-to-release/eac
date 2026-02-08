package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain"
	"gopkg.in/yaml.v3"
)

// Module represents a single module definition.
type Module struct {
	Moniker       string                `yaml:"moniker"`
	Name          string                `yaml:"name"`
	Description   string                `yaml:"description"`
	Template      string                `yaml:"template,omitempty"`      // Reference to a module template name
	ModuleGroup   string                `yaml:"module_group,omitempty"`  // Group name for depends_on expansion
	DependsOn     []string              `yaml:"depends_on"`
	DependsOnCI   []string              `yaml:"depends_on_ci"`            // CI artifact dependencies (merged into DependsOn)
	CIDeps        []string              `yaml:"-"`                        // Computed: CI artifact deps for dispatch layering
	EvidenceBooks []string              `yaml:"evidence_books,omitempty"` // Evidence book names, built via 'update evidence' command
	ReleaseBundle *ReleaseBundle        `yaml:"release_bundle,omitempty"` // Release bundle configuration (for release modules)
	Metadata      map[string]string     `yaml:"metadata,omitempty"`       // Generic key-value store for module-specific data
	Versioning    *ModuleVersioning     `yaml:"versioning,omitempty"`
	Components    ModuleComponents      `yaml:"components"`        // Component types for this module (required)
	Linting       *domain.ModuleLinting `yaml:"linting,omitempty"` // Linting configuration overrides

	// Component discovery
	DiscoverComponents *DiscoverComponentsConfig `yaml:"discover_components,omitempty"`

	ArtifactMatrixRef string `yaml:"artifact_matrix,omitempty"` // Reference to an artifact matrix name

	// ComponentOrder preserves YAML declaration order for display purposes.
	// Populated during UnmarshalYAML; components added later (e.g., by discovery) are appended.
	ComponentOrder []string `yaml:"-"`
}

// UnmarshalYAML implements custom unmarshaling for Module to capture component key order.
func (m *Module) UnmarshalYAML(node *yaml.Node) error {
	// Decode all fields using the default decoder
	type rawModule Module
	if err := node.Decode((*rawModule)(m)); err != nil {
		return err
	}

	// Extract component key order from the YAML node tree
	m.ComponentOrder = extractComponentOrder(node)
	return nil
}

// extractComponentOrder walks a Module YAML node to find the "components" mapping
// and returns its keys in declaration order.
func extractComponentOrder(moduleNode *yaml.Node) []string {
	if moduleNode.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(moduleNode.Content)-1; i += 2 {
		keyNode := moduleNode.Content[i]
		valNode := moduleNode.Content[i+1]
		if keyNode.Value == "components" && valNode.Kind == yaml.MappingNode {
			var order []string
			for j := 0; j < len(valNode.Content)-1; j += 2 {
				order = append(order, valNode.Content[j].Value)
			}
			return order
		}
	}
	return nil
}

// DiscoverComponentsConfig configures automatic component discovery.
type DiscoverComponentsConfig struct {
	Type       string   `yaml:"type"`                  // "containers" triggers container component discovery
	NameSuffix string   `yaml:"name_suffix,omitempty"` // Filter: only containers whose names end with this suffix
	LocalOnly  []string `yaml:"local_only,omitempty"`  // Containers that should not be pushed to registry
}

// UnmarshalYAML allows DiscoverComponentsConfig to be parsed from either
// a plain string (e.g., "containers") or a full object.
func (d *DiscoverComponentsConfig) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		d.Type = node.Value
		return nil
	}
	if node.Kind == yaml.MappingNode {
		type raw DiscoverComponentsConfig
		return node.Decode((*raw)(d))
	}
	return fmt.Errorf("invalid discover_components: expected string or object")
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
	return strings.Join(comps, ", ")
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

// GetChangelog returns the changelog path. Only SemVer modules default to release/<moniker>/CHANGELOG.md.
// CalVer modules have no changelog (auto-managed releases).
func (m *Module) GetChangelog() string {
	if m.Versioning == nil {
		return ""
	}
	if m.Versioning.Changelog != "" {
		return m.Versioning.Changelog
	}
	if m.Versioning.Scheme == "SemVer" {
		return "release/" + m.Moniker + "/CHANGELOG.md"
	}
	return ""
}

// ShouldAggregateFromDependencies returns true if this module should aggregate
// specs/approvals from its dependencies. This is true for:
// - Bundle modules (release_type: bundle)
// - Container modules (template contains "container")
// Regular library modules with compile-time dependencies should NOT aggregate.
func (m *Module) ShouldAggregateFromDependencies() bool {
	// No dependencies means nothing to aggregate
	if len(m.DependsOn) == 0 {
		return false
	}

	// Explicit bundle release type
	if m.Versioning != nil && m.Versioning.ReleaseType == "bundle" {
		return true
	}

	// Container modules (template name contains "container")
	if strings.Contains(m.Template, "container") {
		return true
	}

	return false
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
	Changelog   string `yaml:"changelog,omitempty"`    // Path to changelog (SemVer only; defaults to release/<moniker>/CHANGELOG.md)
	ReleaseType string `yaml:"release_type,omitempty"` // published | internal | bundle | none
}

// ModuleBuild contains per-module build configuration.
type ModuleBuild struct {
	Handler    string           `yaml:"handler,omitempty"`     // Explicit build handler override
	BinaryName string           `yaml:"binary_name,omitempty"` // Output binary name (for artifact matrix expansion)
	Artifacts  []ModuleArtifact `yaml:"artifacts,omitempty"`   // Artifacts to produce
	Options    *BuildOptions    `yaml:"options,omitempty"`     // Build behavior options
	PostBuild  *PostBuildConfig `yaml:"post_build,omitempty"`  // Post-build actions
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
		Handler:    b.Handler,
		BinaryName: b.BinaryName,
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
		if len(b.PostBuild.CopyFiles) > 0 {
			clone.PostBuild.CopyFiles = make([]CopyFileEntry, len(b.PostBuild.CopyFiles))
			copy(clone.PostBuild.CopyFiles, b.PostBuild.CopyFiles)
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

	// Config contains component-specific configuration (e.g., book name, theme).
	// Values here override defaults from component-types.yml.
	Config map[string]string `yaml:"config,omitempty" json:"config,omitempty"`

	// DependsOn lists component names this component depends on within the same module.
	// Used for intra-module build ordering (e.g., pdf-render depends on base-site).
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`

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

// GetConfig returns a config value by key, or empty string if not set.
func (ce *ComponentEntry) GetConfig(key string) string {
	if ce == nil || ce.Config == nil {
		return ""
	}
	return ce.Config[key]
}

// GetBook returns the book name from config, or empty string if not set.
func (ce *ComponentEntry) GetBook() string {
	return ce.GetConfig("book")
}

// GetTheme returns the theme from config, or empty string if not set.
func (ce *ComponentEntry) GetTheme() string {
	return ce.GetConfig("theme")
}

// GetDependsOn returns the list of component dependencies.
func (ce *ComponentEntry) GetDependsOn() []string {
	if ce == nil {
		return nil
	}
	return ce.DependsOn
}

// ComponentPatterns contains optional file pattern overrides for a component.
type ComponentPatterns struct {
	Source []string `yaml:"source,omitempty" json:"source,omitempty"`
	Tests  []string `yaml:"tests,omitempty" json:"tests,omitempty"`
	Config []string `yaml:"config,omitempty" json:"config,omitempty"`
	Data   []string `yaml:"data,omitempty" json:"data,omitempty"` // Owned files not processed by tools (e.g., testdata, fixtures)
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
				if entry.Patterns.Data != nil {
					clonedEntry.Patterns.Data = make([]string, len(entry.Patterns.Data))
					copy(clonedEntry.Patterns.Data, entry.Patterns.Data)
				}
			}
			if entry.Build != nil {
				clonedEntry.Build = entry.Build.Clone()
			}
			if entry.DockerBuild != nil {
				clonedEntry.DockerBuild = entry.DockerBuild.Clone()
			}
			if entry.Config != nil {
				clonedEntry.Config = make(map[string]string, len(entry.Config))
				for k, v := range entry.Config {
					clonedEntry.Config[k] = v
				}
			}
			if entry.DependsOn != nil {
				clonedEntry.DependsOn = make([]string, len(entry.DependsOn))
				copy(clonedEntry.DependsOn, entry.DependsOn)
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

// GetComponentConfig returns the config value for a component by name and key.
// Returns empty string if component or key not found.
func (mc ModuleComponents) GetComponentConfig(compName, key string) string {
	if mc == nil {
		return ""
	}
	entry, ok := mc[compName]
	if !ok || entry == nil {
		return ""
	}
	return entry.GetConfig(key)
}

// GetComponentDependsOn returns intra-module dependencies for a component.
func (mc ModuleComponents) GetComponentDependsOn(compName string) []string {
	if mc == nil {
		return nil
	}
	entry, ok := mc[compName]
	if !ok || entry == nil {
		return nil
	}
	return entry.GetDependsOn()
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

// GetDockerBuildConfigForComponent returns docker_build from a named component.
// Falls back to the "dockerfile" key for backward compatibility.
func (m *Module) GetDockerBuildConfigForComponent(componentName string) *DockerBuildConfig {
	if componentName != "" {
		if entry, ok := m.Components[componentName]; ok && entry != nil && entry.DockerBuild != nil {
			return entry.DockerBuild
		}
	}
	return m.GetDockerBuildConfig()
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
func (c *RepositoryConfig) ApplyComponentDefaults(compTypes *ComponentTypesConfig, repoRoot string) {
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

		// Apply component-type defaults (convention-over-configuration)
		m.applyComponentDefaults(compName, entry, ct)

		entry.Resolved = true
	}
}

// applyComponentDefaults applies convention-over-configuration defaults from component-type.
func (m *Module) applyComponentDefaults(compName string, entry *ComponentEntry, ct *ComponentType) {
	if ct.Defaults == nil {
		return
	}

	// Ensure Config map exists
	if entry.Config == nil {
		entry.Config = make(map[string]string)
	}

	// BookFromName: derive book name from component name if not explicitly set
	// Convention: "tutorials-base" → book="tutorials", "base-site" → requires explicit config
	if ct.Defaults.BookFromName {
		if _, hasBook := entry.Config["book"]; !hasBook {
			bookName := deriveBookName(compName)
			if bookName != "" {
				entry.Config["book"] = bookName
			}
		}
	}

	// Theme: set default theme for PDF rendering if not explicitly set
	if ct.Defaults.Theme != "" {
		if _, hasTheme := entry.Config["theme"]; !hasTheme {
			entry.Config["theme"] = ct.Defaults.Theme
		}
	}
}

// deriveBookName extracts book name from component name using conventions.
// "tutorials-base" → "tutorials" (strip -base suffix)
// "base-site" → "" (name IS "base-site", needs explicit config)
// "tutorials" → "tutorials" (name is the book)
func deriveBookName(compName string) string {
	// If ends with "-base", strip it to get book name
	if strings.HasSuffix(compName, "-base") {
		return strings.TrimSuffix(compName, "-base")
	}
	// If name is exactly "base-site", don't auto-derive (needs explicit config)
	if compName == "base-site" {
		return ""
	}
	// Otherwise, component name IS the book name
	return compName
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

// resolveDerivedPaths applies convention-based path derivation for module paths.
func (m *Module) resolveDerivedPaths() {
	// Auto-derive changelog path if not explicitly set
	if m.Versioning != nil && m.Versioning.Changelog == "" {
		m.Versioning.Changelog = deriveChangelogPath(m)
	}
}

// deriveChangelogPath auto-derives changelog path from release_type and moniker.
// Convention:
//   - CalVer: no changelog (auto-managed releases)
//   - SemVer published/bundle: release/{moniker}/CHANGELOG.md
//   - SemVer internal: {go_root}/CHANGELOG.md (if go component exists with root)
//   - none/empty/other: empty string (no changelog)
func deriveChangelogPath(m *Module) string {
	if m == nil || m.Versioning == nil {
		return ""
	}

	// Explicit changelog takes precedence
	if m.Versioning.Changelog != "" {
		return m.Versioning.Changelog
	}

	// CalVer modules have no changelogs (auto-managed releases)
	if m.Versioning.Scheme == "CalVer" {
		return ""
	}

	// Auto-derive based on release type
	switch m.Versioning.ReleaseType {
	case "published", "bundle":
		return fmt.Sprintf("release/%s/CHANGELOG.md", m.Moniker)
	case "internal":
		// Check for go component to derive from its root
		if goEntry, ok := m.Components["go"]; ok && goEntry != nil && goEntry.Root != "" {
			// Use forward slashes for consistent cross-platform paths
			return filepath.ToSlash(filepath.Join(goEntry.Root, "CHANGELOG.md"))
		}
		return ""
	default:
		return ""
	}
}

// DiscoverContainerModules scans the containers root directory for Dockerfiles and creates
// module definitions for any containers not explicitly defined.
// This enables convention-over-configuration: {containersRoot}/{name}/Dockerfile automatically
// becomes a module with moniker={name} using the "container" template.
//
// Parameters:
//   - repoRoot: repository root directory
//   - explicitMonikers: set of monikers already defined in repository.yml
//   - containersRoot: containers root directory name (e.g., "containers")
//
// Returns discovered modules (empty slice if none found or on error).
func DiscoverContainerModules(repoRoot string, explicitMonikers map[string]bool, containersRoot string) []Module {
	if containersRoot == "" {
		containersRoot = "containers"
	}
	containersDir := filepath.Join(repoRoot, containersRoot)

	// Check if containers directory exists
	if _, err := os.Stat(containersDir); os.IsNotExist(err) {
		return nil
	}

	// Find all Dockerfiles
	pattern := filepath.Join(containersDir, "*", "Dockerfile")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}

	var modules []Module
	for _, dockerfile := range matches {
		// Extract container name from path (parent directory name)
		dir := filepath.Dir(dockerfile)
		moniker := filepath.Base(dir)

		// Skip if already explicitly declared
		if explicitMonikers[moniker] {
			continue
		}

		// Skip non-OCI containers (internal tools like cgo-oci, drawio-oci)
		// Convention: only auto-discover containers with -oci suffix
		if !strings.HasSuffix(moniker, "-oci") {
			continue
		}

		// Create module with container template
		modules = append(modules, Module{
			Moniker:     moniker,
			Name:        deriveContainerName(moniker),
			Description: fmt.Sprintf("Auto-discovered container from %s/%s", containersRoot, moniker),
			Template:    "container",
			DependsOn:   []string{},
		})
	}

	// Sort for deterministic ordering
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Moniker < modules[j].Moniker
	})

	return modules
}

// deriveContainerName creates a human-readable name from a container moniker.
// Examples:
//   - "pdf-oci" -> "PDF Container"
//   - "nginx-oci" -> "Nginx Container"
//   - "mkdocs-dev-oci" -> "Mkdocs Dev Container"
func deriveContainerName(moniker string) string {
	// Remove -oci suffix
	name := strings.TrimSuffix(moniker, "-oci")

	// Split on hyphens and title-case each word
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if len(part) > 0 {
			// Special case: all caps for known acronyms
			upper := strings.ToUpper(part)
			if upper == "PDF" || upper == "CLI" || upper == "API" || upper == "OCI" {
				parts[i] = upper
			} else {
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
	}

	return strings.Join(parts, " ") + " Container"
}
