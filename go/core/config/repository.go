// repository.go provides unified repository configuration loading.
// This is the single source of truth for all repository configuration including modules.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"
)

// BDDComponentNamesFunc is set by the test runner registry to provide BDD component
// names without creating an import cycle (config -> testing -> config).
// When nil, falls back to checking all BDD-runner component types from ComponentKinds.
var BDDComponentNamesFunc func() []string

// RepositoryFileName is the config file for all repository settings.
const RepositoryFileName = "repository.yml"

// RepositoryConfig holds all repository configuration including modules.
// This is the unified config loaded from .eac/repository.yml.
type RepositoryConfig struct {
	// Repository settings
	Repository RepositorySettings `yaml:"repository"`

	// Path configuration
	Paths PathsConfig `yaml:"paths"`

	// Filename conventions
	Conventions ConventionsConfig `yaml:"conventions"`

	// Module definitions (previously in separate repository.yml)
	Modules []Module `yaml:"modules"`

	// Container registry configurations for cleanup policies
	Registries RegistriesConfig `yaml:"registries,omitempty"`

	// DisplayOrder is the precomputed display ordering for modules and components.
	// Populated during config loading after module groups and component types are resolved.
	DisplayOrder *DisplayOrder `yaml:"-"`

	// monikerIndex maps moniker to index in the Modules slice for O(1) lookup.
	// Built once after modules are fully loaded via buildMonikerIndex.
	monikerIndex map[string]int `yaml:"-"`

	// baselineModules records modules that declared depends_on: [root] before stripping.
	// Used internally by computeDisplayOrder to assign depth -1.
	baselineModules map[string]bool `yaml:"-"`
}

// RepositorySettings holds repository-level configuration.
type RepositorySettings struct {
	Type              string              `yaml:"type"`                  // mono, poly, adjunct
	TrunkBranch       string              `yaml:"trunk_branch"`          // main branch name
	MaxBranchAgeDays  int                 `yaml:"max_branch_age_days"`   // max age for feature branches
	Schemes           []string            `yaml:"schemes"`               // valid versioning schemes for releasable modules (Implicit always available)
	PR                PRConfig            `yaml:"pr"`                    // PR workflow config
	Versioning        VersioningConfig    `yaml:"versioning"`            // versioning constraints
	Parallelism       ParallelismConfig   `yaml:"parallelism"`           // parallelism limits
	Remote            RemoteConfig        `yaml:"remote"`                // Remote VCS repository configuration
	OptimizeGitLsInCI bool                `yaml:"optimize_git_ls_in_ci"` // Use GitHub API for file listing in CI (faster than git ls-files)
	GhostTracking     GhostTrackingConfig `yaml:"ghost-tracking"`        // Ghost tracking configuration
}

// GhostTrackingConfig holds configuration for ghost (dark launch) code tracking.
type GhostTrackingConfig struct {
	// Alias is the prefix used to identify ghosts (default: "ghost")
	// Results in patterns: alias-*, alias.*, alias
	Alias string `yaml:"ghost-alias"`
}

// ParallelismConfig holds parallelism limits for build and test operations.
type ParallelismConfig struct {
	CI     int `yaml:"ci"`     // Max parallel workers in CI (default: 8)
	Devbox int `yaml:"devbox"` // Max parallel workers locally (default: 16)
}

// PRConfig holds pull request workflow configuration.
type PRConfig struct {
	DeleteBranchOnMerge bool   `yaml:"delete_branch_on_merge"`
	MergeStrategy       string `yaml:"merge_strategy"` // squash, merge, rebase
}

// VersioningConfig holds repository-wide versioning constraints.
type VersioningConfig struct {
	Constraint string `yaml:"constraint"` // unrestricted, patch-only, calver-only
}

// Versioning constraint constants.
const (
	VersioningUnrestricted = "unrestricted"
	VersioningPatchOnly    = "patch-only"
	VersioningCalverOnly   = "calver-only"
)

// IsPatchOnly returns true if versioning is constrained to patch-only.
func (v VersioningConfig) IsPatchOnly() bool {
	return v.Constraint == VersioningPatchOnly
}

// IsCalverOnly returns true if versioning is forced to calver.
func (v VersioningConfig) IsCalverOnly() bool {
	return v.Constraint == VersioningCalverOnly
}

// IsUnrestricted returns true if versioning is unrestricted.
func (v VersioningConfig) IsUnrestricted() bool {
	return v.Constraint == VersioningUnrestricted || v.Constraint == ""
}

// PathsConfig defines repository-specific directory structures.
type PathsConfig struct {
	SpecsRoot      string    `yaml:"specs_root"`
	ContainersRoot string    `yaml:"containers_root"`
	Templates      string    `yaml:"templates"`
	Out            OutConfig `yaml:"out"`
}

// OutConfig defines output directory structure.
type OutConfig struct {
	Root  string `yaml:"root"`
	Build string `yaml:"build"`
	Test  string `yaml:"test"`
	Logs  string `yaml:"logs"`
	Scan  string `yaml:"scan"`
	Tools string `yaml:"tools"` // CI tools like the commands binary (not build outputs)
}

// ConventionsConfig defines conventional filenames.
type ConventionsConfig struct {
	GodogTest              string `yaml:"godog_test"`
	PackageJSON            string `yaml:"package_json"`
	Changelog              string `yaml:"changelog"`
	BuildLog               string `yaml:"build_log"`
	BuildTiming            string `yaml:"build_timing"`
	TestTiming             string `yaml:"test_timing"`
	Specification          string `yaml:"specification"`
	RiskCatalog            string `yaml:"risk_catalog"`
	RiskControlsDir        string `yaml:"risk_controls_dir"`
	RiskReportsCategory    string `yaml:"risk_reports_category"`
	RiskAssessmentTemplate string `yaml:"risk_assessment_template"`
	TestReportsCategory    string `yaml:"test_reports_category"`
	TestResultsTemplate    string `yaml:"test_results_template"`
	TemplateSpecsDir       string `yaml:"template_specs_dir"`
	TemplateReportsDir     string `yaml:"template_reports_dir"`
	TemplateRiskCatalogDir string `yaml:"template_risk_catalog_dir"`
	DesignDir    string `yaml:"design_dir"`
	WorkspaceDSL string `yaml:"workspace_dsl"`
}

// loadRepositoryConfigUnmerged loads repository configuration from user's YAML file only.
// WARNING: This returns UNMERGED config without defaults. Do NOT use directly.
// Use config.Load().Repository instead to get properly merged config with defaults.
// This is only exported for use by the config package's merge logic.
func loadRepositoryConfigUnmerged(repoRoot string) (*RepositoryConfig, error) {
	configPath := filepath.Join(paths.EACConfigPath(repoRoot), RepositoryFileName)
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &RepositoryConfig{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg RepositoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GetModule returns a module by moniker.
// Uses a pre-built index for O(1) lookup when available.
func (c *RepositoryConfig) GetModule(moniker string) (*Module, bool) {
	if c.monikerIndex != nil {
		if idx, ok := c.monikerIndex[moniker]; ok {
			return &c.Modules[idx], true
		}
		return nil, false
	}
	// Fallback to linear scan if index not yet built (e.g., during loading)
	for i := range c.Modules {
		if c.Modules[i].Moniker == moniker {
			return &c.Modules[i], true
		}
	}
	return nil, false
}

// GetByMoniker returns a module by moniker, or nil if not found.
func (c *RepositoryConfig) GetByMoniker(moniker string) *Module {
	m, ok := c.GetModule(moniker)
	if !ok {
		return nil
	}
	return m
}

// AllMonikers returns a list of all module monikers.
func (c *RepositoryConfig) AllMonikers() []string {
	monikers := make([]string, len(c.Modules))
	for i, m := range c.Modules {
		monikers[i] = m.Moniker
	}
	return monikers
}

// DesignComponents returns a map of module moniker → design facet component name.
// Used by docprep to locate structurizr build output directories.
func (c *RepositoryConfig) DesignComponents() map[string]string {
	if c == nil {
		return nil
	}
	result := make(map[string]string)
	for _, m := range c.Modules {
		for name, entry := range m.Components {
			if entry != nil && entry.FacetName == "design" {
				result[m.Moniker] = name
				break
			}
		}
	}
	return result
}

// DiagramComponents returns a map of diagram type → component name for a module.
// Used by docprep to locate diagram builder output directories.
// Scans the module's components for known diagram types (docs-mermaid, docs-drawio,
// docs-plantuml, docs-markdown-commands) and returns their component names (map keys).
func (c *RepositoryConfig) DiagramComponents(moniker string) map[string]string {
	if c == nil {
		return nil
	}
	result := make(map[string]string)
	mod, ok := c.GetModule(moniker)
	if !ok {
		return result
	}
	for name, entry := range mod.Components {
		compType := name
		if entry != nil && entry.Type != "" {
			compType = entry.Type
		}
		switch compType {
		case "docs-mermaid":
			result["mermaid"] = name
		case "docs-drawio":
			result["drawio"] = name
		case "docs-plantuml":
			result["plantuml"] = name
		case "docs-markdown-commands":
			result["markdown-commands"] = name
		}
	}
	return result
}

// buildMonikerIndex builds the moniker-to-index map for O(1) lookup.
// Must be called after all modules are finalized (after parameter substitution,
// container discovery, and group expansion).
func (c *RepositoryConfig) buildMonikerIndex() {
	c.monikerIndex = make(map[string]int, len(c.Modules))
	for i, m := range c.Modules {
		c.monikerIndex[m.Moniker] = i
	}
}

// ExpandModuleParams substitutes parameter placeholders and expands artifact matrices.
// This should be called after loading and merging configs but before ApplyComponentDefaults.
// All components must be explicitly declared in repository.yml; no auto-discovery is performed.
// The blueprints parameter provides component kinds and artifact matrices.
func (c *RepositoryConfig) ExpandModuleParams(repoRoot string, blueprints *BlueprintsConfig) error {
	var componentKinds map[string]*ComponentType
	if blueprints != nil {
		componentKinds = blueprints.ComponentKinds
	}

	// Pre-resolve default roots from component kinds
	if componentKinds != nil {
		for i := range c.Modules {
			preResolveDefaultRoots(&c.Modules[i], componentKinds)
		}
	}

	// Get owner from repository config
	owner := c.Repository.Remote.Owner

	// Substitute parameters and expand artifact matrices for each module
	for i := range c.Modules {
		mod := &c.Modules[i]

		// Build per-module variables and substitute placeholders
		discoveryVars := buildDiscoveryVars(mod, c)
		discoveryVars["owner"] = owner
		SubstituteModuleParams(mod, discoveryVars)

		// Expand artifact matrix reference into Go component artifacts
		expandArtifactMatrixForModule(mod, blueprints)
	}

	return nil
}

// preResolveDefaultRoots resolves empty component roots from component kind
// defaults. This is a lightweight pre-pass that runs before discovery so that
// derive_from_type rules can find components whose roots come from kind defaults
// (e.g., dockerfile with default_root "containers/{moniker}").
func preResolveDefaultRoots(mod *Module, kinds map[string]*ComponentType) {
	if kinds == nil {
		return
	}
	for name, entry := range mod.Components {
		if entry == nil {
			// Initialize nil entries (YAML "dockerfile:" with no value)
			entry = &ComponentEntry{}
			mod.Components[name] = entry
		}
		if entry.Root != "" {
			continue
		}
		compType := name
		if entry.Type != "" {
			compType = entry.Type
		}
		if ct, ok := kinds[compType]; ok && ct.DefaultRoot != "" {
			entry.Root = ct.GetRoot(mod.Moniker, "")
		}
	}
}


// ToPathConfig converts RepositoryConfig path/convention values to a
// paths.PathConfig for use with config-aware path functions.
func (c *RepositoryConfig) ToPathConfig() paths.PathConfig {
	return paths.PathConfig{
		SpecsRoot:              c.Paths.SpecsRoot,
		Templates:              c.Paths.Templates,
		OutBuild:               c.Paths.Out.Build,
		OutTest:                c.Paths.Out.Test,
		OutLogs:                c.Paths.Out.Logs,
		OutScan:                c.Paths.Out.Scan,
		OutTools:               c.Paths.Out.Tools,
		BuildLog:               c.Conventions.BuildLog,
		BuildTiming:            c.Conventions.BuildTiming,
		TestTiming:             c.Conventions.TestTiming,
		Specification:          c.Conventions.Specification,
		RiskCatalog:            c.Conventions.RiskCatalog,
		RiskControlsDir:        c.Conventions.RiskControlsDir,
		TemplateSpecsDir:       c.Conventions.TemplateSpecsDir,
		TemplateReportsDir:     c.Conventions.TemplateReportsDir,
		TemplateRiskCatalogDir: c.Conventions.TemplateRiskCatalogDir,
		GodogTest:              c.Conventions.GodogTest,
	}
}

// TestImplPath returns the full path to a module's BDD test implementation.
// Checks for explicit BDD runner components by type (queried from the adapter registry).
// Auto-created BDD runners (from expandBDDRunners) are skipped since their root
// points to the specs directory, not the Go test implementation directory.
// Falls back to the primary component root when no explicit runner exists,
// allowing FindTestRoot to search within the source tree.
func (c *RepositoryConfig) TestImplPath(moniker string) string {
	module, found := c.GetModule(moniker)
	if !found {
		return ""
	}

	// Check for explicit (user-declared) BDD runner components
	for _, compType := range getBDDComponentNames() {
		if _, comp := module.Components.GetFirstByType(compType); comp != nil && comp.Root != "" {
			if !comp.AutoBDDRunner {
				return comp.Root
			}
		}
	}

	// Fallback: return primary component root (FindTestRoot searches within).
	if _, comp := module.Components.FindPrimaryComponent(); comp != nil && comp.Root != "" {
		return comp.Root
	}

	return ""
}

// TestImplPathForQualifier returns the BDD test runner root for a specific
// component qualifier within a module. Used when spec directories use the
// <module>_<qualifier> naming convention.
// Matches BDD runner components by stripping type prefix from component name.
func (c *RepositoryConfig) TestImplPathForQualifier(moniker, qualifier string) string {
	module, found := c.GetModule(moniker)
	if !found {
		return ""
	}
	bddTypes := make(map[string]bool)
	for _, t := range getBDDComponentNames() {
		bddTypes[t] = true
	}
	for name, comp := range module.Components {
		if comp == nil || comp.Root == "" {
			continue
		}
		compType := comp.Type
		if compType == "" {
			compType = name
		}
		if !bddTypes[compType] {
			continue
		}
		// Match by stripping type prefix: "godog-work" → "work"
		compQualifier := strings.TrimPrefix(name, compType+"-")
		if compQualifier == qualifier {
			return comp.Root
		}
	}
	return ""
}

// getBDDComponentNames returns BDD component names from the adapter registry
// or falls back to querying ComponentKinds for types with bdd_runner set.
func getBDDComponentNames() []string {
	if BDDComponentNamesFunc != nil {
		if names := BDDComponentNamesFunc(); len(names) > 0 {
			return names
		}
	}
	// Fallback: query ComponentKinds for BDD runner types
	cfg := Global()
	if cfg != nil && cfg.ComponentKinds != nil {
		return cfg.ComponentKinds.GetBDDRunnerTypes()
	}
	return nil
}

// EffectiveParallelism returns the maximum number of parallel workers
// based on the runtime environment. Pass isCI=true for CI environments,
// isCI=false for local development (devbox).
// Returns 0 if not configured, signaling dynamic calculation (CPU×RAM based).
func (c *RepositoryConfig) EffectiveParallelism(isCI bool) int {
	if isCI {
		if c.Repository.Parallelism.CI > 0 {
			return c.Repository.Parallelism.CI
		}
		return 0 // Dynamic - let orchestrator calculate from CPU×RAM
	}
	// Devbox/local environment
	if c.Repository.Parallelism.Devbox > 0 {
		return c.Repository.Parallelism.Devbox
	}
	return 0 // Dynamic - let orchestrator calculate from CPU×RAM
}
