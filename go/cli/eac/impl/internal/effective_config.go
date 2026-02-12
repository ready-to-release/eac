// Package internal provides shared infrastructure for GET and SHOW commands
package internal

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/paths"
)

// EffectiveModule represents a module with paths resolved.
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

	// Resolved package roots
	PackageRoots map[string]string `json:"package_roots,omitempty" yaml:"package_roots,omitempty"`
}

// PathVariables represents repository-wide path variables available for substitution.
type PathVariables map[string]string

// GetEffectiveModuleConfig returns module configuration with paths resolved.
func GetEffectiveModuleConfig(
	module *config.Module,
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
	if repoConfig == nil {
		return make(PathVariables)
	}

	// Get config-aware path variables and add out_root
	vars := PathVariables(paths.GetPathVariablesC(repoConfig.ToPathConfig()))
	vars["out_root"] = repoConfig.Paths.Out.Root
	return vars
}
