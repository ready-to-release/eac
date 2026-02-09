package domain

import (
	"sort"

	"github.com/ready-to-release/eac/go/core/config"
)

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

// NOTE: ModuleVersioning, ReleaseBundle, and component types (ModuleComponents, ComponentEntry, etc.)
// are defined in config package as the single source of truth.

// BaseContract represents the base structure for module domain.
type BaseContract struct {
	Moniker       string                  `yaml:"moniker"`
	Name          string                  `yaml:"name"`
	Description   string                  `yaml:"description"`
	ModuleGroup   string                  `yaml:"module_group,omitempty"` // Group name for depends_on expansion
	DependsOn     []string                `yaml:"depends_on"`
	Versioning    *config.ModuleVersioning `yaml:"versioning,omitempty"`
	EvidenceBooks []string                `yaml:"evidence_books,omitempty"` // Evidence book names
	ReleaseBundle *config.ReleaseBundle   `yaml:"release_bundle,omitempty"`
	Metadata      map[string]string       `yaml:"metadata,omitempty"`
	Components     config.ModuleComponents `yaml:"components"`        // Component types mapped to their roots
	ComponentOrder []string               `yaml:"-"`                 // YAML declaration order of component keys
	Linting        *config.ModuleLinting  `yaml:"linting,omitempty"` // Linting configuration overrides
}

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

func (b *BaseContract) GetModuleGroup() string {
	return b.ModuleGroup
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
func (b *BaseContract) GetBuildArtifacts() []config.ModuleArtifact {
	var result []config.ModuleArtifact
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			result = append(result, comp.Build.Artifacts...)
		}
	}
	return result
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

// GetArtifactsByType returns all artifacts of the specified type from all components.
func (b *BaseContract) GetArtifactsByType(artifactType string) []config.ModuleArtifact {
	var result []config.ModuleArtifact
	for _, comp := range b.Components {
		if comp != nil && comp.Build != nil {
			for _, a := range comp.Build.Artifacts {
				if a.Type == artifactType {
					result = append(result, a)
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

// GetChangelog returns the changelog path.
// Only SemVer modules default to release/<moniker>/CHANGELOG.md.
// CalVer (auto-managed) and Implicit (non-releasable) return empty unless explicitly configured.
func (b *BaseContract) GetChangelog() string {
	if b.Versioning != nil && b.Versioning.Changelog != "" {
		return b.Versioning.Changelog
	}
	// Only SemVer modules get the default release/ changelog
	if b.Versioning != nil && b.Versioning.Scheme == "SemVer" {
		return "release/" + b.Moniker + "/CHANGELOG.md"
	}
	return ""
}

// GetDependsOn returns the list of module dependencies.
// Implements interfaces.ModuleContractPort.
func (b *BaseContract) GetDependsOn() []string {
	return b.DependsOn
}

// GetVersioningScheme returns the versioning scheme (SemVer, CalVer, Implicit).
// Implements interfaces.ModuleContractPort.
func (b *BaseContract) GetVersioningScheme() string {
	if b.Versioning == nil {
		return ""
	}
	return b.Versioning.Scheme
}

// GetReleaseType returns the release type (published, internal, bundle, none).
// Implements interfaces.ModuleContractPort.
func (b *BaseContract) GetReleaseType() string {
	if b.Versioning == nil {
		return ""
	}
	return b.Versioning.ReleaseType
}

// HasVersioning returns true if versioning is configured for this module.
// Implements interfaces.ModuleContractPort.
func (b *BaseContract) HasVersioning() bool {
	return b.Versioning != nil && b.Versioning.Scheme != ""
}

// GetMetadata returns the module metadata as interface{} map.
// Implements interfaces.ModuleContractPort.
func (b *BaseContract) GetMetadata() map[string]interface{} {
	if b.Metadata == nil {
		return nil
	}
	result := make(map[string]interface{}, len(b.Metadata))
	for k, v := range b.Metadata {
		result[k] = v
	}
	return result
}

// GetAllComponentDeps returns all inter-module component_deps across all components.
func (b *BaseContract) GetAllComponentDeps() []string {
	return b.Components.GetAllComponentDeps()
}

// GetComponentGroup returns the component group for a named component.
func (b *BaseContract) GetComponentGroup(compName string) string {
	return b.Components.GetComponentGroup(compName)
}

// GetComponentAmp returns the weight amplifier for a component and operation.
// Returns 1.0 if not configured (no amplification).
// Implements interfaces.ModuleContractPort.
func (b *BaseContract) GetComponentAmp(componentName, operation string) float64 {
	comp, ok := b.Components[componentName]
	if !ok || comp == nil {
		return 1.0
	}
	return comp.GetAmpForOperation(operation)
}
