package contracts

import "sort"

// RepositoryContract represents repository-level configuration.
type RepositoryContract struct {
	Repository RepositoryConfig `yaml:"repository"`
}

// RepositoryConfig contains repository-level settings.
type RepositoryConfig struct {
	Type             string           `yaml:"type"`                // mono | poly | adjunct
	Schemes          []string         `yaml:"schemes"`             // Valid versioning schemes (SemVer, CalVer)
	TrunkBranch      string           `yaml:"trunk_branch"`        // Main branch name
	MaxBranchAgeDays int              `yaml:"max_branch_age_days"` // Maximum branch age in days
	PR               PRConfig         `yaml:"pr"`                  // Pull request configuration
	Versioning       VersioningConfig `yaml:"versioning"`          // Versioning constraints
}

// PRConfig contains pull request workflow configuration.
type PRConfig struct {
	DeleteBranchOnMerge bool   `yaml:"delete_branch_on_merge"` // Auto-delete branch after merge
	MergeStrategy       string `yaml:"merge_strategy"`         // squash | merge | rebase
}

// VersioningConfig contains repository-wide versioning constraints.
type VersioningConfig struct {
	Constraint string `yaml:"constraint"` // unrestricted | patch-only | calver-only
}

// ModulesContract represents the modules configuration file.
type ModulesContract struct {
	Defaults *ModuleDefaults `yaml:"defaults,omitempty"` // Module-level defaults
	Modules  []BaseContract  `yaml:"modules"`            // Module definitions
}

// ModuleDefaults contains default values inherited by all modules.
type ModuleDefaults struct {
	Paths       DefaultPaths       `yaml:"paths"`
	Conventions DefaultConventions `yaml:"conventions"`
}

// DefaultPaths contains default path configuration.
type DefaultPaths struct {
	SpecsRoot string   `yaml:"specs_root"` // Default specs root
	Templates string   `yaml:"templates"`  // Default templates directory
	Out       OutPaths `yaml:"out"`        // Output directory structure
}

// OutPaths contains output directory structure.
type OutPaths struct {
	Root     string `yaml:"root"`     // Root output directory
	Build    string `yaml:"build"`    // Build output directory
	Test     string `yaml:"test"`     // Test output directory
	Logs     string `yaml:"logs"`     // Logs output directory
	Security string `yaml:"security"` // Security scan output directory
	Tools    string `yaml:"tools"`    // Tools output directory
}

// DefaultConventions contains default filename conventions.
type DefaultConventions struct {
	GodogTest   string `yaml:"godog_test"`   // Godog test file name
	PackageJSON string `yaml:"package_json"` // Node.js package file name
	Changelog   string `yaml:"changelog"`    // Changelog file name
}

// ModuleVersioning contains module versioning configuration.
type ModuleVersioning struct {
	Scheme    string `yaml:"scheme"`              // SemVer | CalVer
	Current   string `yaml:"current,omitempty"`   // Current version (optional)
	Changelog string `yaml:"changelog,omitempty"` // Path to changelog (defaults to release/<moniker>/CHANGELOG.md)
}

// BaseContract represents the base structure for module contracts.
type BaseContract struct {
	Moniker       string            `yaml:"moniker"`
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	DependsOn     []string          `yaml:"depends_on"`
	Versioning    *ModuleVersioning `yaml:"versioning,omitempty"`
	EvidenceBooks []string          `yaml:"evidence_books,omitempty"` // Evidence book names
	ReleaseBundle *ReleaseBundle    `yaml:"release_bundle,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	Components    ModuleComponents  `yaml:"components"`         // Component types mapped to their roots
	Linting       *ModuleLinting    `yaml:"linting,omitempty"`  // Linting configuration overrides
}

// ModuleLinting configures linting behavior for a module.
type ModuleLinting struct {
	// Enabled lists specific lint providers to use (empty = use all applicable from lint-providers.yml)
	Enabled []string `yaml:"enabled,omitempty"`

	// Disabled lists lint providers to skip, or "all" to disable linting entirely
	Disabled []string `yaml:"disabled,omitempty"`
}

// ReleaseBundle configures how the release module creates GitHub releases.
type ReleaseBundle struct {
	TitleFormat string                  `yaml:"title_format" json:"title_format"` // Title template, e.g., "{r2r} ({r2r_version}) + {eac} ({eac_version})"
	Headline    map[string]string       `yaml:"headline" json:"headline"`         // Map of label -> moniker for title modules
	Categories  []ReleaseBundleCategory `yaml:"categories" json:"categories"`     // Grouped modules for release notes
}

// ReleaseBundleCategory groups modules in release notes.
type ReleaseBundleCategory struct {
	Name        string   `yaml:"name" json:"name"`               // Category name, e.g., "Core Tools"
	Description string   `yaml:"description" json:"description"` // Category description
	Modules     []string `yaml:"modules" json:"modules"`         // Module monikers in this category
}

// ModuleBuild contains per-module build configuration
// This allows modules to define their own artifacts instead of relying on type-level defaults.
type ModuleBuild struct {
	Handler   string           `yaml:"handler,omitempty"`   // Explicit build handler override (e.g., "mkdocs", "docker")
	Artifacts []ModuleArtifact `yaml:"artifacts,omitempty"` // Artifacts to produce
	Options   *BuildOptions    `yaml:"options,omitempty"`   // Build behavior options
}

// ModuleArtifact defines an artifact to be produced by a module build.
type ModuleArtifact struct {
	ID          string `yaml:"id"`                    // Unique artifact identifier
	Type        string `yaml:"type"`                  // executable, file, directory, test
	Pattern     string `yaml:"pattern"`               // Output path pattern with variables: {moniker}, {ext}
	Compression string `yaml:"compression,omitempty"` // none, strip, upx
	DeriveFrom  string `yaml:"derive_from,omitempty"` // Source artifact to derive from (for compressed variants)
}

// BuildOptions contains optional build behavior flags.
type BuildOptions struct {
	// Reserved for future build options
}

// NOTE: Files, Workflows, Flags, RepoPatterns structs removed.
// File ownership is now determined by components.

// Getter methods for BaseContract

func (b *BaseContract) GetMoniker() string {
	return b.Moniker
}

func (b *BaseContract) GetName() string {
	return b.Name
}

func (b *BaseContract) GetDescription() string {
	return b.Description
}

// GetComponentRoot returns the root for a specific component type.
func (b *BaseContract) GetComponentRoot(compType string) string {
	return b.Components.GetComponentRoot(compType)
}

// HasBuildArtifacts returns true if any component has build artifacts defined.
func (b *BaseContract) HasBuildArtifacts() bool {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil && len(comp.Build.Artifacts) > 0 {
			return true
		}
	}
	return false
}

// GetBuildArtifacts returns all build artifacts from all components.
func (b *BaseContract) GetBuildArtifacts() []ModuleArtifact {
	var result []ModuleArtifact
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				result = append(result, ModuleArtifact{
					ID:          a.ID,
					Type:        a.Type,
					Pattern:     a.Pattern,
					Compression: a.Compression,
					DeriveFrom:  a.DeriveFrom,
				})
			}
		}
	}
	return result
}

// GetComponentBuildArtifacts returns build artifacts for a specific component.
func (b *BaseContract) GetComponentBuildArtifacts(compName string) []ComponentArtifact {
	if comp, ok := b.Components[compName]; ok && comp != nil && comp.Build != nil {
		return comp.Build.Artifacts
	}
	return nil
}

// HasExecutableArtifacts returns true if any component has executable artifacts.
func (b *BaseContract) HasExecutableArtifacts() bool {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				if a.Type == "executable" {
					return true
				}
			}
		}
	}
	return false
}

// HasTestArtifacts returns true if any component has test artifacts.
func (b *BaseContract) HasTestArtifacts() bool {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				if a.Type == "test" {
					return true
				}
			}
		}
	}
	return false
}

// GetBuildHandler returns the build handler from the first component that has one.
// This is used for module-level handler override detection.
func (b *BaseContract) GetBuildHandler() string {
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil && comp.Build.Handler != "" {
			return comp.Build.Handler
		}
	}
	return ""
}

// GetComponentBuildHandler returns the build handler for a specific component.
func (b *BaseContract) GetComponentBuildHandler(compName string) string {
	if comp, ok := b.Components[compName]; ok && comp != nil && comp.Build != nil {
		return comp.Build.Handler
	}
	return ""
}

// GetArtifactsByType returns all artifacts of the specified type from all components.
func (b *BaseContract) GetArtifactsByType(artifactType string) []ModuleArtifact {
	var result []ModuleArtifact
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				if a.Type == artifactType {
					result = append(result, ModuleArtifact{
						ID:          a.ID,
						Type:        a.Type,
						Pattern:     a.Pattern,
						Compression: a.Compression,
						DeriveFrom:  a.DeriveFrom,
					})
				}
			}
		}
	}
	return result
}

// HasBooks returns true if the module has any components of type 'book'.
func (b *BaseContract) HasBooks() bool {
	return b.Components.HasComponentType("book")
}

// GetBooks returns the list of book component names (components with type 'book').
// Names are sorted for consistent ordering.
func (b *BaseContract) GetBooks() []string {
	books := b.Components.GetComponentsByType("book")
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
func (b *BaseContract) GetChangelog() string {
	if b.Versioning != nil && b.Versioning.Changelog != "" {
		return b.Versioning.Changelog
	}
	return "release/" + b.Moniker + "/CHANGELOG.md"
}
