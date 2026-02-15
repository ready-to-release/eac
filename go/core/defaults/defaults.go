// Package defaults provides default values and path derivation for module domain.
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

// Module field defaults.
const (
	// Changelog is the default changelog filename.
	Changelog = "CHANGELOG.md"
)

// Path derivation functions - use these to derive default paths from moniker.
// All paths use forward slashes (/) for cross-platform config file compatibility.
// Note: TestImplPath is not provided here - it must come from repository.yml configuration.

// DesignPath returns the default design workspace path for a moniker.
// Pattern: specs/{moniker}/.design.
func DesignPath(moniker string) string {
	return "specs/" + moniker + "/.design"
}

// SpecsPath returns the default specs directory for a moniker.
// Pattern: specs/{moniker}.
func SpecsPath(moniker string) string {
	return "specs/" + moniker
}

// SpecsPattern returns the default specs glob pattern for a moniker.
// Pattern: specs/{moniker}/**.
func SpecsPattern(moniker string) string {
	return "specs/" + moniker + "/**"
}

// FeaturePath returns the default feature file path for a moniker and feature name.
// Pattern: specs/{moniker}/{feature}/{feature}.feature.
func FeaturePath(moniker, feature string) string {
	return "specs/" + moniker + "/" + feature + "/" + feature + ".feature"
}

// FeatureDir returns the default feature directory for a moniker and feature name.
// Pattern: specs/{moniker}/{feature}.
func FeatureDir(moniker, feature string) string {
	return "specs/" + moniker + "/" + feature
}

// SpecsComponentPath returns the spec directory for a module with component qualifier.
// Pattern: specs/{moniker}_{qualifier}.
func SpecsComponentPath(moniker, qualifier string) string {
	return "specs/" + moniker + "_" + qualifier
}

// WorkflowCIPath returns the default CI workflow path for a moniker.
// Pattern: .github/workflows/ci-{moniker}.yaml.
func WorkflowCIPath(moniker string) string {
	return ".github/workflows/ci-" + moniker + ".yaml"
}

// WorkflowReleasePath returns the default release workflow path for a moniker.
// Pattern: .github/workflows/release-{moniker}.yaml.
func WorkflowReleasePath(moniker string) string {
	return ".github/workflows/release-" + moniker + ".yaml"
}
