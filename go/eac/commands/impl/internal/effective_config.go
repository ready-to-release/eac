// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// EffectiveModule represents a module with all type defaults merged and paths resolved
type EffectiveModule struct {
	// Core fields from Module
	Moniker     string            `json:"moniker" yaml:"moniker"`
	Name        string            `json:"name" yaml:"name"`
	Type        string            `json:"type" yaml:"type"`
	Description string            `json:"description" yaml:"description"`
	DependsOn   []string          `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	Files       config.Files      `json:"files" yaml:"files"`
	Flags       config.Flags      `json:"flags" yaml:"flags"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Calculated fields from type
	EffectiveBuildDeps    []string `json:"effective_build_deps,omitempty" yaml:"effective_build_deps,omitempty"`
	EffectiveCapabilities []string `json:"effective_capabilities,omitempty" yaml:"effective_capabilities,omitempty"`
	EffectiveTestFramework string  `json:"effective_test_framework,omitempty" yaml:"effective_test_framework,omitempty"`
	EffectiveBDDFramework  string  `json:"effective_bdd_framework,omitempty" yaml:"effective_bdd_framework,omitempty"`
	ArtifactCount          int      `json:"artifact_count,omitempty" yaml:"artifact_count,omitempty"`

	// Resolved path patterns
	ResolvedPaths map[string]string `json:"resolved_paths,omitempty" yaml:"resolved_paths,omitempty"`
}

// PathVariables represents repository-wide path variables available for substitution
type PathVariables map[string]string

// GetEffectiveModuleConfig merges module configuration with type defaults and resolves paths
func GetEffectiveModuleConfig(
	module *config.Module,
	moduleType *config.ModuleTypeDef,
	pathVars PathVariables,
) (*EffectiveModule, error) {
	if module == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	effective := &EffectiveModule{
		Moniker:      module.Moniker,
		Name:         module.Name,
		Type:         module.Type,
		Description:  module.Description,
		DependsOn:    module.DependsOn,
		Files:        module.Files,
		Flags:        module.Flags,
		Metadata:     module.Metadata,
		ResolvedPaths: make(map[string]string),
	}

	// If type is available, extract calculated fields
	if moduleType != nil {
		effective.EffectiveBuildDeps = moduleType.BuildDeps
		effective.EffectiveCapabilities = moduleType.Capabilities
		effective.EffectiveTestFramework = moduleType.TestFramework
		effective.EffectiveBDDFramework = moduleType.BDDFramework

		// Count artifacts
		if moduleType.Build != nil {
			effective.ArtifactCount = len(moduleType.Build.Artifacts)
		}

		// Merge type defaults into files if not already set
		if moduleType.Defaults != nil && moduleType.Defaults.Files != nil {
			mergeFileDefaults(&effective.Files, moduleType.Defaults.Files)
		}

		// Merge repo defaults
		if moduleType.Defaults != nil && moduleType.Defaults.Repo != nil {
			mergeRepoDefaults(&effective.Files.Repo, moduleType.Defaults.Repo, module.Moniker, pathVars)
		}
	}

	// Resolve path variables in repo paths
	effective.ResolvedPaths = resolveRepoPaths(effective.Files.Repo, module.Moniker, pathVars)

	return effective, nil
}

// mergeFileDefaults merges type defaults into module files (only if module didn't override)
func mergeFileDefaults(files *config.Files, defaults *config.FilesDefaults) {
	if len(files.Source) == 0 && len(defaults.Source) > 0 {
		files.Source = defaults.Source
	}
	if len(files.Tests) == 0 && len(defaults.Tests) > 0 {
		files.Tests = defaults.Tests
	}
	if len(files.Config) == 0 && len(defaults.Config) > 0 {
		files.Config = defaults.Config
	}
	if len(files.Assets) == 0 && len(defaults.Assets) > 0 {
		files.Assets = defaults.Assets
	}
	if files.Changelog == "" && defaults.Changelog != "" {
		files.Changelog = defaults.Changelog
	}
	if defaults.Workflows != nil {
		if files.Workflows.CI == "" && defaults.Workflows.CI != "" {
			files.Workflows.CI = defaults.Workflows.CI
		}
		if files.Workflows.Release == "" && defaults.Workflows.Release != "" {
			files.Workflows.Release = defaults.Workflows.Release
		}
	}
}

// mergeRepoDefaults merges type repo defaults into module repo files
func mergeRepoDefaults(
	repo *config.RepoFiles,
	defaults *config.RepoDefaults,
	moniker string,
	pathVars PathVariables,
) {
	// Merge specs patterns
	if len(repo.Specs) == 0 && len(defaults.Specs) > 0 {
		repo.Specs = resolvePatterns(defaults.Specs, moniker, pathVars)
	}

	// Merge test_impl path
	if repo.TestImpl == "" && defaults.TestImpl != "" {
		repo.TestImpl = resolvePattern(defaults.TestImpl, moniker, pathVars)
	}

	// Merge design path
	if repo.Design == "" && defaults.Design != "" {
		repo.Design = resolvePattern(defaults.Design, moniker, pathVars)
	}
}

// resolveRepoPaths resolves all path patterns in repo files to actual paths
func resolveRepoPaths(repo config.RepoFiles, moniker string, pathVars PathVariables) map[string]string {
	resolved := make(map[string]string)

	// Resolve specs patterns (multiple)
	if len(repo.Specs) > 0 {
		resolved["specs"] = strings.Join(repo.Specs, ", ")
	}

	// Resolve test_impl
	if repo.TestImpl != "" {
		resolved["test_impl"] = repo.TestImpl
	}

	// Resolve design
	if repo.Design != "" {
		resolved["design"] = repo.Design
	}

	return resolved
}

// resolvePatterns resolves multiple patterns with variable substitution
func resolvePatterns(patterns []string, moniker string, pathVars PathVariables) []string {
	resolved := make([]string, len(patterns))
	for i, pattern := range patterns {
		resolved[i] = resolvePattern(pattern, moniker, pathVars)
	}
	return resolved
}

// resolvePattern resolves a single pattern with variable substitution
func resolvePattern(pattern string, moniker string, pathVars PathVariables) string {
	result := pattern
	result = strings.ReplaceAll(result, "{moniker}", moniker)

	// Replace all path variables
	for key, value := range pathVars {
		result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", key), value)
	}

	return result
}

// GetPathVariables extracts path variables from repository configuration
func GetPathVariables(repoConfig *config.RepositoryConfig) PathVariables {
	vars := make(PathVariables)

	if repoConfig == nil {
		return vars
	}

	// Use the existing GetPathVariables method
	return repoConfig.GetPathVariables()
}
