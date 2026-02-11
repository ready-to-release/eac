// Package defaults provides default values and path derivation for module domain.
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

// ModuleDefaults holds the resolved default values for a module.
// These values are applied to fields that are nil/empty (doesn't override explicit values).
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

// resolveSliceField resolves a slice field using a 2-tier priority:
// explicit value (non-nil) > type default (non-nil, with variable substitution) > fallback.
// A nil explicit value means "not set"; a non-nil slice (even if empty) counts as explicit.
func resolveSliceField(
	explicit, typeDefault, fallback []string,
	moniker, root, moduleType string,
	pathVars map[string]string,
) []string {
	if explicit != nil {
		return explicit
	}
	if typeDefault != nil {
		return SubstituteAll(typeDefault, moniker, root, moduleType, pathVars)
	}
	return fallback
}

// resolveStringField resolves a string field using a 2-tier priority:
// explicit value (non-empty) > type default (non-empty, with optional variable substitution) > fallback.
// When substitute is true, SubstituteVariables is applied to the type default.
func resolveStringField(
	explicit, typeDefault, fallback string,
	substitute bool,
	moniker, root, moduleType string,
	pathVars map[string]string,
) string {
	if explicit != "" {
		return explicit
	}
	if typeDefault != "" {
		if substitute {
			return SubstituteVariables(typeDefault, moniker, root, moduleType, pathVars)
		}
		return typeDefault
	}
	return fallback
}

// ResolveDefaults resolves all defaults for a module, combining type-specific
// defaults with repository path variables. Explicit values in the module take precedence.
//
// Priority order (highest to lowest):
// 1. Explicit value in module config
// 2. Type-specific default (with variable substitution from repository.yml).
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
	// Extract sub-structs once to avoid repeated nil checks on typeDef.
	var files *FilesDefaults
	var repo *RepoDefaults
	if typeDef != nil {
		files = typeDef.Files
		repo = typeDef.Repo
	}

	// File-level slice helpers: pull the type default from files (may be nil).
	var fileSrc, fileCfg, fileAssets, fileTests []string
	if files != nil {
		fileSrc = files.Source
		fileCfg = files.Config
		fileAssets = files.Assets
		fileTests = files.Tests
	}

	// Repo-level slice helper: pull from repo (may be nil).
	var repoSpecs []string
	if repo != nil {
		repoSpecs = repo.Specs
	}

	// File-level string helpers.
	var fileChangelog, fileWorkflowCI, fileWorkflowRelease string
	if files != nil {
		fileChangelog = files.Changelog
		fileWorkflowCI = files.WorkflowCI
		fileWorkflowRelease = files.WorkflowRelease
	}

	// Repo-level string helpers.
	var repoTestImpl, repoDesign string
	if repo != nil {
		repoTestImpl = repo.TestImpl
		repoDesign = repo.Design
	}

	return ModuleDefaults{
		Source: resolveSliceField(source, fileSrc, nil, moniker, root, moduleType, pathVars),
		Config: resolveSliceField(config, fileCfg, nil, moniker, root, moduleType, pathVars),
		Assets: resolveSliceField(assets, fileAssets, nil, moniker, root, moduleType, pathVars),
		Tests:  resolveSliceField(tests, fileTests, nil, moniker, root, moduleType, pathVars),
		Specs:  resolveSliceField(specs, repoSpecs, []string{SpecsPattern(moniker)}, moniker, root, moduleType, pathVars),

		Changelog:       resolveStringField(changelog, fileChangelog, Changelog, false, moniker, root, moduleType, pathVars),
		WorkflowCI:      resolveStringField(workflowCI, fileWorkflowCI, "", true, moniker, root, moduleType, pathVars),
		WorkflowRelease: resolveStringField(workflowRelease, fileWorkflowRelease, "", true, moniker, root, moduleType, pathVars),
		TestImpl:        resolveStringField(testImpl, repoTestImpl, "", true, moniker, root, moduleType, pathVars),
		Design:          resolveStringField(design, repoDesign, DesignPath(moniker), true, moniker, root, moduleType, pathVars),
	}
}
