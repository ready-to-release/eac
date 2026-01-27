// repository.go provides unified repository configuration loading.
// This is the single source of truth for all repository configuration including modules.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

// sanitizeMonikerForPath converts a moniker to a filesystem-safe path component.
// Replaces : with _ (Windows doesn't allow : in paths).
func sanitizeMonikerForPath(moniker string) string {
	return strings.ReplaceAll(moniker, ":", "_")
}

// RepositoryFileName is the config file for all repository settings.
const RepositoryFileName = "repository.yml"

// RepositoryConfig holds all repository configuration including modules.
// This is the unified config loaded from .r2r/eac/repository.yml.
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

	// Container tool configurations with version pinning
	Containers ContainersConfig `yaml:"containers,omitempty"`
}

// RepositorySettings holds repository-level configuration.
type RepositorySettings struct {
	Type              string            `yaml:"type"`                  // mono, poly, adjunct
	TrunkBranch       string            `yaml:"trunk_branch"`          // main branch name
	MaxBranchAgeDays  int               `yaml:"max_branch_age_days"`   // max age for feature branches
	Schemes           []string          `yaml:"schemes"`               // valid versioning schemes for releasable modules (Implicit always available)
	PR                PRConfig          `yaml:"pr"`                    // PR workflow config
	Versioning        VersioningConfig  `yaml:"versioning"`            // versioning constraints
	Parallelism       ParallelismConfig `yaml:"parallelism"`           // parallelism limits
	OptimizeGitLsInCI bool              `yaml:"optimize_git_ls_in_ci"` // Use GitHub API for file listing in CI (faster than git ls-files)
}

// ParallelismConfig holds parallelism limits for build and test operations.
type ParallelismConfig struct {
	CI     int `yaml:"ci"`     // Max parallel workers in CI (default: 2)
	Devbox int `yaml:"devbox"` // Max parallel workers locally (default: 4)
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
	SpecsRoot string    `yaml:"specs_root"`
	Templates string    `yaml:"templates"`
	Out       OutConfig `yaml:"out"`
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

// TestImplPath returns the full path to a module's test implementation.
// Returns empty string if module not found or has no test-impl component.
// Modules must explicitly define a test-impl component - there is no fallback.
func (c *RepositoryConfig) TestImplPath(moniker string) string {
	module, found := c.GetModule(moniker)
	if !found {
		return ""
	}

	// Require explicit test-impl component - no fallback
	if comp, ok := module.Components["test-impl"]; ok && comp != nil && comp.Root != "" {
		return comp.Root
	}

	return ""
}

// SpecsPath returns the full path to a module's specs directory.
func (c *RepositoryConfig) SpecsPath(moniker string) string {
	return c.Paths.SpecsRoot + "/" + moniker
}

// BuildOutputPath returns the relative path to a module's build output.
// For absolute paths, use paths.BuildOutputPath(workspaceRoot, moniker) instead.
func (c *RepositoryConfig) BuildOutputPath(moniker string) string {
	return c.Paths.Out.Build + "/" + moniker
}

// BuildOutputPathAbs returns the absolute path to a module's build output.
func (c *RepositoryConfig) BuildOutputPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Build, moniker)
}

// TemplatePath returns the relative path to a template file within the templates directory.
// Example: TemplatePath("reports", "release", "release-notes-template.md") returns "templates/reports/release/release-notes-template.md".
func (c *RepositoryConfig) TemplatePath(pathComponents ...string) string {
	parts := append([]string{c.Paths.Templates}, pathComponents...)
	return filepath.Join(parts...)
}

// TemplatePathAbs returns the absolute path to a template file.
func (c *RepositoryConfig) TemplatePathAbs(workspaceRoot string, pathComponents ...string) string {
	return filepath.Join(workspaceRoot, c.TemplatePath(pathComponents...))
}

// TestModuleDir returns the path to a module's test output directory.
// New structure: out/test/<module>
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestModuleDir(moniker string) string {
	return c.Paths.Out.Test + "/" + sanitizeMonikerForPath(moniker)
}

// TestModuleDirAbs returns the absolute path to a module's test output directory.
// New structure: out/test/<module>
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestModuleDirAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Test, sanitizeMonikerForPath(moniker))
}

// TestManifestPath returns the path to a module's test manifest file.
// Path: out/test/<module>/test.manifest.json
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestManifestPath(moniker string) string {
	return c.Paths.Out.Test + "/" + sanitizeMonikerForPath(moniker) + "/test.manifest.json"
}

// TestManifestPathAbs returns the absolute path to a module's test manifest file.
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestManifestPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Test, sanitizeMonikerForPath(moniker), "test.manifest.json")
}

// TestPackageDir returns the path to a package's test output within a module.
// Path: out/test/<module>/packages/<package>
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestPackageDir(moniker, pkgPath string) string {
	return c.Paths.Out.Test + "/" + sanitizeMonikerForPath(moniker) + "/packages/" + pkgPath
}

// TestPackageDirAbs returns the absolute path to a package's test output within a module.
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestPackageDirAbs(workspaceRoot, moniker, pkgPath string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Test, sanitizeMonikerForPath(moniker), "packages", pkgPath)
}

// LogsPath returns the path to logs for a command.
func (c *RepositoryConfig) LogsPath(command string) string {
	return c.Paths.Out.Logs + "/" + command
}

// ToolsPath returns the path to the tools directory.
func (c *RepositoryConfig) ToolsPath() string {
	return c.Paths.Out.Tools
}

// IsGodogTestFile checks if a filename is the godog test file.
func (c *RepositoryConfig) IsGodogTestFile(filename string) bool {
	return filename == c.Conventions.GodogTest
}

// GetPathVariables returns a map of path variables for template substitution.
func (c *RepositoryConfig) GetPathVariables() map[string]string {
	return map[string]string{
		"specs_root": c.Paths.SpecsRoot,
		"templates":  c.Paths.Templates,
		"out_root":   c.Paths.Out.Root,
		"out_build":  c.Paths.Out.Build,
		"out_test":   c.Paths.Out.Test,
		"out_logs":   c.Paths.Out.Logs,
		"out_scan":   c.Paths.Out.Scan,
		"out_tools":  c.Paths.Out.Tools,
	}
}

// ============================================================================
// Additional Path Methods (Contract-Aware)
// ============================================================================

// BuildOutputDir returns the root build output directory.
func (c *RepositoryConfig) BuildOutputDir() string {
	return c.Paths.Out.Build
}

// BuildOutputDirAbs returns the absolute root build output directory.
func (c *RepositoryConfig) BuildOutputDirAbs(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Build)
}

// BuildLogPath returns the path to a module's build.log file.
func (c *RepositoryConfig) BuildLogPath(moniker string) string {
	return c.Paths.Out.Build + "/" + moniker + "/" + c.Conventions.BuildLog
}

// BuildLogPathAbs returns the absolute path to a module's build.log file.
func (c *RepositoryConfig) BuildLogPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Build, moniker, c.Conventions.BuildLog)
}

// BuildTimingPath returns the path to a module's build-timing.txt file.
func (c *RepositoryConfig) BuildTimingPath(moniker string) string {
	return c.Paths.Out.Build + "/" + moniker + "/" + c.Conventions.BuildTiming
}

// BuildTimingPathAbs returns the absolute path to a module's build-timing.txt file.
func (c *RepositoryConfig) BuildTimingPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Build, moniker, c.Conventions.BuildTiming)
}

// TestOutputDir returns the root test output directory.
func (c *RepositoryConfig) TestOutputDir() string {
	return c.Paths.Out.Test
}

// TestOutputDirAbs returns the absolute root test output directory.
func (c *RepositoryConfig) TestOutputDirAbs(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Test)
}

// TestModuleTimingPath returns the path to a module's test timing file
// Path: out/test/<module>/test-timing.txt
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestModuleTimingPath(moniker string) string {
	return c.Paths.Out.Test + "/" + sanitizeMonikerForPath(moniker) + "/" + c.Conventions.TestTiming
}

// TestModuleTimingPathAbs returns the absolute path to a module's test timing file
// Sanitizes moniker for filesystem safety (replaces : with _).
func (c *RepositoryConfig) TestModuleTimingPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Test, sanitizeMonikerForPath(moniker), c.Conventions.TestTiming)
}

// ScanOutputDir returns the root scan output directory.
func (c *RepositoryConfig) ScanOutputDir() string {
	return c.Paths.Out.Scan
}

// ScanOutputDirAbs returns the absolute root scan output directory.
func (c *RepositoryConfig) ScanOutputDirAbs(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Scan)
}

// ScanModuleOutputPath returns the path to a module's scan output directory.
func (c *RepositoryConfig) ScanModuleOutputPath(moduleName string) string {
	return c.Paths.Out.Scan + "/" + moduleName
}

// ScanModuleOutputPathAbs returns the absolute path to a module's scan output directory.
func (c *RepositoryConfig) ScanModuleOutputPathAbs(workspaceRoot, moduleName string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Scan, moduleName)
}

// ScanManifestPath returns the path to a module's scan manifest file.
// Path: out/scan/<module>/scan.manifest.json.
func (c *RepositoryConfig) ScanManifestPath(moniker string) string {
	return c.Paths.Out.Scan + "/" + moniker + "/scan.manifest.json"
}

// ScanManifestPathAbs returns the absolute path to a module's scan manifest file.
func (c *RepositoryConfig) ScanManifestPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.Paths.Out.Scan, moniker, "scan.manifest.json")
}

// LogsPathAbs returns the absolute path to logs for a command with optional path segments
// Delegates to paths.CommandLogsPath for consistency
// Examples:
//
//	LogsPathAbs(root, "design") → out/design/
//	LogsPathAbs(root, "build", "eac-core") → out/build/eac-core/
//	LogsPathAbs(root, "templates", "apply") → out/templates/apply/
func (c *RepositoryConfig) LogsPathAbs(workspaceRoot, command string, pathSegments ...string) string {
	return paths.CommandLogsPath(workspaceRoot, command, pathSegments...)
}

// SpecsPathAbs returns the absolute path to a module's specs directory.
func (c *RepositoryConfig) SpecsPathAbs(workspaceRoot, moniker string) string {
	return filepath.Join(workspaceRoot, c.SpecsPath(moniker))
}

// RiskControlsPath returns the path to the risk controls directory.
func (c *RepositoryConfig) RiskControlsPath() string {
	// Risk controls directory is under specs root (not configurable separately)
	return c.Paths.SpecsRoot + "/" + c.Conventions.RiskControlsDir
}

// RiskControlsPathAbs returns the absolute path to the risk controls directory.
func (c *RepositoryConfig) RiskControlsPathAbs(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, c.RiskControlsPath())
}

// RiskCatalogPath returns the path to the risk catalog file.
func (c *RepositoryConfig) RiskCatalogPath() string {
	return c.TemplatePath(c.Conventions.TemplateSpecsDir, c.Conventions.TemplateRiskCatalogDir, c.Conventions.RiskCatalog)
}

// RiskCatalogPathAbs returns the absolute path to the risk catalog file.
func (c *RepositoryConfig) RiskCatalogPathAbs(workspaceRoot string) string {
	return c.TemplatePathAbs(workspaceRoot, c.Conventions.TemplateSpecsDir, c.Conventions.TemplateRiskCatalogDir, c.Conventions.RiskCatalog)
}

// TemplateSpecsPath returns the path to specs templates subdirectory.
func (c *RepositoryConfig) TemplateSpecsPath(subpaths ...string) string {
	parts := append([]string{c.Conventions.TemplateSpecsDir}, subpaths...)
	return c.TemplatePath(parts...)
}

// TemplateSpecsPathAbs returns the absolute path to specs templates subdirectory.
func (c *RepositoryConfig) TemplateSpecsPathAbs(workspaceRoot string, subpaths ...string) string {
	parts := append([]string{c.Conventions.TemplateSpecsDir}, subpaths...)
	return c.TemplatePathAbs(workspaceRoot, parts...)
}

// TemplateReportsPath returns the path to reports templates subdirectory.
func (c *RepositoryConfig) TemplateReportsPath(subpaths ...string) string {
	parts := append([]string{c.Conventions.TemplateReportsDir}, subpaths...)
	return c.TemplatePath(parts...)
}

// TemplateReportsPathAbs returns the absolute path to reports templates subdirectory.
func (c *RepositoryConfig) TemplateReportsPathAbs(workspaceRoot string, subpaths ...string) string {
	parts := append([]string{c.Conventions.TemplateReportsDir}, subpaths...)
	return c.TemplatePathAbs(workspaceRoot, parts...)
}

// SpecsFeaturePath returns the path to a feature specification file
// For module-scoped features: SpecsFeaturePath(moduleName, featureName)
// For top-level features: SpecsFeaturePath("", featureName).
func (c *RepositoryConfig) SpecsFeaturePath(moduleName, featureName string) string {
	if moduleName == "" {
		return c.Paths.SpecsRoot + "/" + featureName + "/" + c.Conventions.Specification
	}
	return c.SpecsPath(moduleName) + "/" + featureName + "/" + c.Conventions.Specification
}

// SpecsFeaturePathAbs returns the absolute path to a feature specification file.
func (c *RepositoryConfig) SpecsFeaturePathAbs(workspaceRoot, moduleName, featureName string) string {
	return filepath.Join(workspaceRoot, c.SpecsFeaturePath(moduleName, featureName))
}

// GetModule returns a module by moniker.
func (c *RepositoryConfig) GetModule(moniker string) (*Module, bool) {
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

// GetContainerConfig returns the configuration for a specific container tool.
// Returns nil if the container is not configured.
func (c *RepositoryConfig) GetContainerConfig(name string) *ContainerToolConfig {
	if c.Containers.Tools == nil {
		return nil
	}
	return c.Containers.Tools[name]
}

// GetBaseImageVersion returns the pinned version for a base image.
// Returns empty string if not configured.
func (c *RepositoryConfig) GetBaseImageVersion(name string) string {
	if c.Containers.BaseImages == nil {
		return ""
	}
	return c.Containers.BaseImages[name]
}
