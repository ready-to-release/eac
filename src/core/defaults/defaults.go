// Package defaults provides default values and path derivation for module contracts.
//
// Use these when:
//   - Creating new specs/designs without an existing module contract
//   - A module doesn't have explicit values configured
//   - The loader needs to apply defaults during contract loading
//
// This package is the single source of truth for all default values.
//
// IMPORTANT: All paths use forward slashes (/) regardless of OS.
// Config files require forward slashes for cross-platform compatibility.
package defaults

// Module field defaults
const (
	// ModuleType is the default type when not specified in module contract.
	ModuleType = "no-module-type"

	// Parent is the default parent when not specified (root level).
	Parent = "."

	// Changelog is the default changelog filename.
	Changelog = "CHANGELOG.md"
)

// Flag defaults
const (
	// FlagCatchAll is the default for flags.catch_all.
	FlagCatchAll = false

	// FlagOwnChildrenFiles is the default for flags.own_children_files.
	FlagOwnChildrenFiles = false
)

// Path derivation functions - use these to derive default paths from moniker.
// All paths use forward slashes (/) for cross-platform config file compatibility.

// TestImplPath returns the default test implementation path for a moniker.
// Pattern: src/{moniker}/tests
func TestImplPath(moniker string) string {
	return "src/" + moniker + "/tests"
}

// DesignPath returns the default design workspace path for a moniker.
// Pattern: specs/{moniker}/.design
func DesignPath(moniker string) string {
	return "specs/" + moniker + "/.design"
}

// SpecsPath returns the default specs directory for a moniker.
// Pattern: specs/{moniker}
func SpecsPath(moniker string) string {
	return "specs/" + moniker
}

// SpecsPattern returns the default specs glob pattern for a moniker.
// Pattern: specs/{moniker}/**
func SpecsPattern(moniker string) string {
	return "specs/" + moniker + "/**"
}

// FeaturePath returns the default feature file path for a moniker and feature name.
// Pattern: specs/{moniker}/{feature}/{feature}.feature
func FeaturePath(moniker, feature string) string {
	return "specs/" + moniker + "/" + feature + "/" + feature + ".feature"
}

// FeatureDir returns the default feature directory for a moniker and feature name.
// Pattern: specs/{moniker}/{feature}
func FeatureDir(moniker, feature string) string {
	return "specs/" + moniker + "/" + feature
}

// WorkflowCIPath returns the default CI workflow path for a moniker.
// Pattern: .github/workflows/ci-{moniker}.yaml
func WorkflowCIPath(moniker string) string {
	return ".github/workflows/ci-" + moniker + ".yaml"
}

// WorkflowReleasePath returns the default release workflow path for a moniker.
// Pattern: .github/workflows/release-{moniker}.yml
func WorkflowReleasePath(moniker string) string {
	return ".github/workflows/release-" + moniker + ".yml"
}
