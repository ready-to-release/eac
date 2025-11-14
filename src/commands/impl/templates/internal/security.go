// File: src/commands/impl/templates/internal/security.go
// Security validation functions to prevent path traversal and command injection attacks
//
// Intent: Protect template system from malicious input by validating paths and URLs.
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear function names (ValidatePath, SanitizeGitURL)
//   - Explicit error messages with security context
//   - Simple validation logic with clear intent
//
// Easy to change:
//   - Pure functions with no side effects
//   - Validation logic separated from business logic
//   - Can add new validation rules without changing callers
//
// Hard to break:
//   - Comprehensive input validation
//   - Defense in depth (multiple checks)
//   - Clear error messages for debugging
//   - Table-driven tests cover all cases
package templates

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath ensures a user-provided path doesn't escape the designated base directory.
// This prevents path traversal attacks where malicious templates try to write files
// outside the intended output directory.
//
// Security Requirements:
//   - Reject absolute paths (e.g., /etc/passwd, C:\Windows)
//   - Reject paths starting with .. (e.g., ../../../etc/passwd)
//   - Reject paths that would resolve outside basePath after cleaning
//
// Returns error if path is unsafe, nil if safe.
func ValidatePath(basePath, userPath string) error {
	// Empty path is safe - it becomes basePath
	if userPath == "" {
		return nil
	}

	// Clean the path to resolve . and .. references
	cleaned := filepath.Clean(userPath)

	// SECURITY CHECK 1: Reject absolute paths
	// Absolute paths bypass the basePath entirely
	// Note: filepath.IsAbs behaves differently on Windows vs Unix
	// On Windows: C:\path is absolute, /path is relative
	// On Unix: /path is absolute
	// We reject both to be safe across platforms
	if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
		return fmt.Errorf("security: absolute paths not allowed: %s", userPath)
	}

	// SECURITY CHECK 2: Reject paths starting with ..
	// Even after cleaning, these try to escape the base directory
	if strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return fmt.Errorf("security: path traversal detected: %s", userPath)
	}

	// SECURITY CHECK 3: Verify final path stays within basePath
	// Join with basePath and check the result is still under basePath
	finalPath := filepath.Join(basePath, cleaned)

	// Get relative path from basePath to finalPath
	relToBase, err := filepath.Rel(basePath, finalPath)
	if err != nil {
		return fmt.Errorf("security: failed to validate path %s: %w", userPath, err)
	}

	// If relative path starts with .., it means finalPath is outside basePath
	if strings.HasPrefix(relToBase, ".."+string(filepath.Separator)) || relToBase == ".." {
		return fmt.Errorf("security: path escapes base directory: %s", userPath)
	}

	return nil
}

// SanitizeGitURL validates and sanitizes a Git repository URL to prevent command injection.
// Git commands are executed via exec.Command, so we must ensure the URL doesn't contain
// shell metacharacters that could be exploited.
//
// Security Requirements:
//   - Reject URLs containing shell metacharacters (; | & $ ` etc.)
//   - Reject URLs containing command substitution patterns ($() ${} ``)
//   - Reject URLs containing control characters (newline, null byte, etc.)
//   - Verify URL matches expected Git repository format
//
// Returns error if URL is unsafe or invalid, nil if safe.
func SanitizeGitURL(url string) error {
	// Validate URL is a Git repository format
	if !IsGitRepository(url) {
		return fmt.Errorf("invalid git repository URL: %s", url)
	}

	// SECURITY CHECK: Detect shell metacharacters and command injection patterns
	// These could be exploited in exec.Command if not properly escaped
	dangerous := []struct {
		char string
		desc string
	}{
		{";", "command separator"},
		{"|", "pipe operator"},
		{"&", "background operator"},
		{"$", "variable expansion"},
		{"`", "command substitution"},
		{"\n", "newline"},
		{"\r", "carriage return"},
		{"\x00", "null byte"},
		{"$(", "command substitution"},
		{"${", "variable expansion"},
	}

	for _, check := range dangerous {
		if strings.Contains(url, check.char) {
			return fmt.Errorf("security: invalid characters in URL (%s): %s", check.desc, url)
		}
	}

	return nil
}

// SecureFilePath combines ValidatePath with filepath.Join for safe file operations.
// Use this when constructing file paths from user input.
//
// Example:
//   outputPath, err := SecureFilePath(outputDir, userProvidedPath)
//   if err != nil {
//       return fmt.Errorf("invalid path: %w", err)
//   }
//   // Safe to use outputPath now
func SecureFilePath(basePath, userPath string) (string, error) {
	if err := ValidatePath(basePath, userPath); err != nil {
		return "", err
	}

	// Clean and join paths
	cleaned := filepath.Clean(userPath)
	result := filepath.Join(basePath, cleaned)

	return result, nil
}
