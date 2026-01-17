package specs

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

// ExtractFeatureName extracts module and feature name from Gherkin content
// Expected format: "Feature: module_feature-name" or "Feature: feature-name"
//
// Security: Validates feature line for path traversal, path separators, and control characters.
func ExtractFeatureName(gherkin string) (string, string, error) {
	// Find Feature: line
	re := regexp.MustCompile(`(?m)^Feature:\s+(.+?)$`)
	matches := re.FindStringSubmatch(gherkin)
	if len(matches) < 2 {
		return "", "", fmt.Errorf("no Feature: declaration found in generated content")
	}

	featureLine := strings.TrimSpace(matches[1])

	// Security: Validate feature line before processing
	if err := ValidateFeatureLineSecurity(featureLine); err != nil {
		return "", "", fmt.Errorf("invalid feature name: %w", err)
	}

	var moduleName, featureName string

	// Check for module_feature format
	if strings.Contains(featureLine, "_") {
		parts := strings.SplitN(featureLine, "_", 2)
		if len(parts) == 2 {
			moduleName = parts[0]
			featureName = parts[1]
		}
	} else {
		// No module prefix
		featureName = featureLine
	}

	// Validate module name (if present)
	if moduleName != "" {
		if err := ValidateWindowsReservedName(moduleName); err != nil {
			return "", "", fmt.Errorf("invalid module name: %w", err)
		}
	}

	// Validate feature name
	if err := ValidateWindowsReservedName(featureName); err != nil {
		return "", "", fmt.Errorf("invalid feature name: %w", err)
	}

	return moduleName, featureName, nil
}

// ValidateFeatureLineSecurity validates feature line for security issues
//
// This function prevents:
// - Path traversal attacks (../)
// - Path separators (/ and \)
// - Control characters (except tab which is trimmed).
func ValidateFeatureLineSecurity(featureLine string) error {
	// Reject path traversal attempts first (before checking separators)
	if strings.Contains(featureLine, "..") {
		return fmt.Errorf("feature name contains path traversal attempt")
	}

	// Reject path separators
	if strings.ContainsAny(featureLine, "/\\") {
		return fmt.Errorf("feature name contains invalid path separator")
	}

	// Reject control characters (except tab which gets trimmed)
	for _, r := range featureLine {
		if r < 32 && r != '\t' {
			return fmt.Errorf("feature name contains control character (0x%02X)", r)
		}
	}

	return nil
}

// ValidateWindowsReservedName checks if a name is a Windows reserved device name
//
// Windows reserved names: CON, PRN, AUX, NUL, COM1-9, LPT1-9
// These names are reserved even with extensions (e.g., CON.txt).
func ValidateWindowsReservedName(name string) error {
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	// Extract name without extension if present
	baseName := name
	if idx := strings.Index(name, "."); idx != -1 {
		baseName = name[:idx]
	}

	upperName := strings.ToUpper(baseName)
	for _, reserved := range reservedNames {
		if upperName == reserved {
			return fmt.Errorf("'%s' is a Windows reserved device name", name)
		}
	}

	return nil
}

// ValidateOutputPath ensures the output path is within the repository (prevents path traversal attacks).
func ValidateOutputPath(outputPath, templateRoot string) error {
	// Clean both paths to resolve any .. or . components
	cleanOutput := filepath.Clean(outputPath)
	cleanRoot := filepath.Clean(templateRoot)

	// Convert to absolute paths for comparison
	absOutput := cleanOutput
	if !filepath.IsAbs(cleanOutput) {
		absOutput = filepath.Join(cleanRoot, cleanOutput)
		absOutput = filepath.Clean(absOutput)
	}

	absRoot := cleanRoot
	if !filepath.IsAbs(cleanRoot) {
		var err error
		absRoot, err = filepath.Abs(cleanRoot)
		if err != nil {
			return fmt.Errorf("failed to resolve repository root: %w", err)
		}
	}

	// Ensure output path is within repository
	// Use filepath.Rel to check if outputPath is under templateRoot
	rel, err := filepath.Rel(absRoot, absOutput)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// If the relative path starts with "..", it's trying to escape the repository
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return fmt.Errorf("output path must be within repository (attempted: %s)", outputPath)
	}

	return nil
}

// DetermineOutputPath constructs the output file path in specs directory.
func DetermineOutputPath(templateRoot, moduleName, featureName string, cfg *config.EACConfig) string {
	return cfg.Repository.SpecsFeaturePathAbs(templateRoot, moduleName, featureName)
}
