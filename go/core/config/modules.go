package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Module represents a single module definition.
type Module struct {
	Moniker       string                `yaml:"moniker"`
	Description   string                `yaml:"description"`
	Group         string                `yaml:"group,omitempty"`         // Group name for depends_on expansion
	DependsOn     []string              `yaml:"depends_on,omitempty"`
	DependsOnCI   []string              `yaml:"depends_on_ci,omitempty"`  // CI artifact dependencies (merged into DependsOn)
	CIDeps        []string              `yaml:"-"`                        // Computed: CI artifact deps for dispatch layering
	Metadata      map[string]string     `yaml:"metadata,omitempty"`       // Generic key-value store for module-specific data
	Versioning    *ModuleVersioning     `yaml:"versioning,omitempty"`
	Components    ModuleComponents      `yaml:"components"`        // Component types for this module (required)
	Linting       *ModuleLinting        `yaml:"linting,omitempty"` // Linting configuration overrides

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

// FacetSeparator is the character used to separate parent name from facet suffix
// in synthetic component names (e.g., "godog~specs").
const FacetSeparator = "~"

// FacetConfig defines how a facet maps to a component kind.
type FacetConfig struct {
	Kind       string // blueprints.yml component-kind name
	NameSuffix string // suffix appended after "~"
}

// DefaultFacetMappings returns the built-in facet-to-kind mappings.
func DefaultFacetMappings() map[string]FacetConfig {
	return map[string]FacetConfig{
		"specs":  {Kind: "gherkin", NameSuffix: "specs"},
		"design": {Kind: "structurizr", NameSuffix: "design"},
		"docs":   {Kind: "docs-assets", NameSuffix: "docs"},
	}
}

// FacetDeclaration represents a facet that can be either:
//   - Simple: a list of glob patterns (inherits parent root)
//   - Rooted: a mapping with explicit root and patterns (independent root)
type FacetDeclaration struct {
	Root     string   // Independent root path (empty = inherit parent root)
	Patterns []string // Glob patterns for source files
}

// IsEmpty returns true if the facet has no patterns declared.
func (fd *FacetDeclaration) IsEmpty() bool {
	return fd == nil || len(fd.Patterns) == 0
}

// UnmarshalYAML handles both list and mapping forms:
//
//	specs: ["**/*.feature"]           -> FacetDeclaration{Patterns: [...]}
//	design:
//	  root: specs/foo/.design
//	  patterns: ["**/*.dsl"]          -> FacetDeclaration{Root: "...", Patterns: [...]}
func (fd *FacetDeclaration) UnmarshalYAML(node *yaml.Node) error {
	// Sequence node -> simple list of patterns
	if node.Kind == yaml.SequenceNode {
		return node.Decode(&fd.Patterns)
	}

	// Mapping node -> rooted facet
	if node.Kind == yaml.MappingNode {
		type raw struct {
			Root     string   `yaml:"root"`
			Patterns []string `yaml:"patterns"`
		}
		var r raw
		if err := node.Decode(&r); err != nil {
			return err
		}
		fd.Root = r.Root
		fd.Patterns = r.Patterns
		return nil
	}

	return fmt.Errorf("facet: expected sequence or mapping, got %v", node.Kind)
}

// Clone creates a deep copy of FacetDeclaration.
func (fd *FacetDeclaration) Clone() *FacetDeclaration {
	if fd == nil {
		return nil
	}
	return &FacetDeclaration{
		Root:     fd.Root,
		Patterns: slices.Clone(fd.Patterns),
	}
}

// expandFacets processes all components and expands facet declarations
// (specs, design, docs) into synthetic ComponentEntry objects.
// For each component with non-empty facet fields, a synthetic entry is created
// with name "{parent}~{facet}", type from FacetConfig.Kind, and source patterns
// from the facet field. The root is either inherited from the parent or taken
// from the facet's own Root field (for rooted facets).
func expandFacets(components ModuleComponents) error {
	facets := DefaultFacetMappings()
	type addition struct {
		name  string
		entry *ComponentEntry
	}
	var additions []addition

	for parentName, entry := range components {
		if entry == nil {
			continue
		}

		facetFields := map[string]*FacetDeclaration{
			"specs":  entry.Specs,
			"design": entry.Design,
			"docs":   entry.Docs,
		}

		for facetName, decl := range facetFields {
			if decl == nil || decl.IsEmpty() {
				continue
			}

			cfg := facets[facetName]
			synthName := parentName + FacetSeparator + cfg.NameSuffix

			if _, exists := components[synthName]; exists {
				return fmt.Errorf(
					"facet expansion conflict: component %q would be created from %s.%s, "+
						"but a component with that name already exists",
					synthName, parentName, facetName,
				)
			}

			// Use facet's own root if specified, otherwise inherit parent root
			root := entry.Root
			if decl.Root != "" {
				root = decl.Root
			}

			additions = append(additions, addition{
				name: synthName,
				entry: &ComponentEntry{
					Type: cfg.Kind,
					Root: root,
					Patterns: &ComponentPatterns{
						Source: slices.Clone(decl.Patterns),
					},
					ParentComponent: parentName,
					FacetName:       facetName,
				},
			})
		}
	}

	for _, a := range additions {
		components[a.name] = a.entry
	}

	return nil
}

// IsFacetComponent returns true if this component was created by facet expansion.
func (ce *ComponentEntry) IsFacetComponent() bool {
	return ce != nil && ce.FacetName != ""
}

// expandCompanions processes all components and creates companion ComponentEntry
// objects from the With field. Each companion shares the parent's root and uses
// its type as its name. Existing components with the same name are not overwritten.
func expandCompanions(components ModuleComponents) error {
	type addition struct {
		name  string
		entry *ComponentEntry
	}
	var additions []addition

	for _, entry := range components {
		if entry == nil || len(entry.With) == 0 {
			continue
		}
		for _, companionType := range entry.With {
			if _, exists := components[companionType]; exists {
				continue // Skip if already declared
			}
			alreadyQueued := false
			for _, a := range additions {
				if a.name == companionType {
					alreadyQueued = true
					break
				}
			}
			if alreadyQueued {
				continue
			}
			additions = append(additions, addition{
				name: companionType,
				entry: &ComponentEntry{
					Name:        companionType,
					Type:        companionType,
					Root:        entry.Root,
					IsCompanion: true,
				},
			})
		}
	}

	for _, a := range additions {
		components[a.name] = a.entry
	}
	return nil
}

// extractComponentOrder walks a Module YAML node to find the "components" key
// and returns component names in declaration order.
func extractComponentOrder(moduleNode *yaml.Node) []string {
	if moduleNode.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(moduleNode.Content)-1; i += 2 {
		keyNode := moduleNode.Content[i]
		valNode := moduleNode.Content[i+1]
		if keyNode.Value != "components" {
			continue
		}

		if valNode.Kind != yaml.SequenceNode {
			return nil
		}

		// List format: derive names from each item's name/root/type
		facets := DefaultFacetMappings()
		var order []string
		for _, itemNode := range valNode.Content {
			if itemNode.Kind != yaml.MappingNode {
				continue
			}
			var entry struct {
				Name   string            `yaml:"name"`
				Root   string            `yaml:"root"`
				Type   string            `yaml:"type"`
				Specs  *FacetDeclaration `yaml:"specs"`
				Design *FacetDeclaration `yaml:"design"`
				Docs   *FacetDeclaration `yaml:"docs"`
				With   []string          `yaml:"with"`
			}
			_ = itemNode.Decode(&entry)
			name := entry.Name
			if name == "" {
				name = DefaultDeriveName(entry.Root)
			}
			if name == "" {
				name = entry.Type
			}
			order = append(order, name)

			// Append facet-derived names
			facetFields := map[string]*FacetDeclaration{
				"specs":  entry.Specs,
				"design": entry.Design,
				"docs":   entry.Docs,
			}
			for facetName, decl := range facetFields {
				if decl != nil && !decl.IsEmpty() {
					cfg := facets[facetName]
					order = append(order, name+FacetSeparator+cfg.NameSuffix)
				}
			}

			// Append companion-derived names
			for _, companionType := range entry.With {
				order = append(order, companionType)
			}
		}
		return order
	}
	return nil
}

// DefaultDeriveName infers a component name from a root path using last-segment heuristics.
// It strips leading dots, replaces underscores with hyphens, and truncates to 16 characters.
// Returns empty string if rootPath is empty or "/".
//
// Examples:
//
//	"go/adapters/godog"   → "godog"
//	"specs/docs/.design"  → "design"
//	"go/core"             → "core"
//	"contracts/core/0.1.0" → "0-1-0"  (version paths need explicit name)
func DefaultDeriveName(rootPath string) string {
	if rootPath == "" || rootPath == "/" {
		return ""
	}
	// Use path.Base for forward-slash paths (YAML always uses forward slashes)
	name := path.Base(filepath.ToSlash(rootPath))
	name = strings.TrimLeft(name, ".")
	name = strings.ReplaceAll(name, "_", "-")
	// Strip dots in remaining name (e.g., "0.1.0" → "0-1-0" isn't great, but these need explicit names)
	if len(name) > 16 {
		name = name[:16]
	}
	if name == "" {
		return ""
	}
	return name
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

// GetEvidenceBooks returns the list of evidence book names from evidence-book components.
// The book name is taken from the component's config["book"], falling back to the component name.
// Names are sorted for consistent ordering.
func (m *Module) GetEvidenceBooks() []string {
	evidenceBooks := m.Components.GetComponentsByType("evidence-book")
	if len(evidenceBooks) == 0 {
		return nil
	}
	names := make([]string, 0, len(evidenceBooks))
	for name, entry := range evidenceBooks {
		bookName := name
		if entry != nil && entry.GetConfig("book") != "" {
			bookName = entry.GetConfig("book")
		}
		names = append(names, bookName)
	}
	sort.Strings(names)
	return names
}

// GetReleaseBundle returns the release bundle configuration from a release-bundle component.
// Returns nil if no release-bundle component exists.
func (m *Module) GetReleaseBundle() *ReleaseBundle {
	_, comp := m.Components.GetFirstByType("release-bundle")
	if comp == nil {
		return nil
	}
	return comp.Bundle
}


// GetPushableContainerComponents returns sorted names of container components
// where push is not explicitly false. Used by CI config to generate matrix builds.
func (m *Module) GetPushableContainerComponents() []string {
	if m == nil || m.Components == nil {
		return nil
	}
	var names []string
	for name, entry := range m.Components {
		compType := name
		if entry != nil && entry.Type != "" {
			compType = entry.Type
		}
		if compType != "container" {
			continue
		}
		if entry != nil && entry.DockerBuild != nil && entry.DockerBuild.Push != nil && !*entry.DockerBuild.Push {
			continue
		}
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
// - Container modules (has container component type)
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

	// Container modules (has container component)
	if m.Components.HasComponentType("container") {
		return true
	}

	return false
}

// ReleaseBundle configures how the release module creates GitHub releases.
type ReleaseBundle struct {
	TitleFormat string                  `yaml:"title_format"` // Title template, e.g., "{clie} ({clie_version}) + {eac} ({eac_version})"
	Headline    map[string]string       `yaml:"headline"`     // Map of label -> moniker for title, e.g., {"clie": "clie", "eac": "eac-ext"}
	Categories  []ReleaseBundleCategory `yaml:"categories"`   // Grouped modules for release notes
}

// Clone returns a deep copy of the ReleaseBundle.
func (r *ReleaseBundle) Clone() *ReleaseBundle {
	clone := &ReleaseBundle{
		TitleFormat: r.TitleFormat,
	}
	if r.Headline != nil {
		clone.Headline = make(map[string]string, len(r.Headline))
		for k, v := range r.Headline {
			clone.Headline[k] = v
		}
	}
	if r.Categories != nil {
		clone.Categories = make([]ReleaseBundleCategory, len(r.Categories))
		for i, cat := range r.Categories {
			clone.Categories[i] = ReleaseBundleCategory{
				Name:        cat.Name,
				Description: cat.Description,
			}
			if cat.Modules != nil {
				clone.Categories[i].Modules = make([]string, len(cat.Modules))
				copy(clone.Categories[i].Modules, cat.Modules)
			}
		}
	}
	return clone
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
	Handler           string           `yaml:"handler,omitempty"`          // Explicit build handler override
	BinaryName        string           `yaml:"binary_name,omitempty"`     // Output binary name (for artifact matrix expansion)
	ArtifactMatrixRef string           `yaml:"artifact_matrix,omitempty"` // Reference to an artifact matrix name
	Artifacts         []ModuleArtifact `yaml:"artifacts,omitempty"`       // Artifacts to produce
	Options           *BuildOptions    `yaml:"options,omitempty"`         // Build behavior options
	PostBuild         *PostBuildConfig `yaml:"post_build,omitempty"`      // Post-build actions
}

// PostBuildConfig contains post-build actions for a component.
// Configured at the component level within a module's build configuration.
type PostBuildConfig struct {
	// CopyFiles specifies files to copy after build.
	// Each entry maps a source path (relative to component root) to a
	// target path (relative to workspace root).
	// The source path supports glob patterns (*, **, ?); when a glob is used,
	// the target is treated as a directory and matched files are copied into it
	// preserving their relative directory structure.
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
		Handler:           b.Handler,
		BinaryName:        b.BinaryName,
		ArtifactMatrixRef: b.ArtifactMatrixRef,
	}
	if b.Artifacts != nil {
		clone.Artifacts = make([]ModuleArtifact, len(b.Artifacts))
		copy(clone.Artifacts, b.Artifacts)
	}
	if b.Options != nil {
		clone.Options = &BuildOptions{}
	}
	if b.PostBuild != nil {
		clone.PostBuild = &PostBuildConfig{}
		if len(b.PostBuild.CopyFiles) > 0 {
			clone.PostBuild.CopyFiles = make([]CopyFileEntry, len(b.PostBuild.CopyFiles))
			copy(clone.PostBuild.CopyFiles, b.PostBuild.CopyFiles)
		}
	}
	return clone
}

// ModuleLinting configures linting behavior for a module.
type ModuleLinting struct {
	// Enabled lists specific lint providers to use (empty = use all applicable from lint-providers.yml)
	Enabled []string `yaml:"enabled,omitempty"`

	// Disabled lists lint providers to skip, or "all" to disable linting entirely
	Disabled []string `yaml:"disabled,omitempty"`
}

// ModuleComponents is a map of component name to its configuration.
// The first component in the map becomes the default for build operations.
// Components are declared as a YAML list; names are auto-inferred from root paths.
type ModuleComponents map[string]*ComponentEntry

// UnmarshalYAML implements custom unmarshaling for ModuleComponents.
// Components must be declared as a YAML sequence (list format).
// Names are auto-inferred: explicit name > last segment of root > type fallback.
func (mc *ModuleComponents) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("components: expected sequence (list format), got %v", node.Kind)
	}

	result := make(ModuleComponents)
	for i, itemNode := range node.Content {
		var entry ComponentEntry
		if itemNode.Kind == yaml.MappingNode {
			type rawComponentEntry ComponentEntry
			if err := itemNode.Decode((*rawComponentEntry)(&entry)); err != nil {
				return fmt.Errorf("components[%d]: %w", i, err)
			}
		} else {
			return fmt.Errorf("components[%d]: expected mapping, got %v", i, itemNode.Kind)
		}

		// Determine component name: explicit > derive from root > type fallback
		name := entry.Name
		if name == "" {
			name = DefaultDeriveName(entry.Root)
		}
		if name == "" {
			name = entry.Type
		}
		if name == "" {
			return fmt.Errorf("components[%d]: cannot determine component name (no name, root, or type)", i)
		}

		if _, exists := result[name]; exists {
			return fmt.Errorf("components[%d]: duplicate component name %q", i, name)
		}
		result[name] = &entry
	}

	// Expand facets into synthetic components
	if err := expandFacets(result); err != nil {
		return err
	}

	// Expand companion components from with: fields
	if err := expandCompanions(result); err != nil {
		return err
	}

	*mc = result
	return nil
}

// MarshalYAML implements custom marshaling for ModuleComponents.
// Outputs the list (sequence) format required by UnmarshalYAML, so that a
// config round-tripped through yaml.Marshal → yaml.Unmarshal remains valid.
// Synthetic entries (facet-expanded and companion components) are omitted
// because they are reconstructed at parse time from their parent declarations.
func (mc ModuleComponents) MarshalYAML() (interface{}, error) {
	// Collect keys to produce deterministic output order
	keys := make([]string, 0, len(mc))
	for k := range mc {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	entries := make([]*ComponentEntry, 0, len(mc))
	for _, k := range keys {
		entry := mc[k]
		if entry == nil {
			continue
		}
		// Skip synthetic entries — they are rebuilt from parent declarations at parse time
		if entry.IsFacetComponent() || entry.IsCompanion || entry.AutoBDDRunner {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

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
	// Name is an optional explicit name override. When components are declared in list format,
	// the name is auto-inferred from the root path. This field allows overriding the inferred name.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Type is the component type (e.g., "go", "book"). If omitted, the map key (name) is used as the type.
	// This allows multiple components of the same type with different names.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Root is the root path for this component (relative to repo root)
	Root string `yaml:"root,omitempty" json:"root,omitempty"`

	// Patterns contains optional pattern overrides for this component
	Patterns *ComponentPatterns `yaml:"patterns,omitempty" json:"patterns,omitempty"`

	// Build contains build configuration for this component (artifacts, handler override)
	Build *ModuleBuild `yaml:"build,omitempty" json:"build,omitempty"`

	// DockerBuild contains Docker build configuration (for container components)
	DockerBuild *DockerBuildConfig `yaml:"docker_build,omitempty" json:"docker_build,omitempty"`

	// Bundle contains release bundle configuration (for release-bundle components)
	Bundle *ReleaseBundle `yaml:"bundle,omitempty" json:"bundle,omitempty"`

	// Amp contains per-operation resource amplifiers
	Amp *AmpConfig `yaml:"amp,omitempty" json:"amp,omitempty"`

	// Config contains component-specific configuration (e.g., book name, theme).
	// Values here override defaults from blueprints.yml component-kinds.
	Config map[string]string `yaml:"config,omitempty" json:"config,omitempty"`

	// ComponentGroup is the group name for this component.
	// Components with the same group can be referenced as a single dependency target.
	// Defaults to the component name if not explicitly set (via applyComponentGroupDefaults).
	ComponentGroup string `yaml:"component_group,omitempty" json:"component_group,omitempty"`

	// DependsOn lists component names this component depends on within the same module.
	// Used for intra-module build ordering (e.g., pdf-render depends on base-site).
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`

	// ComponentDeps lists inter-module component dependencies using the addressing
	// format "module:componentName[:componentType[:tool]]". Unlike module-level
	// depends_on (which causes all-to-all fan-out), component_deps enables narrowed
	// dependency matching: only the specified remote components are awaited.
	ComponentDeps []string `yaml:"component_deps,omitempty" json:"component_deps,omitempty"`

	// Facet fields: expand into synthetic components with type from DefaultFacetMappings.
	// Each accepts either a list of patterns (inherits parent root) or a mapping with
	// explicit root and patterns (independent root).
	Specs  *FacetDeclaration `yaml:"specs,omitempty" json:"specs,omitempty"`
	Design *FacetDeclaration `yaml:"design,omitempty" json:"design,omitempty"`
	Docs   *FacetDeclaration `yaml:"docs,omitempty" json:"docs,omitempty"`

	// With lists companion component types to auto-create alongside this component.
	// Each companion shares this component's root and uses its type as its name.
	// Example: with: [assets, markdown, yaml]
	With []string `yaml:"with,omitempty" json:"with,omitempty"`

	// Metadata set on facet-expanded synthetic components (not serialized to YAML).
	ParentComponent string `yaml:"-" json:"-"` // Name of the parent component
	FacetName       string `yaml:"-" json:"-"` // "specs", "design", or "docs"

	// AutoBDDRunner indicates this component was auto-created as a BDD runner
	// from a parent component's specs: facet. Set during ApplyComponentDefaults.
	AutoBDDRunner bool `yaml:"-" json:"-"`

	// IsCompanion indicates this component was auto-created by expandCompanions from
	// a parent's with: field. These are reconstructed at parse time from with: declarations
	// and should not be emitted during marshaling.
	IsCompanion bool `yaml:"-" json:"-"`

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

// GetComponentDeps returns the list of inter-module component dependencies.
func (ce *ComponentEntry) GetComponentDeps() []string {
	if ce == nil {
		return nil
	}
	return ce.ComponentDeps
}

// ComponentPatterns contains optional file pattern overrides for a component.
type ComponentPatterns struct {
	Source  []string `yaml:"source,omitempty" json:"source,omitempty"`
	Tests   []string `yaml:"tests,omitempty" json:"tests,omitempty"`
	Config  []string `yaml:"config,omitempty" json:"config,omitempty"`
	Data    []string `yaml:"data,omitempty" json:"data,omitempty"`    // Owned files not processed by tools (e.g., testdata, fixtures)
	Exclude []string `yaml:"exclude,omitempty" json:"exclude,omitempty"` // Patterns to exclude from ownership
}

// AllOwnershipPatterns returns all patterns that contribute to file ownership.
// This includes source, tests, config, and data patterns (everything except excludes).
func (p *ComponentPatterns) AllOwnershipPatterns() []string {
	if p == nil {
		return nil
	}
	var all []string
	all = append(all, p.Source...)
	all = append(all, p.Tests...)
	all = append(all, p.Config...)
	all = append(all, p.Data...)
	return all
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
				Name:            entry.Name,
				Type:            entry.Type,
				Root:            entry.Root,
				ComponentGroup:  entry.ComponentGroup,
				ParentComponent: entry.ParentComponent,
				FacetName:       entry.FacetName,
				AutoBDDRunner:   entry.AutoBDDRunner,
				Resolved:        entry.Resolved,
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
				if entry.Patterns.Exclude != nil {
					clonedEntry.Patterns.Exclude = make([]string, len(entry.Patterns.Exclude))
					copy(clonedEntry.Patterns.Exclude, entry.Patterns.Exclude)
				}
			}
			if entry.Build != nil {
				clonedEntry.Build = entry.Build.Clone()
			}
			if entry.DockerBuild != nil {
				clonedEntry.DockerBuild = entry.DockerBuild.Clone()
			}
			if entry.Bundle != nil {
				clonedEntry.Bundle = entry.Bundle.Clone()
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
			if entry.ComponentDeps != nil {
				clonedEntry.ComponentDeps = make([]string, len(entry.ComponentDeps))
				copy(clonedEntry.ComponentDeps, entry.ComponentDeps)
			}
			clonedEntry.Specs = entry.Specs.Clone()
			clonedEntry.Design = entry.Design.Clone()
			clonedEntry.Docs = entry.Docs.Clone()
			if entry.With != nil {
				clonedEntry.With = make([]string, len(entry.With))
				copy(clonedEntry.With, entry.With)
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

// GetAllRoots returns a map of component name to root path.
func (mc ModuleComponents) GetAllRoots() map[string]string {
	if mc == nil {
		return nil
	}
	roots := make(map[string]string, len(mc))
	for name, entry := range mc {
		if entry != nil {
			roots[name] = entry.Root
		}
	}
	return roots
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

// GetFirstByType returns the first component matching the given type.
// Returns the component name and entry, or ("", nil) if not found.
// The type is determined by ComponentEntry.Type if set, otherwise by the map key (name).
func (mc ModuleComponents) GetFirstByType(typeName string) (string, *ComponentEntry) {
	if mc == nil {
		return "", nil
	}
	for name, entry := range mc {
		t := name
		if entry != nil && entry.Type != "" {
			t = entry.Type
		}
		if t == typeName {
			return name, entry
		}
	}
	return "", nil
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

// GetAllComponentDeps returns all inter-module component_deps across all components,
// deduplicated and sorted.
func (mc ModuleComponents) GetAllComponentDeps() []string {
	if mc == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []string
	for _, entry := range mc {
		if entry == nil {
			continue
		}
		for _, dep := range entry.ComponentDeps {
			if !seen[dep] {
				seen[dep] = true
				result = append(result, dep)
			}
		}
	}
	sort.Strings(result)
	return result
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

// FindPrimaryComponent returns the first primary source component (go or typescript).
// Returns the component name and entry, or ("", nil) if none found.
// Used by template companion expansion to determine where to attach companions.
func (mc ModuleComponents) FindPrimaryComponent() (string, *ComponentEntry) {
	for _, typeName := range []string{"go", "typescript", "python", "dotnet", "rust"} {
		if name, entry := mc.GetFirstByType(typeName); entry != nil {
			return name, entry
		}
	}
	return "", nil
}

// GetBuildableRoot returns the root path of the first buildable component
// (go or typescript). Returns empty string if none found.
// Searches by component type, not name, to support both map and list formats.
func (mc ModuleComponents) GetBuildableRoot() string {
	if mc == nil {
		return ""
	}
	// Prefer "go" component type
	if _, entry := mc.GetFirstByType("go"); entry != nil && entry.Root != "" {
		return entry.Root
	}
	// Try "typescript" component type
	if _, entry := mc.GetFirstByType("typescript"); entry != nil && entry.Root != "" {
		return entry.Root
	}
	return ""
}

// GetComponentGroup returns the component group for a named component.
// Returns empty string if the component doesn't exist or has no group.
func (mc ModuleComponents) GetComponentGroup(compName string) string {
	if mc == nil {
		return ""
	}
	entry, ok := mc[compName]
	if !ok || entry == nil {
		return ""
	}
	return entry.ComponentGroup
}

// applyComponentGroupDefaults sets ComponentGroup to the component name
// for any entry that doesn't already have an explicit ComponentGroup.
// This provides a self-named default so that every component belongs to a group.
func (mc ModuleComponents) applyComponentGroupDefaults() {
	for name, entry := range mc {
		if entry == nil {
			continue
		}
		if entry.ComponentGroup == "" {
			entry.ComponentGroup = name
		}
	}
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

// GetDockerBuildConfig returns the docker_build configuration from the container component.
// Returns nil if no container component exists or no docker_build is configured.
// Searches by component type to support both map and list formats.
func (m *Module) GetDockerBuildConfig() *DockerBuildConfig {
	_, entry := m.Components.GetFirstByType("container")
	if entry == nil || entry.DockerBuild == nil {
		return nil
	}
	return entry.DockerBuild
}

// GetDockerBuildConfigForComponent returns docker_build from a named component.
// Falls back to container type lookup.
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
			m.Description = m.Moniker
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

// ApplyComponentDefaults resolves component roots and patterns from blueprints.yml component-kinds.
// This should be called after ComponentTypes are loaded.
func (c *RepositoryConfig) ApplyComponentDefaults(compTypes *ComponentKindsConfig, repoRoot string) {
	specsRoot := c.Paths.SpecsRoot
	if specsRoot == "" {
		specsRoot = "specs"
	}

	// Collect all module monikers for cross-module specs deduplication
	monikers := make(map[string]bool, len(c.Modules))
	for _, m := range c.Modules {
		monikers[m.Moniker] = true
	}

	// Pass 1: Resolve roots, default specs, BDD specs, and design components
	for i := range c.Modules {
		m := &c.Modules[i]
		m.resolveComponentRoots(compTypes)
		m.inferDefaultSpecsComponent(repoRoot, specsRoot)
		m.inferBDDSpecsRoots(repoRoot, monikers)
		m.inferDesignComponents(repoRoot, specsRoot)
		m.resolveDerivedPaths()
	}

	// Pass 2: Claim unclaimed <specsRoot>/{moniker}_* directories.
	// This runs after all modules have created their gherkin components,
	// so we can check for conflicts (e.g., specs/eac_* claimed by commands
	// module via inferBDDSpecsRoots should not also be claimed by eac module).
	claimedRoots := make(map[string]bool)
	for _, m := range c.Modules {
		for _, entry := range m.Components {
			if entry != nil && (entry.Type == "gherkin" || entry.Type == "specs") {
				claimedRoots[entry.Root] = true
			}
		}
	}
	for i := range c.Modules {
		c.Modules[i].inferSubSpecsComponents(repoRoot, specsRoot, claimedRoots)
	}
}

// expandBDDRunners creates BDD runner components for primary language components
// whose component kind defines a bdd_runner.
// The runner root is set to the parent component's root since step implementations
// live within the source tree (e.g., godog steps live under the Go component root).
// Default patterns from blueprints.yml component-kinds are applied during resolveComponentRoots.
//
// Called during resolveComponentRoots after component kinds are loaded.
func (m *Module) expandBDDRunners(compTypes *ComponentKindsConfig) {
	if compTypes == nil {
		return
	}

	type addition struct {
		name  string
		entry *ComponentEntry
	}
	var additions []addition

	for compName, entry := range m.Components {
		if entry == nil {
			continue
		}

		// Get the component's kind
		compType := compName
		if entry.Type != "" {
			compType = entry.Type
		}
		ct := compTypes.Get(compType)
		if ct == nil || ct.BDDRunner == "" {
			continue
		}

		runnerKind := ct.BDDRunner

		// Skip if a component of this runner type already exists
		if m.Components.HasComponentType(runnerKind) {
			continue
		}

		// Skip if already queued
		alreadyQueued := false
		for _, a := range additions {
			if a.entry.Type == runnerKind {
				alreadyQueued = true
				break
			}
		}
		if alreadyQueued {
			continue
		}

		// Runner root = parent component root (steps live in the source tree).
		// The specs facet root points to where .feature files are, NOT where
		// step implementations are, so we always use the parent root.
		runnerRoot := entry.Root

		// Use runner kind as name, but add suffix if a component with that name
		// already exists (e.g., adapters module has a Go component derived-named "godog"
		// from root go/adapters/godog — avoid overwriting it).
		runnerName := runnerKind
		if _, exists := m.Components[runnerName]; exists {
			runnerName = runnerKind + "-runner"
		}

		additions = append(additions, addition{
			name: runnerName,
			entry: &ComponentEntry{
				Name:          runnerName,
				Type:          runnerKind,
				Root:          runnerRoot,
				AutoBDDRunner: true,
			},
		})
	}

	for _, a := range additions {
		m.Components[a.name] = a.entry
		m.ComponentOrder = append(m.ComponentOrder, a.name)
	}
}

// inferDefaultSpecsComponent creates gherkin components for specs directories
// that belong to this module by convention:
//   - specs/{moniker}   — the module's default specs root
//   - specs/{moniker}_* — sub-specs directories (e.g., specs/adapters_godog for the adapters module)
//
// Directories that already have a gherkin/specs component are skipped.
// Only directories that exist on disk are added.
func (m *Module) inferDefaultSpecsComponent(repoRoot, specsRoot string) {
	if repoRoot == "" {
		return
	}

	type addition struct {
		name  string
		entry *ComponentEntry
	}
	var additions []addition
	seen := make(map[string]bool)

	// Record specs roots already declared in module components
	for _, entry := range m.Components {
		if entry == nil {
			continue
		}
		compType := entry.Type
		if compType == "" {
			compType = entry.Name
		}
		if compType == "gherkin" || compType == "specs" || compType == "structurizr" {
			seen[entry.Root] = true
		}
	}

	addDir := func(specsRoot, compName string) {
		if seen[specsRoot] {
			return
		}
		absRoot := filepath.Join(repoRoot, specsRoot)
		info, err := os.Stat(absRoot)
		if err != nil || !info.IsDir() {
			return
		}
		if _, exists := m.Components[compName]; exists {
			compName = compName + "-specs"
		}
		seen[specsRoot] = true
		additions = append(additions, addition{
			name: compName,
			entry: &ComponentEntry{
				Name: compName,
				Type: "gherkin",
				Root: specsRoot,
			},
		})
	}

	// Default specs root: <specsRoot>/{moniker}
	addDir(path.Join(specsRoot, m.Moniker), m.Moniker+"-specs")

	// Sort additions by name for deterministic ordering
	sort.Slice(additions, func(i, j int) bool {
		return additions[i].name < additions[j].name
	})

	for _, a := range additions {
		m.Components[a.name] = a.entry
		m.ComponentOrder = append(m.ComponentOrder, a.name)
	}
}

// inferSubSpecsComponents claims unclaimed specs/{moniker}_* directories.
// This runs in pass 2 after all modules have processed their gherkin components,
// so we skip directories already claimed by another module's inferBDDSpecsRoots.
// Example: specs/adapters_godog belongs to the adapters module; specs/eac_create
// belongs to the commands module (claimed via godog_test.go scanning in pass 1).
func (m *Module) inferSubSpecsComponents(repoRoot, specsRoot string, claimedRoots map[string]bool) {
	if repoRoot == "" {
		return
	}

	type addition struct {
		name  string
		entry *ComponentEntry
	}
	var additions []addition

	specsDir := filepath.Join(repoRoot, specsRoot)
	prefix := m.Moniker + "_"
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return
	}
	for _, dirEntry := range entries {
		if !dirEntry.IsDir() {
			continue
		}
		name := dirEntry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		specsPath := path.Join(specsRoot, name)
		if claimedRoots[specsPath] {
			continue
		}
		compName := name
		if _, exists := m.Components[compName]; exists {
			compName = name + "-specs"
		}
		claimedRoots[specsPath] = true
		additions = append(additions, addition{
			name: compName,
			entry: &ComponentEntry{
				Name: compName,
				Type: "gherkin",
				Root: specsPath,
			},
		})
	}

	sort.Slice(additions, func(i, j int) bool {
		return additions[i].name < additions[j].name
	})
	for _, a := range additions {
		m.Components[a.name] = a.entry
		m.ComponentOrder = append(m.ComponentOrder, a.name)
	}
}

// inferDesignComponents auto-discovers structurizr design components from
// workspace.dsl files at conventional paths. Each module:component pair with a
// workspace.dsl gets its own design component and builds as an independent UoW.
//
// Convention:
//   - specs/{moniker}/.design/workspace.dsl     → module-level design ({moniker}~design)
//   - specs/{moniker}_{qualifier}/.design/workspace.dsl → component-level design ({qualifier}~design)
//
// Components already having a design facet (from explicit declaration) are skipped.
func (m *Module) inferDesignComponents(repoRoot, specsRoot string) {
	if repoRoot == "" {
		return
	}

	// Collect existing design component roots (from explicit facet declarations)
	// to avoid creating duplicates pointing to the same workspace.dsl.
	hasDesignRoot := make(map[string]bool)
	for _, entry := range m.Components {
		if entry != nil && entry.FacetName == "design" {
			hasDesignRoot[entry.Root] = true
		}
	}

	type addition struct {
		name  string
		entry *ComponentEntry
	}
	var additions []addition

	addDesign := func(compName, designRoot string) {
		if hasDesignRoot[designRoot] {
			return
		}
		if _, exists := m.Components[compName]; exists {
			return
		}
		additions = append(additions, addition{
			name: compName,
			entry: &ComponentEntry{
				Name:      compName,
				Type:      "structurizr",
				Root:      designRoot,
				FacetName: "design",
				Patterns: &ComponentPatterns{
					Source: []string{"workspace.dsl", "**/*.dsl"},
				},
			},
		})
	}

	// 1. Module-level design: <specsRoot>/{moniker}/.design/workspace.dsl
	designRoot := path.Join(specsRoot, m.Moniker, ".design")
	designFile := filepath.Join(repoRoot, designRoot, "workspace.dsl")
	if _, err := os.Stat(designFile); err == nil {
		addDesign(m.Moniker+FacetSeparator+"design", designRoot)
	}

	// 2. Component-level design: <specsRoot>/{moniker}_{qualifier}/.design/workspace.dsl
	specsDir := filepath.Join(repoRoot, specsRoot)
	prefix := m.Moniker + "_"
	entries, err := os.ReadDir(specsDir)
	if err == nil {
		for _, dirEntry := range entries {
			if !dirEntry.IsDir() {
				continue
			}
			name := dirEntry.Name()
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			qualifier := strings.TrimPrefix(name, prefix)
			subDesignRoot := path.Join(specsRoot, name, ".design")
			subDesignFile := filepath.Join(repoRoot, subDesignRoot, "workspace.dsl")
			if _, err := os.Stat(subDesignFile); err == nil {
				addDesign(qualifier+FacetSeparator+"design", subDesignRoot)
			}
		}
	}

	// 3. Sub-component design: <specsRoot>/{moniker}/{sub}/.design/workspace.dsl
	moduleSpecsDir := filepath.Join(repoRoot, specsRoot, m.Moniker)
	subEntries, err := os.ReadDir(moduleSpecsDir)
	if err == nil {
		for _, dirEntry := range subEntries {
			if !dirEntry.IsDir() || dirEntry.Name() == ".design" {
				continue
			}
			subDesignRoot := path.Join(specsRoot, m.Moniker, dirEntry.Name(), ".design")
			subDesignFile := filepath.Join(repoRoot, subDesignRoot, "workspace.dsl")
			if _, err := os.Stat(subDesignFile); err == nil {
				addDesign(dirEntry.Name()+FacetSeparator+"design", subDesignRoot)
			}
		}
	}

	sort.Slice(additions, func(i, j int) bool {
		return additions[i].name < additions[j].name
	})
	for _, a := range additions {
		m.Components[a.name] = a.entry
		m.ComponentOrder = append(m.ComponentOrder, a.name)
	}
}

// reSpecsPath matches SpecsPath: "..." in godog_test.go files.
var reSpecsPath = regexp.MustCompile(`SpecsPath:\s*"([^"]+)"`)

// inferBDDSpecsRoots scans auto-BDD runner roots for test runner files and extracts
// specs directory references to create gherkin components. This allows specs directories
// referenced by godog_test.go SpecsPath fields to be automatically discovered as
// owned by the module that contains the runner, without explicit YAML declarations.
//
// allMonikers contains all module monikers in the repository. Specs directories whose
// name matches another module's moniker are skipped (they belong to that module's
// default specs root, not this module).
func (m *Module) inferBDDSpecsRoots(repoRoot string, allMonikers map[string]bool) {
	if repoRoot == "" {
		return
	}

	type addition struct {
		name  string
		entry *ComponentEntry
	}
	var additions []addition
	seen := make(map[string]bool)

	// Record specs roots already declared in module components
	for _, entry := range m.Components {
		if entry == nil {
			continue
		}
		compType := entry.Type
		if compType == "" {
			compType = entry.Name
		}
		if compType == "gherkin" || compType == "specs" {
			seen[entry.Root] = true
		}
	}

	// Scan all Go component roots for godog_test.go files (not just the single auto-BDD runner).
	// A module may have multiple Go components (e.g., commands module has go/commands/repository,
	// go/commands/test, etc.), each potentially containing godog runners that reference different
	// specs directories.
	for _, entry := range m.Components {
		if entry == nil || entry.Root == "" {
			continue
		}

		// Only scan components whose type has a godog BDD runner (i.e., Go components)
		compType := entry.Type
		if compType == "" {
			compType = entry.Name
		}
		if compType != "go" {
			continue
		}

		absRoot := filepath.Join(repoRoot, entry.Root)
		_ = filepath.Walk(absRoot, func(fpath string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || info.Name() != "godog_test.go" {
				return nil
			}

			content, readErr := os.ReadFile(fpath)
			if readErr != nil {
				return nil
			}

			matches := reSpecsPath.FindStringSubmatch(string(content))
			if len(matches) < 2 {
				return nil
			}

			// Resolve relative SpecsPath to repo-relative path
			relSpecsPath := matches[1]
			absSpecs := filepath.Join(filepath.Dir(fpath), relSpecsPath)
			absSpecs = filepath.Clean(absSpecs)

			repoRelSpecs, relErr := filepath.Rel(repoRoot, absSpecs)
			if relErr != nil {
				return nil
			}
			repoRelSpecs = filepath.ToSlash(repoRelSpecs)

			// Extract the top-level specs directory (specs/eac_design, not specs/eac_create/commit-message)
			parts := strings.SplitN(repoRelSpecs, "/", 3)
			if len(parts) < 2 {
				return nil
			}
			specsRoot := parts[0] + "/" + parts[1] // e.g., "specs/eac_design"
			specsDirName := parts[1]               // e.g., "eac_design"

			// Skip if this specs directory is another module's default specs root.
			// Each module's default specs root is specs/{moniker}. If the directory
			// name matches a known moniker, it belongs to that module, not this one.
			if specsDirName != m.Moniker && allMonikers[specsDirName] {
				return nil
			}

			if seen[specsRoot] {
				return nil
			}

			// Verify directory exists
			if info, statErr := os.Stat(filepath.Join(repoRoot, specsRoot)); statErr != nil || !info.IsDir() {
				return nil
			}

			seen[specsRoot] = true

			// Use the directory base as the component name
			compName := parts[1] // e.g., "eac_design"
			if _, exists := m.Components[compName]; exists {
				compName = compName + "-specs"
			}

			additions = append(additions, addition{
				name: compName,
				entry: &ComponentEntry{
					Name:            compName,
					Type:            "gherkin",
					Root:            specsRoot,
					ParentComponent: entry.Name,
					FacetName:       "specs",
				},
			})

			return nil
		})
	}

	// Sort additions by name for deterministic ordering
	sort.Slice(additions, func(i, j int) bool {
		return additions[i].name < additions[j].name
	})

	for _, a := range additions {
		m.Components[a.name] = a.entry
		m.ComponentOrder = append(m.ComponentOrder, a.name)
	}
}

// resolveComponentRoots resolves default roots and patterns for all components in the module.
func (m *Module) resolveComponentRoots(compTypes *ComponentKindsConfig) {
	if compTypes == nil {
		return
	}

	// Expand BDD runners before resolving defaults so they participate in resolution
	m.expandBDDRunners(compTypes)

	for compName, entry := range m.Components {
		// Get type (explicit or name as fallback)
		compType := compName
		if entry != nil && entry.Type != "" {
			compType = entry.Type
		}

		// Look up in blueprints.yml component-kinds by TYPE, not name
		ct := compTypes.Get(compType)
		if ct == nil {
			continue
		}

		// Ensure entry exists
		if entry == nil {
			entry = &ComponentEntry{}
			m.Components[compName] = entry
		}

		// Resolve root from component-kind default if not set.
		// For multi-container modules, use the component name for {moniker}
		// substitution when the component has an explicit name different from
		// the module moniker (e.g., "mkdocs-render-oci" in the "oci-tools" module).
		if entry.Root == "" && ct.DefaultRoot != "" {
			rootMoniker := m.Moniker
			if entry.Name != "" && entry.Name != m.Moniker {
				rootMoniker = entry.Name
			}
			entry.Root = ct.GetRoot(rootMoniker, "")
		}

		// Merge patterns from component-kind if not overridden
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
		if entry.Patterns.Data == nil && ct.Files != nil {
			entry.Patterns.Data = substituteMoniker(ct.Files.Data, m.Moniker)
		}

		// Apply ComponentGroup from component-kind if not already set
		if entry.ComponentGroup == "" && ct.ComponentGroup != "" {
			entry.ComponentGroup = ct.ComponentGroup
		}

		// Apply DependsOn from component-kind if not already set
		if len(entry.DependsOn) == 0 && len(ct.DependsOn) > 0 {
			entry.DependsOn = make([]string, len(ct.DependsOn))
			copy(entry.DependsOn, ct.DependsOn)
		}

		// Apply DockerBuildDefaults from component-kind if component has no docker_build
		if ct.DockerBuildDefaults != nil {
			if entry.DockerBuild == nil {
				entry.DockerBuild = cloneDockerBuildConfig(ct.DockerBuildDefaults)
			} else {
				mergeDockerBuildDefaults(entry.DockerBuild, ct.DockerBuildDefaults)
			}
		}

		// Substitute {moniker} in DockerBuild fields after defaults are applied.
		// SubstituteModuleParams runs before docker_build_defaults are merged,
		// so {moniker} placeholders from defaults remain unresolved.
		// For multi-container modules, use the component name instead.
		if entry.DockerBuild != nil {
			compMoniker := m.Moniker
			if entry.Name != "" && entry.Name != m.Moniker {
				compMoniker = entry.Name
			}
			substituteDockerBuildParams(entry.DockerBuild, map[string]string{"moniker": compMoniker})
		}

		// Apply component-kind defaults (convention-over-configuration)
		m.applyComponentDefaults(compName, entry, ct)

		entry.Resolved = true
	}
}

// applyComponentDefaults applies convention-over-configuration defaults from component-kind.
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
		if _, goEntry := m.Components.GetFirstByType("go"); goEntry != nil && goEntry.Root != "" {
			// Use forward slashes for consistent cross-platform paths
			return filepath.ToSlash(filepath.Join(goEntry.Root, "CHANGELOG.md"))
		}
		return ""
	default:
		return ""
	}
}

