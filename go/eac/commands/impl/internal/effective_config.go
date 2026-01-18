// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// EffectiveModule represents a module with all type defaults merged and paths resolved.
type EffectiveModule struct {
	// Core fields from Module
	Moniker     string            `json:"moniker" yaml:"moniker"`
	Name        string            `json:"name" yaml:"name"`
	Type        string            `json:"type" yaml:"type"`
	Description string            `json:"description" yaml:"description"`
	DependsOn   []string          `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	Packages    []string          `json:"packages,omitempty" yaml:"packages,omitempty"`
	Changelog   string            `json:"changelog,omitempty" yaml:"changelog,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Calculated fields from type
	EffectiveBuildDeps     []string `json:"effective_build_deps,omitempty" yaml:"effective_build_deps,omitempty"`
	EffectiveCapabilities  []string `json:"effective_capabilities,omitempty" yaml:"effective_capabilities,omitempty"`
	EffectiveTestFramework string   `json:"effective_test_framework,omitempty" yaml:"effective_test_framework,omitempty"`
	EffectiveBDDFramework  string   `json:"effective_bdd_framework,omitempty" yaml:"effective_bdd_framework,omitempty"`
	ArtifactCount          int      `json:"artifact_count,omitempty" yaml:"artifact_count,omitempty"`

	// Resolved package roots
	PackageRoots map[string]string `json:"package_roots,omitempty" yaml:"package_roots,omitempty"`
}

// PathVariables represents repository-wide path variables available for substitution.
type PathVariables map[string]string

// GetEffectiveModuleConfig merges module configuration with type defaults and resolves paths.
func GetEffectiveModuleConfig(
	module *config.Module,
	moduleType *config.ModuleTypeDef,
	pathVars PathVariables,
) (*EffectiveModule, error) {
	if module == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	// Collect package roots
	packageRoots := make(map[string]string)
	packages := []string{}
	for pkgName, entry := range module.Components {
		packages = append(packages, pkgName)
		if entry != nil && entry.Root != "" {
			packageRoots[pkgName] = entry.Root
		}
	}

	effective := &EffectiveModule{
		Moniker:      module.Moniker,
		Name:         module.Name,
		Type:         module.GetComponentTypesDisplay(),
		Description:  module.Description,
		DependsOn:    module.DependsOn,
		Packages:     packages,
		Changelog:    module.GetChangelog(),
		Metadata:     module.Metadata,
		PackageRoots: packageRoots,
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
	}

	return effective, nil
}

// resolvePattern resolves a single pattern with variable substitution.
func resolvePattern(pattern, moniker string, pathVars PathVariables) string {
	result := pattern
	result = strings.ReplaceAll(result, "{moniker}", moniker)

	// Replace all path variables
	for key, value := range pathVars {
		result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", key), value)
	}

	return result
}

// GetPathVariables extracts path variables from repository configuration.
func GetPathVariables(repoConfig *config.RepositoryConfig) PathVariables {
	vars := make(PathVariables)

	if repoConfig == nil {
		return vars
	}

	// Use the existing GetPathVariables method
	return repoConfig.GetPathVariables()
}
