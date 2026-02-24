// Package config provides defaults loading from contract YAML files.
// Defaults are loaded from the embedded filesystem in contracts/core/0.1.0
// and merged with user config from .eac/*.yml.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"gopkg.in/yaml.v3"
)

// ErrNoDefaults is returned when a defaults file doesn't exist.
// This is not an error condition - it just means defaults should be skipped.
var ErrNoDefaults = errors.New("defaults file not found")

// DefaultModuleDescription is the description of the placeholder module inserted when
// no modules are declared in repository.yml. This constant is used by the init command
// to identify and remove the placeholder during re-initialization.
// Must match the value in contracts/core/0.1.0/schemas/defaults/repository.yml.
const DefaultModuleDescription = "Auto-generated module for unconfigured repositories"

// peekRepositoryType reads only the repository.type field from user config.
// This is a minimal read to determine which type-specific defaults to load.
// Returns empty string if file doesn't exist or type is not specified.
func peekRepositoryType(configRoot string) (string, error) {
	repoPath := filepath.Join(configRoot, RepositoryFileName)
	data, err := os.ReadFile(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No user config, will use defaults
		}
		return "", fmt.Errorf("reading repository config: %w", err)
	}

	// Minimal struct to extract just the type
	var peek struct {
		Repository struct {
			Type string `yaml:"type"`
		} `yaml:"repository"`
	}

	if err := yaml.Unmarshal(data, &peek); err != nil {
		return "", fmt.Errorf("parsing repository type: %w", err)
	}

	return peek.Repository.Type, nil
}

// LoadRepositoryTypeDefaults loads type-specific repository defaults.
// Returns ErrNoDefaults if the type-specific defaults file doesn't exist (not an error condition).
// Type-specific defaults are merged BETWEEN base defaults and user config.
func LoadRepositoryTypeDefaults(repoRoot, repoType string) (*RepositoryConfig, error) {
	cfg, err := cloneRepoTypeDefaults(repoType)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading repository type defaults (%s): %w", repoType, err)
	}
	return cfg, nil
}

// LoadRepositoryDefaults loads default repository config from contract defaults.
// This now includes modules (unified config).
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadRepositoryDefaults(repoRoot string) (*RepositoryConfig, error) {
	cfg, err := cloneRepositoryDefaults()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading repository defaults: %w", err)
	}

	// Load registries defaults and merge
	registries, err := LoadRegistriesDefaults(repoRoot)
	if err == nil && registries != nil {
		cfg.Registries = registries
	}

	// Apply module defaults
	cfg.applyModuleDefaults()

	return cfg, nil
}

// LoadRegistriesDefaults loads default registries config from contract defaults.
// Returns ErrNoDefaults when defaults don't exist.
func LoadRegistriesDefaults(repoRoot string) (RegistriesConfig, error) {
	cfg, err := cloneRegistriesDefaults()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading registries defaults: %w", err)
	}
	return cfg, nil
}

// LoadEnvironmentsDefaults loads default environments from contract defaults.
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadEnvironmentsDefaults(repoRoot string) (*EnvironmentsConfig, error) {
	cfg, err := cloneEnvironmentsDefaults()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading environments defaults: %w", err)
	}
	return cfg, nil
}

// LoadTestingTagsDefaults loads default testing-tags from contract defaults.
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadTestingTagsDefaults(repoRoot string) (*TestingTagsConfig, error) {
	cfg, err := cloneTestingTagsDefaults()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading testing-tags defaults: %w", err)
	}
	return cfg, nil
}

// LoadTimeoutsDefaults loads default timeouts from contract defaults.
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadTimeoutsDefaults(repoRoot string) (*TimeoutConfig, error) {
	cfg, err := cloneTimeoutsDefaults()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading timeouts defaults: %w", err)
	}
	return cfg, nil
}

// loadDefaultFile loads a default YAML file from the embedded contract filesystem.
// The repoRoot parameter is retained for API compatibility but is no longer used;
// defaults are read from the compiled-in core.FS embedded filesystem.
// Returns an fs.ErrNotExist-wrapping error when the file doesn't exist,
// compatible with os.IsNotExist checks.
func loadDefaultFile(_ string, filename string) ([]byte, error) {
	return core.FS.ReadFile(core.DefaultPath(filename))
}

// MergeRepository merges user repository config with defaults at field level.
// User values override defaults. Now includes modules.
func MergeRepository(defaults, user *RepositoryConfig) *RepositoryConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	result := *defaults // Copy defaults

	// Override repository settings if set
	if user.Repository.Type != "" {
		result.Repository.Type = user.Repository.Type
	}
	if user.Repository.TrunkBranch != "" {
		result.Repository.TrunkBranch = user.Repository.TrunkBranch
	}
	if user.Repository.MaxBranchAgeDays != 0 {
		result.Repository.MaxBranchAgeDays = user.Repository.MaxBranchAgeDays
	}
	if len(user.Repository.Schemes) > 0 {
		result.Repository.Schemes = user.Repository.Schemes
	}
	if user.Repository.PR.MergeStrategy != "" {
		result.Repository.PR = user.Repository.PR
	}
	if user.Repository.Versioning.Constraint != "" {
		result.Repository.Versioning = user.Repository.Versioning
	}
	if user.Repository.Parallelism.CI != 0 {
		result.Repository.Parallelism.CI = user.Repository.Parallelism.CI
	}
	if user.Repository.Parallelism.Devbox != 0 {
		result.Repository.Parallelism.Devbox = user.Repository.Parallelism.Devbox
	}
	if user.Repository.OptimizeGitLsInCI {
		result.Repository.OptimizeGitLsInCI = user.Repository.OptimizeGitLsInCI
	}

	// Override paths if set
	if user.Paths.SpecsRoot != "" {
		result.Paths.SpecsRoot = user.Paths.SpecsRoot
	}
	if user.Paths.Templates != "" {
		result.Paths.Templates = user.Paths.Templates
	}
	if user.Paths.Out.Root != "" {
		result.Paths.Out.Root = user.Paths.Out.Root
	}
	if user.Paths.Out.Build != "" {
		result.Paths.Out.Build = user.Paths.Out.Build
	}
	if user.Paths.Out.Test != "" {
		result.Paths.Out.Test = user.Paths.Out.Test
	}
	if user.Paths.Out.Logs != "" {
		result.Paths.Out.Logs = user.Paths.Out.Logs
	}
	if user.Paths.Out.Scan != "" {
		result.Paths.Out.Scan = user.Paths.Out.Scan
	}
	if user.Paths.Out.Tools != "" {
		result.Paths.Out.Tools = user.Paths.Out.Tools
	}

	// Override conventions if set
	if user.Conventions.GodogTest != "" {
		result.Conventions.GodogTest = user.Conventions.GodogTest
	}
	if user.Conventions.PackageJSON != "" {
		result.Conventions.PackageJSON = user.Conventions.PackageJSON
	}
	if user.Conventions.Changelog != "" {
		result.Conventions.Changelog = user.Conventions.Changelog
	}
	if user.Conventions.BuildLog != "" {
		result.Conventions.BuildLog = user.Conventions.BuildLog
	}
	if user.Conventions.BuildTiming != "" {
		result.Conventions.BuildTiming = user.Conventions.BuildTiming
	}
	if user.Conventions.TestTiming != "" {
		result.Conventions.TestTiming = user.Conventions.TestTiming
	}
	if user.Conventions.Specification != "" {
		result.Conventions.Specification = user.Conventions.Specification
	}
	if user.Conventions.RiskCatalog != "" {
		result.Conventions.RiskCatalog = user.Conventions.RiskCatalog
	}
	if user.Conventions.RiskControlsDir != "" {
		result.Conventions.RiskControlsDir = user.Conventions.RiskControlsDir
	}
	if user.Conventions.RiskReportsCategory != "" {
		result.Conventions.RiskReportsCategory = user.Conventions.RiskReportsCategory
	}
	if user.Conventions.RiskAssessmentTemplate != "" {
		result.Conventions.RiskAssessmentTemplate = user.Conventions.RiskAssessmentTemplate
	}
	if user.Conventions.TestReportsCategory != "" {
		result.Conventions.TestReportsCategory = user.Conventions.TestReportsCategory
	}
	if user.Conventions.TestResultsTemplate != "" {
		result.Conventions.TestResultsTemplate = user.Conventions.TestResultsTemplate
	}
	if user.Conventions.TemplateSpecsDir != "" {
		result.Conventions.TemplateSpecsDir = user.Conventions.TemplateSpecsDir
	}
	if user.Conventions.TemplateReportsDir != "" {
		result.Conventions.TemplateReportsDir = user.Conventions.TemplateReportsDir
	}
	if user.Conventions.TemplateRiskCatalogDir != "" {
		result.Conventions.TemplateRiskCatalogDir = user.Conventions.TemplateRiskCatalogDir
	}
	if user.Conventions.DesignDir != "" {
		result.Conventions.DesignDir = user.Conventions.DesignDir
	}
	if user.Conventions.WorkspaceDSL != "" {
		result.Conventions.WorkspaceDSL = user.Conventions.WorkspaceDSL
	}

	// Override paths for new fields
	if user.Paths.ContainersRoot != "" {
		result.Paths.ContainersRoot = user.Paths.ContainersRoot
	}

	// Merge remote config if set
	if user.Repository.Remote.Type != "" {
		result.Repository.Remote.Type = user.Repository.Remote.Type
	}
	if user.Repository.Remote.Owner != "" {
		result.Repository.Remote.Owner = user.Repository.Remote.Owner
	}
	if user.Repository.Remote.RepoName != "" {
		result.Repository.Remote.RepoName = user.Repository.Remote.RepoName
	}
	if user.Repository.Remote.URL != "" {
		result.Repository.Remote.URL = user.Repository.Remote.URL
	}
	if user.Repository.Remote.PagesURL != "" {
		result.Repository.Remote.PagesURL = user.Repository.Remote.PagesURL
	}
	if user.Repository.Remote.RegistryURL != "" {
		result.Repository.Remote.RegistryURL = user.Repository.Remote.RegistryURL
	}

	// Merge registries: defaults + user overrides
	if result.Registries == nil {
		result.Registries = make(RegistriesConfig)
	}
	for hostname, userReg := range user.Registries {
		if result.Registries[hostname] == nil {
			result.Registries[hostname] = userReg
		} else {
			// Merge individual registry config
			if userReg.Org != "" {
				result.Registries[hostname].Org = userReg.Org
			}
			if userReg.Cleanup != nil {
				result.Registries[hostname].Cleanup = userReg.Cleanup
			}
		}
	}

	// Modules: user modules completely override defaults (no merge)
	if len(user.Modules) > 0 {
		result.Modules = user.Modules
	}

	// Apply module defaults (type placeholder, description from name, etc.)
	result.applyModuleDefaults()

	return &result
}

// LoadTestSuitesDefaults loads default test suites from contract defaults.
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadTestSuitesDefaults(repoRoot string) (*TestSuitesConfig, error) {
	cfg, err := cloneTestSuitesDefaults()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading test-suites defaults: %w", err)
	}
	return cfg, nil
}

// MergeTestSuites merges user-defined test suites with defaults.
// User suites with same moniker override defaults, new suites are appended.
func MergeTestSuites(defaults, user *TestSuitesConfig) *TestSuitesConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	// Start with defaults
	result := &TestSuitesConfig{
		Suites: make([]TestSuiteDef, len(defaults.Suites)),
	}
	copy(result.Suites, defaults.Suites)

	// Build map for fast lookup
	suiteMap := make(map[string]int)
	for i, suite := range result.Suites {
		suiteMap[suite.Moniker] = i
	}

	// Merge user suites (override or append)
	for _, userSuite := range user.Suites {
		if idx, exists := suiteMap[userSuite.Moniker]; exists {
			// User suite completely overrides default suite with same moniker
			result.Suites[idx] = userSuite
		} else {
			// Append new suite
			result.Suites = append(result.Suites, userSuite)
		}
	}

	result.buildSuiteMap()
	return result
}

