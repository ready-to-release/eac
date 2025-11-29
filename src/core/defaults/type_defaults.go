// Package defaults provides default values and path derivation for module contracts.
package defaults

import (
	"strings"
)

// SubstituteVariables replaces placeholders in a pattern with actual values.
// Supported variables: {moniker}, {root}, {type}
// All paths use forward slashes (/) for cross-platform compatibility.
func SubstituteVariables(pattern, moniker, root, moduleType string) string {
	result := pattern
	result = strings.ReplaceAll(result, "{moniker}", moniker)
	result = strings.ReplaceAll(result, "{root}", root)
	result = strings.ReplaceAll(result, "{type}", moduleType)
	return result
}

// SubstituteAll applies variable substitution to all patterns in a slice.
func SubstituteAll(patterns []string, moniker, root, moduleType string) []string {
	if patterns == nil {
		return nil
	}
	result := make([]string, len(patterns))
	for i, p := range patterns {
		result[i] = SubstituteVariables(p, moniker, root, moduleType)
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
	Flags *FlagsDefaults
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

// FlagsDefaults contains default flag values.
type FlagsDefaults struct {
	CatchAll         *bool
	OwnChildrenFiles *bool
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
	CatchAll        bool
	OwnChildren     bool
}

// ResolveDefaults resolves all defaults for a module, combining type-specific
// defaults with generic defaults. Explicit values in the module take precedence.
//
// Priority order (highest to lowest):
// 1. Explicit value in module config
// 2. Type-specific default (with variable substitution)
// 3. Generic default (from this package)
func ResolveDefaults(
	typeDef *TypeDefaults,
	moniker, root, moduleType string,
	// Current values (nil means not set)
	source, config, assets, tests []string,
	changelog string,
	workflowCI, workflowRelease string,
	specs []string,
	testImpl, design string,
	catchAll, ownChildren *bool,
) ModuleDefaults {
	result := ModuleDefaults{}

	// Source - type default only, no generic default
	if source != nil {
		result.Source = source
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Source != nil {
		result.Source = SubstituteAll(typeDef.Files.Source, moniker, root, moduleType)
	}

	// Config - type default only, no generic default
	if config != nil {
		result.Config = config
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Config != nil {
		result.Config = SubstituteAll(typeDef.Files.Config, moniker, root, moduleType)
	}

	// Assets - type default only, no generic default
	if assets != nil {
		result.Assets = assets
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Assets != nil {
		result.Assets = SubstituteAll(typeDef.Files.Assets, moniker, root, moduleType)
	}

	// Tests - type default only, no generic default
	if tests != nil {
		result.Tests = tests
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Tests != nil {
		result.Tests = SubstituteAll(typeDef.Files.Tests, moniker, root, moduleType)
	}

	// Changelog - type default, then generic default
	if changelog != "" {
		result.Changelog = changelog
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.Changelog != "" {
		result.Changelog = typeDef.Files.Changelog
	} else {
		result.Changelog = Changelog // Generic default
	}

	// WorkflowCI - type default, then generic default
	if workflowCI != "" {
		result.WorkflowCI = workflowCI
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.WorkflowCI != "" {
		result.WorkflowCI = SubstituteVariables(typeDef.Files.WorkflowCI, moniker, root, moduleType)
	} else {
		result.WorkflowCI = WorkflowCIPath(moniker) // Generic default
	}

	// WorkflowRelease - type default, then generic default
	if workflowRelease != "" {
		result.WorkflowRelease = workflowRelease
	} else if typeDef != nil && typeDef.Files != nil && typeDef.Files.WorkflowRelease != "" {
		result.WorkflowRelease = SubstituteVariables(typeDef.Files.WorkflowRelease, moniker, root, moduleType)
	} else {
		result.WorkflowRelease = WorkflowReleasePath(moniker) // Generic default
	}

	// Specs - type default, then generic default
	if specs != nil {
		result.Specs = specs
	} else if typeDef != nil && typeDef.Repo != nil && typeDef.Repo.Specs != nil {
		result.Specs = SubstituteAll(typeDef.Repo.Specs, moniker, root, moduleType)
	} else {
		result.Specs = []string{SpecsPattern(moniker)} // Generic default
	}

	// TestImpl - type default, then generic default
	if testImpl != "" {
		result.TestImpl = testImpl
	} else if typeDef != nil && typeDef.Repo != nil && typeDef.Repo.TestImpl != "" {
		result.TestImpl = SubstituteVariables(typeDef.Repo.TestImpl, moniker, root, moduleType)
	} else {
		result.TestImpl = TestImplPath(moniker) // Generic default
	}

	// Design - type default, then generic default
	if design != "" {
		result.Design = design
	} else if typeDef != nil && typeDef.Repo != nil && typeDef.Repo.Design != "" {
		result.Design = SubstituteVariables(typeDef.Repo.Design, moniker, root, moduleType)
	} else {
		result.Design = DesignPath(moniker) // Generic default
	}

	// Flags - type default, then generic default
	if catchAll != nil {
		result.CatchAll = *catchAll
	} else if typeDef != nil && typeDef.Flags != nil && typeDef.Flags.CatchAll != nil {
		result.CatchAll = *typeDef.Flags.CatchAll
	} else {
		result.CatchAll = FlagCatchAll
	}

	if ownChildren != nil {
		result.OwnChildren = *ownChildren
	} else if typeDef != nil && typeDef.Flags != nil && typeDef.Flags.OwnChildrenFiles != nil {
		result.OwnChildren = *typeDef.Flags.OwnChildrenFiles
	} else {
		result.OwnChildren = FlagOwnChildrenFiles
	}

	return result
}
