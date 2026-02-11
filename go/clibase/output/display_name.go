package output

import (
	"fmt"
	"strings"
)

// PackageDisplayName extracts the display name from a package path.
// For BDD tests with format "featureName:implPath:featurePath", returns "module/featureName"
// to ensure uniqueness (e.g., "vscode-commit/progress-buffer").
// For unit tests with format "path", returns "path" unchanged.
func PackageDisplayName(pkgPath string) string {
	parts := strings.Split(pkgPath, ":")
	if len(parts) >= 2 {
		specName := parts[0]
		implPath := parts[1]
		// Extract module name from implPath (last component)
		// e.g., "typescript/vscode-commit" -> "vscode-commit"
		// e.g., "go/eac/specs/impl/eac" -> "eac"
		implParts := strings.Split(implPath, "/")
		moduleName := implParts[len(implParts)-1]
		return moduleName + "/" + specName
	}
	// Unit test or other format: return as-is
	return pkgPath
}

// PackageDisplayNames converts a list of package paths to display names.
func PackageDisplayNames(pkgPaths []string) []string {
	result := make([]string, len(pkgPaths))
	for i, path := range pkgPaths {
		result[i] = PackageDisplayName(path)
	}
	return result
}

// FormatComponentName creates the "module:component" display name.
// This is the standard format for displaying component names in build/test/scan results.
func FormatComponentName(module, component string) string {
	return fmt.Sprintf("%s:%s", module, component)
}

// TruncateComponentName truncates a component name to fit display width.
// It prefers truncating the module name while preserving the component name.
// If maxWidth is too small, truncates the whole string.
func TruncateComponentName(module, component string, maxWidth int) string {
	full := FormatComponentName(module, component)
	if len(full) <= maxWidth {
		return full
	}

	// Reserve space for colon and at least 8 chars of component
	minCompLen := 8
	if len(component) < minCompLen {
		minCompLen = len(component)
	}
	maxModLen := maxWidth - minCompLen - 1 // -1 for colon

	if maxModLen < 3 {
		// Just truncate the whole thing
		if maxWidth <= 3 {
			return full[:maxWidth]
		}
		return full[:maxWidth-3] + "..."
	}

	truncMod := module
	if len(module) > maxModLen {
		truncMod = module[:maxModLen-3] + "..."
	}

	return FormatComponentName(truncMod, component)
}
