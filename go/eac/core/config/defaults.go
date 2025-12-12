// Package config provides defaults loading from contract YAML files.
// Defaults are loaded at runtime from contracts/eac-core/0.1.0/defaults/*.yml
// and merged with user config from .r2r/eac/*.yml.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultsVersion is the contract version for defaults
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
// Returns nil if the type-specific defaults file doesn't exist (not an error).
// Type-specific defaults are merged BETWEEN base defaults and user config.
func LoadRepositoryTypeDefaults(repoRoot, repoType string) (*RepositoryConfig, error) {
	if repoType == "" {
		return nil, nil
	}

	filename := fmt.Sprintf("repository-%s.yml", repoType)
	data, err := loadDefaultFile(repoRoot, filename)
	if err != nil {
		// Type-specific defaults are optional
		if os.IsNotExist(err) {
			return nil, nil
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
// Returns nil (not error) when defaults don't exist - allows tests to work without contracts folder.
func LoadRepositoryDefaults(repoRoot string) (*RepositoryConfig, error) {
	data, err := loadDefaultFile(repoRoot, "repository.yml")
	if err != nil {
		// Defaults are optional - return nil if they don't exist
		if os.IsNotExist(err) {
			return nil, nil
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

// LoadModuleTypesDefaults loads default module types from contract defaults.
// These are merged with user-defined types (user types override defaults).
// Returns nil (not error) when defaults don't exist - allows tests to use their own types.
func LoadModuleTypesDefaults(repoRoot string) (*ModuleTypesConfig, error) {
	data, err := loadDefaultFile(repoRoot, "module-types.yml")
	if err != nil {
		// Defaults are optional - return nil if they don't exist
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading module-types defaults: %w", err)
	}

	var cfg ModuleTypesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing module-types defaults: %w", err)
	}

	cfg.buildTypeMap()
	return &cfg, nil
}

// LoadSystemDependenciesDefaults loads default system dependencies from contract defaults.
// Returns nil (not error) when defaults don't exist - allows tests to work without contracts folder.
func LoadSystemDependenciesDefaults(repoRoot string) (*SystemDependenciesConfig, error) {
	data, err := loadDefaultFile(repoRoot, "system-dependencies.yml")
	if err != nil {
		// Defaults are optional - return nil if they don't exist
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading system-dependencies defaults: %w", err)
	}

	var cfg SystemDependenciesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing system-dependencies defaults: %w", err)
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

	log.Debugf("Loaded defaults from: %s", fsPath)
	return data, nil
}

// MergeModuleTypes merges user-defined module types with defaults.
// User types are merged with defaults of the same name at field level.
// User values override defaults when specified.
// Returns a new config with all types.
func MergeModuleTypes(defaults, user *ModuleTypesConfig) *ModuleTypesConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	// Start with defaults
	result := &ModuleTypesConfig{
		Types: make([]ModuleTypeDef, len(defaults.Types)),
	}
	copy(result.Types, defaults.Types)

	// Build map for fast lookup
	typeMap := make(map[string]int)
	for i, t := range result.Types {
		typeMap[t.Name] = i
	}

	// Merge user types (field-level merge or append)
	for _, userType := range user.Types {
		if idx, exists := typeMap[userType.Name]; exists {
			// Field-level merge with existing type
			result.Types[idx] = mergeModuleTypeDef(result.Types[idx], userType)
		} else {
			// Append new type
			result.Types = append(result.Types, userType)
		}
	}

	result.buildTypeMap()
	return result
}

// mergeModuleTypeDef merges a user type definition with a default type definition.
// User values override defaults when specified (non-empty).
func mergeModuleTypeDef(base, user ModuleTypeDef) ModuleTypeDef {
	result := base

	// Override simple fields if user specifies them
	if user.Description != "" {
		result.Description = user.Description
	}
	if len(user.BuildDeps) > 0 {
		result.BuildDeps = user.BuildDeps
	}
	if len(user.Capabilities) > 0 {
		result.Capabilities = user.Capabilities
	}
	if user.TestFramework != "" {
		result.TestFramework = user.TestFramework
	}
	if user.BDDFramework != "" {
		result.BDDFramework = user.BDDFramework
	}
	if user.DockerBuild != nil {
		result.DockerBuild = user.DockerBuild
	}
	if user.Build != nil {
		result.Build = user.Build
	}

	// Merge defaults at field level if user has partial defaults
	if user.Defaults != nil {
		if result.Defaults == nil {
			result.Defaults = user.Defaults
		} else {
			result.Defaults = mergeTypeDefaults(result.Defaults, user.Defaults)
		}
	}

	return result
}

// mergeTypeDefaults merges user type defaults with base defaults.
func mergeTypeDefaults(base, user *TypeDefaults) *TypeDefaults {
	if base == nil {
		return user
	}
	if user == nil {
		return base
	}

	result := &TypeDefaults{}

	// Merge Files
	if base.Files != nil || user.Files != nil {
		result.Files = mergeFilesDefaults(base.Files, user.Files)
	}

	// Merge Repo
	if base.Repo != nil || user.Repo != nil {
		result.Repo = mergeRepoDefaults(base.Repo, user.Repo)
	}

	return result
}

// mergeFilesDefaults merges file defaults.
func mergeFilesDefaults(base, user *FilesDefaults) *FilesDefaults {
	if base == nil {
		return user
	}
	if user == nil {
		return base
	}

	result := *base // Copy base

	// Override with user values if specified
	if len(user.Source) > 0 {
		result.Source = user.Source
	}
	if len(user.Config) > 0 {
		result.Config = user.Config
	}
	if len(user.Assets) > 0 {
		result.Assets = user.Assets
	}
	if len(user.Tests) > 0 {
		result.Tests = user.Tests
	}
	if user.Changelog != "" {
		result.Changelog = user.Changelog
	}
	if user.Workflows != nil {
		if result.Workflows == nil {
			result.Workflows = user.Workflows
		} else {
			merged := *result.Workflows
			if user.Workflows.CI != "" {
				merged.CI = user.Workflows.CI
			}
			if user.Workflows.Release != "" {
				merged.Release = user.Workflows.Release
			}
			result.Workflows = &merged
		}
	}

	return &result
}

// mergeRepoDefaults merges repo defaults.
func mergeRepoDefaults(base, user *RepoDefaults) *RepoDefaults {
	if base == nil {
		return user
	}
	if user == nil {
		return base
	}

	result := *base // Copy base

	// Override with user values if specified
	if len(user.Specs) > 0 {
		result.Specs = user.Specs
	}
	if user.TestImpl != "" {
		result.TestImpl = user.TestImpl
	}
	if user.Design != "" {
		result.Design = user.Design
	}

	return &result
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

	// Override paths if set
	if user.Paths.SpecsRoot != "" {
		result.Paths.SpecsRoot = user.Paths.SpecsRoot
	}
	if user.Paths.TestImplRoot != "" {
		result.Paths.TestImplRoot = user.Paths.TestImplRoot
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
	if user.Paths.Out.Security != "" {
		result.Paths.Out.Security = user.Paths.Out.Security
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

// MergeSystemDependencies merges user system dependencies with defaults.
// User entries with same moniker override defaults, new entries are appended.
func MergeSystemDependencies(defaults, user *SystemDependenciesConfig) *SystemDependenciesConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	// Start with defaults
	result := &SystemDependenciesConfig{
		Dependencies:           make([]SystemDependency, len(defaults.Dependencies)),
		CapabilityRequirements: defaults.CapabilityRequirements,
	}
	copy(result.Dependencies, defaults.Dependencies)

	// Build map for fast lookup
	depMap := make(map[string]int)
	for i, dep := range result.Dependencies {
		depMap[dep.Moniker] = i
	}

	// Merge user dependencies (override or append)
	for _, userDep := range user.Dependencies {
		if idx, exists := depMap[userDep.Moniker]; exists {
			// Merge: override only non-empty fields
			existing := &result.Dependencies[idx]
			if userDep.Name != "" {
				existing.Name = userDep.Name
			}
			if userDep.Description != "" {
				existing.Description = userDep.Description
			}
			if userDep.Version != "" {
				existing.Version = userDep.Version
			}
			if userDep.Verify.Command != "" || len(userDep.Verify.EnvVars) > 0 {
				existing.Verify = userDep.Verify
			}
		} else {
			// Append new dependency
			result.Dependencies = append(result.Dependencies, userDep)
			depMap[userDep.Moniker] = len(result.Dependencies) - 1
		}
	}

	// Merge capability requirements if user provides any
	if user.CapabilityRequirements != nil {
		if result.CapabilityRequirements == nil {
			result.CapabilityRequirements = make(map[string][]string)
		}
		for cap, deps := range user.CapabilityRequirements {
			result.CapabilityRequirements[cap] = deps
		}
	}

	result.buildDepMap()
	return result
}
