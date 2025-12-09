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

// LoadModulesDefaults loads default modules from contract defaults.
// Returns default modules config when .r2r/eac/modules.yml doesn't exist.
func LoadModulesDefaults(repoRoot string) (*ModulesConfig, error) {
	data, err := loadDefaultFile(repoRoot, "modules.yml")
	if err != nil {
		return nil, fmt.Errorf("loading modules defaults: %w", err)
	}

	var cfg ModulesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing modules defaults: %w", err)
	}

	cfg.applyDefaults()
	return &cfg, nil
}

// LoadModuleTypesDefaults loads default module types from contract defaults.
// These are merged with user-defined types (user types override defaults).
func LoadModuleTypesDefaults(repoRoot string) (*ModuleTypesConfig, error) {
	data, err := loadDefaultFile(repoRoot, "module-types.yml")
	if err != nil {
		return nil, fmt.Errorf("loading module-types defaults: %w", err)
	}

	var cfg ModuleTypesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing module-types defaults: %w", err)
	}

	cfg.buildTypeMap()
	return &cfg, nil
}

// LoadRepositoryDefaults loads default repository config from contract defaults.
func LoadRepositoryDefaults(repoRoot string) (*RepositoryConfig, error) {
	data, err := loadDefaultFile(repoRoot, "repository.yml")
	if err != nil {
		return nil, fmt.Errorf("loading repository defaults: %w", err)
	}

	var cfg RepositoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing repository defaults: %w", err)
	}

	return &cfg, nil
}

// LoadSystemDependenciesDefaults loads default system dependencies from contract defaults.
func LoadSystemDependenciesDefaults(repoRoot string) (*SystemDependenciesConfig, error) {
	data, err := loadDefaultFile(repoRoot, "system-dependencies.yml")
	if err != nil {
		return nil, fmt.Errorf("loading system-dependencies defaults: %w", err)
	}

	var cfg SystemDependenciesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing system-dependencies defaults: %w", err)
	}

	return &cfg, nil
}

// defaultsRoot returns the root directory for loading contract defaults.
// In container mode (R2R_CONTAINER_ROOT set), uses the container's internal
// /app directory where contracts are baked in. Otherwise uses repoRoot.
func defaultsRoot(repoRoot string) string {
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
		return containerRoot
	}
	return repoRoot
}

// loadDefaultFile loads a default YAML file from the contracts folder.
// Container-aware: uses R2R_CONTAINER_ROOT/contracts/ when running in container,
// otherwise uses repoRoot/contracts/.
func loadDefaultFile(repoRoot, filename string) ([]byte, error) {
	root := defaultsRoot(repoRoot)
	if root == "" {
		return nil, fmt.Errorf("no root available for loading defaults (repoRoot empty and not in container)")
	}

	fsPath := filepath.Join(root, "contracts", "eac-core", DefaultsVersion, "defaults", filename)
	data, err := os.ReadFile(fsPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", fsPath, err)
	}

	log.Debugf("Loaded defaults from: %s", fsPath)
	return data, nil
}

// MergeModuleTypes merges user-defined module types with defaults.
// User types override defaults with the same name.
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

	// Merge user types (override or append)
	for _, userType := range user.Types {
		if idx, exists := typeMap[userType.Name]; exists {
			// Override existing type
			result.Types[idx] = userType
		} else {
			// Append new type
			result.Types = append(result.Types, userType)
		}
	}

	result.buildTypeMap()
	return result
}

// MergeRepository merges user repository config with defaults at field level.
// User values override defaults.
func MergeRepository(defaults, user *RepositoryConfig) *RepositoryConfig {
	if defaults == nil {
		return user
	}
	if user == nil {
		return defaults
	}

	result := *defaults // Copy defaults

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
