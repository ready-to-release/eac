// Package defaults provides default values and path derivation for module contracts.
package defaults

import (
	"strings"
)

// SubstituteVariables replaces placeholders in a pattern with actual values.
// Supported variables: {moniker}, {root}, {type}, and any custom path variables.
// All paths use forward slashes (/) for cross-platform compatibility.
func SubstituteVariables(pattern, moniker, root, moduleType string, pathVars map[string]string) string {
	result := pattern
	result = strings.ReplaceAll(result, "{moniker}", moniker)
	result = strings.ReplaceAll(result, "{root}", root)
	result = strings.ReplaceAll(result, "{type}", moduleType)

	// Apply repository-level path variables
	for key, value := range pathVars {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}

	return result
}

// SubstituteAll applies variable substitution to all patterns in a slice.
func SubstituteAll(patterns []string, moniker, root, moduleType string, pathVars map[string]string) []string {
	if patterns == nil {
		return nil
	}
	result := make([]string, len(patterns))
	for i, p := range patterns {
		result[i] = SubstituteVariables(p, moniker, root, moduleType, pathVars)
	}
	return result
}

// TypeDefaultsApplier provides methods to apply type-based defaults to modules.
// This interface allows the defaults package to work with any type configuration
// without importing the config package (avoiding import cycles).
type TypeDefaultsApplier interface {
	// GetTypeDefaults returns the defaults for a given module type.
	// Returns nil if the type has no defaults configured.
	GetTypeDefaults(typeName string) *TypeDefaults
}

// TypeDefaults represents the default values for a module type.
// This mirrors config.TypeDefaults but lives in the defaults package
// to avoid import cycles.
type TypeDefaults struct {
	Files *FilesDefaults
	Repo  *RepoDefaults
}

// FilesDefaults contains default file patterns for a module type.
type FilesDefaults struct {
	Source          []string
	Config          []string
	Assets          []string
	Tests           []string
	Changelog       string
	WorkflowCI      string
	WorkflowRelease string
}

// RepoDefaults contains default repo-level configurations.
type RepoDefaults struct {
	Specs    []string
	TestImpl string
	Design   string
}

// ApplyTypeDefaults applies type-specific defaults to module fields.
// Only applies defaults to fields that are nil/empty (doesn't override explicit values).
// Returns the applied values.
type ModuleDefaults struct {
	Source          []string
	Config          []string
	Assets          []string
	Tests           []string
	Changelog       string
	WorkflowCI      string
	WorkflowRelease string
	Specs           []string
	TestImpl        string
	Design          string
}

// ResolveDefaults resolves all defaults for a module, combining type-specific
// defaults with repository path variables. Explicit values in the module take precedence.
//
// Priority order (highest to lowest):
// 1. Explicit value in module config
// 2. Type-specific default (with variable substitution from repository.yml)
func ResolveDefaults(
	typeDef *TypeDefaults,
	moniker, root, moduleType string,
	pathVars map[string]string,
	// Current values (nil means not set)
	source, config, assets, tests []string,
	changelog string,
	workflowCI, workflowRelease string,
	specs []string,
	testImpl, design string,
) ModuleDefaults {
	result := ModuleDefaults{}

	// Source - type default only (root-based ownership is handled in MatchesFile)
	if source != nil {
		result.Source = source
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Source != nil {
		result.Source = SubstituteAll(typeDef.Files.Source, moniker, root, moduleType, pathVars)
	}

	// Config - type default only
	if config != nil {
		result.Config = config
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Config != nil {
		result.Config = SubstituteAll(typeDef.Files.Config, moniker, root, moduleType, pathVars)
	}

	// Assets - type default only
	if assets != nil {
		result.Assets = assets
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Assets != nil {
		result.Assets = SubstituteAll(typeDef.Files.Assets, moniker, root, moduleType, pathVars)
	}

	// Tests - type default only
	if tests != nil {
		result.Tests = tests
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Tests != nil {
		result.Tests = SubstituteAll(typeDef.Files.Tests, moniker, root, moduleType, pathVars)
	}

	// Changelog - type default, then generic default
	if changelog != "" {
		result.Changelog = changelog
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Changelog != "" {
		result.Changelog = typeDef.Files.Changelog
	} else {
		result.Changelog = Changelog
	}

	// WorkflowCI - type default, then generic default
	if workflowCI != "" {
		result.WorkflowCI = workflowCI
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.WorkflowCI != "" {
		result.WorkflowCI = SubstituteVariables(typeDef.Files.WorkflowCI, moniker, root, moduleType, pathVars)
	} else {
		result.WorkflowCI = WorkflowCIPath(moniker)
	}

	// WorkflowRelease - type default, then generic default
	if workflowRelease != "" {
		result.WorkflowRelease = workflowRelease
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.WorkflowRelease != "" {
		result.WorkflowRelease = SubstituteVariables(typeDef.Files.WorkflowRelease, moniker, root, moduleType, pathVars)
	} else {
		result.WorkflowRelease = WorkflowReleasePath(moniker)
	}

	// Specs - type default, then generic default
	if specs != nil {
		result.Specs = specs
	} else if typeDef != nil && typeDef.Repo != nil && typeDef.Repo.Specs != nil {
		result.Specs = SubstituteAll(typeDef.Repo.Specs, moniker, root, moduleType, pathVars)
	} else {
		result.Specs = []string{SpecsPattern(moniker)}
	}

	// TestImpl - type default only (must come from repository.yml via type defaults)
	if testImpl != "" {
		result.TestImpl = testImpl
	} else if typeDef != nil && typeDef.Repo != nil && typeDef.Repo.TestImpl != "" {
		result.TestImpl = SubstituteVariables(typeDef.Repo.TestImpl, moniker, root, moduleType, pathVars)
	}

	// Design - type default, then generic default
	if design != "" {
		result.Design = design
	} else if typeDef != nil && typeDef.Repo != nil && typeDef.Repo.Design != "" {
		result.Design = SubstituteVariables(typeDef.Repo.Design, moniker, root, moduleType, pathVars)
	} else {
		result.Design = DesignPath(moniker)
	}

	return result
}
