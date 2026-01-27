// Package config provides defaults loading from contract YAML files.
// Defaults are loaded at runtime from contracts/eac-core/0.1.0/defaults/*.yml
// and merged with user config from .r2r/eac/*.yml.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ErrNoDefaults is returned when a defaults file doesn't exist.
// This is not an error condition - it just means defaults should be skipped.
var ErrNoDefaults = errors.New("defaults file not found")

// DefaultsVersion is the contract version for defaults.
const DefaultsVersion = "0.1.0"

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
	if repoType == "" {
		return nil, ErrNoDefaults
	}

	filename := fmt.Sprintf("repository-%s.yml", repoType)
	data, err := loadDefaultFile(repoRoot, filename)
	if err != nil {
		// Type-specific defaults are optional
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading repository type defaults (%s): %w", repoType, err)
	}

	var cfg RepositoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing repository type defaults (%s): %w", repoType, err)
	}

	return &cfg, nil
}

// LoadRepositoryDefaults loads default repository config from contract defaults.
// This now includes modules (unified config).
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadRepositoryDefaults(repoRoot string) (*RepositoryConfig, error) {
	data, err := loadDefaultFile(repoRoot, "repository.yml")
	if err != nil {
		// Defaults are optional - return ErrNoDefaults if they don't exist
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading repository defaults: %w", err)
	}

	var cfg RepositoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing repository defaults: %w", err)
	}

	// Apply module defaults
	cfg.applyModuleDefaults()

	return &cfg, nil
}

// LoadComponentTypesDefaults loads default component types from contract defaults.
// Returns ErrNoDefaults when defaults don't exist - allows tests to work without contracts folder.
func LoadComponentTypesDefaults(repoRoot string) (*ComponentTypesConfig, error) {
	data, err := loadDefaultFile(repoRoot, ComponentTypesFileName)
	if err != nil {
		// Defaults are optional - return ErrNoDefaults if they don't exist
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading component-types defaults: %w", err)
	}

	var cfg ComponentTypesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing component-types defaults: %w", err)
	}

	return &cfg, nil
}

// defaultsRoot returns the root directory for loading contract defaults.
// Uses the distribution root (container root if in container, otherwise repoRoot).
// Note: Can't import repository package here to avoid cycles, so inline the check.
// See repository.GetDistRoot() for the canonical implementation.
func defaultsRoot(repoRoot string) string {
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
		return containerRoot
	}
	return repoRoot
}

// loadDefaultFile loads a default YAML file from the contracts folder.
// Container-aware: uses R2R_CONTAINER_ROOT/contracts/ when running in container,
// otherwise uses repoRoot/contracts/.
// Returns the raw os error (not wrapped) so callers can check os.IsNotExist.
func loadDefaultFile(repoRoot, filename string) ([]byte, error) {
	root := defaultsRoot(repoRoot)
	if root == "" {
		return nil, fmt.Errorf("no root available for loading defaults (repoRoot empty and not in container)")
	}

	fsPath := filepath.Join(root, "contracts", "eac-core", DefaultsVersion, "defaults", filename)
	data, err := os.ReadFile(fsPath)
	if err != nil {
		// Return raw error for IsNotExist checks, wrapped for other errors
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("reading %s: %w", fsPath, err)
	}

	return data, nil
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
	data, err := loadDefaultFile(repoRoot, "test-suites.yml")
	if err != nil {
		// Defaults are optional - return ErrNoDefaults if they don't exist
		if os.IsNotExist(err) {
			return nil, ErrNoDefaults
		}
		return nil, fmt.Errorf("loading test-suites defaults: %w", err)
	}

	var cfg TestSuitesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing test-suites defaults: %w", err)
	}

	cfg.buildSuiteMap()
	return &cfg, nil
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

