// Package design provides input validation utilities
package design

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateModuleName validates that a module name is safe and doesn't contain path traversal
func ValidateModuleName(module string) error {
	if module == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	if len(module) > 255 {
		return fmt.Errorf("module name too long (max 255 characters): %d", len(module))
	}

	// Check for path traversal attempts
	if strings.Contains(module, "..") {
		return fmt.Errorf("module name cannot contain '..': %s", module)
	}

	// Check for absolute paths (both OS-specific and Unix-style)
	if filepath.IsAbs(module) || strings.HasPrefix(module, "/") {
		return fmt.Errorf("module name cannot be an absolute path: %s", module)
	}

	// Check for invalid characters (OS-specific path separators, special chars)
	invalidChars := []string{"\x00", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(module, char) {
			return fmt.Errorf("module name contains invalid character '%s': %s", char, module)
		}
	}

	// Check for reserved names on Windows
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperModule := strings.ToUpper(module)
	for _, res := range reserved {
		if upperModule == res {
			return fmt.Errorf("module name is a reserved Windows name: %s", module)
		}
	}

	return nil
}

// CleanModuleName removes common prefixes and suffixes from module names
func CleanModuleName(module string) string {
	// Remove specs/ prefix (both Unix and Windows)
	module = strings.TrimPrefix(module, SpecsDirectory+"/")
	module = strings.TrimPrefix(module, SpecsDirectory+"\\")

	// Remove /design suffix (both Unix and Windows)
	module = strings.TrimSuffix(module, "/"+DesignDirectory)
	module = strings.TrimSuffix(module, "\\"+DesignDirectory)

	// Remove src/ prefix (both Unix and Windows)
	module = strings.TrimPrefix(module, SourceDirectory+"/")
	module = strings.TrimPrefix(module, SourceDirectory+"\\")

	return module
}

// ValidIdentifierPattern matches valid DSL identifiers
var ValidIdentifierPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// ValidateIdentifier checks if a string is a valid DSL identifier
func ValidateIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	if !ValidIdentifierPattern.MatchString(identifier) {
		return fmt.Errorf("invalid identifier '%s': must start with letter and contain only alphanumeric, underscore, or dash", identifier)
	}

	return nil
}
