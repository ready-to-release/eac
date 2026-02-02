package test

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/testing"
)

// writeln writes a formatted string with platform-specific line ending to the writer.
func writeln(w io.Writer, format string, args ...interface{}) {
	output.Writeln(w, format, args...)
}

// extractModuleFromPath extracts the module moniker from a test file path
// Handles go/eac/<module>/..., go/r2r/<module>/..., and specs/<module>/... formats
// Supports both absolute and relative paths.
func extractModuleFromPath(filePath string) string {
	// Normalize path separators to forward slashes
	normalizedPath := filepath.ToSlash(filePath)

	// Special case: go/eac/specs/impl/<module>/... or go/r2r/specs/impl/<module>/...
	// These are test implementations that belong to <module>, not eac-specs/r2r-specs
	// Must check this BEFORE the general /go/eac/ or /go/r2r/ boundary check
	for _, implPattern := range []string{"/go/eac/specs/impl/", "/go/r2r/specs/impl/", "go/eac/specs/impl/", "go/r2r/specs/impl/"} {
		idx := strings.Index(normalizedPath, implPattern)
		if idx >= 0 {
			// Extract module name after impl/
			relativePath := normalizedPath[idx+len(implPattern):]
			parts := strings.Split(relativePath, "/")
			if len(parts) >= 1 && parts[0] != "" {
				return parts[0] // Return module name directly (e.g., "core", "r2r-cli")
			}
		}
	}

	// Find "/go/eac/" in the path (handles both absolute and relative paths)
	// For paths like /project/go/cli/eac/..., extract "eac-cli"
	for _, boundary := range []string{"/go/eac/", "/go/r2r/"} {
		idx := strings.Index(normalizedPath, boundary)
		if idx >= 0 {
			// Extract module name from path after boundary
			relativePath := normalizedPath[idx+len(boundary):]
			parts := strings.Split(relativePath, "/")
			if len(parts) >= 1 && parts[0] != "" {
				// Return as eac-<part1> or r2r-<part1>
				prefix := "eac"
				if boundary == "/go/r2r/" {
					prefix = "r2r"
				}
				return prefix + "-" + parts[0]
			}
		}
	}

	// Check if path contains "/specs/" (handles specs/eac-*, specs/github, specs/repository, etc.)
	specsIndex := strings.Index(normalizedPath, "/specs/")
	if specsIndex >= 0 {
		// Extract from "/specs/" onwards
		relativePath := normalizedPath[specsIndex+1:]
		// Format: specs/<module>/...
		parts := strings.Split(strings.TrimPrefix(relativePath, "specs/"), "/")
		if len(parts) >= 1 && parts[0] != "" {
			// Return the first part (e.g., "eac-cli", "github", "repository")
			return parts[0]
		}
	}

	// Also check for paths starting with "go/eac/" or "go/r2r/" (relative paths)
	for _, prefix := range []string{"go/eac/", "go/r2r/"} {
		if strings.HasPrefix(normalizedPath, prefix) {
			parts := strings.Split(normalizedPath[len(prefix):], "/")
			if len(parts) >= 1 && parts[0] != "" {
				monikerPrefix := "eac"
				if prefix == "go/r2r/" {
					monikerPrefix = "r2r"
				}
				return monikerPrefix + "-" + parts[0]
			}
		}
	}

	if strings.HasPrefix(normalizedPath, "specs/") {
		parts := strings.Split(strings.TrimPrefix(normalizedPath, "specs/"), "/")
		if len(parts) >= 1 && parts[0] != "" {
			return parts[0]
		}
	}

	// Fallback: return empty string
	return ""
}

// mapGOOSToDepTag maps runtime.GOOS values to dependency tag names.
func mapGOOSToDepTag(goos string) string {
	switch goos {
	case "linux":
		return "linux"
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return goos
	}
}

// filterByOSCompatibility filters tests based on OS-specific dependencies
// Tests with deps:linux only run on Linux, deps:macos only on macOS, deps:windows only on Windows
// Tests without any OS-specific deps run everywhere (OS-agnostic by default)
// Tests with multiple OS deps (e.g., deps:linux AND deps:macos) run on any of those OSes.
func filterByOSCompatibility(tests []testing.TestReference, _ io.Writer) []testing.TestReference {
	currentOS := mapGOOSToDepTag(runtime.GOOS)
	compatible := []testing.TestReference{}

	for i := range tests {
		test := &tests[i]
		// Check if test has any OS-specific dependencies
		hasOSDep := false
		matchesCurrentOS := false

		for _, dep := range test.SystemDependencies {
			// Check if this is an OS dependency
			if testing.IsOSPlatformDep(dep) {
				hasOSDep = true
				if dep == currentOS {
					matchesCurrentOS = true
				}
			}
		}

		// Include test if:
		// 1. It has no OS-specific deps (runs on any OS), OR
		// 2. It has an OS dep that matches the current OS
		if !hasOSDep || matchesCurrentOS {
			compatible = append(compatible, *test)
		}
	}

	return compatible
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readLastLines reads the last N lines from a file, parsing JSON test output.
func readLastLines(filePath string, n int) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{"Error reading file: " + err.Error()}
	}
	defer file.Close()

	var outputLines []string
	scanner := bufio.NewScanner(file)

	// Parse JSON test output and extract "Output" fields
	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse as JSON test event
		var event struct {
			Action string `json:"Action"`
			Output string `json:"Output"`
		}

		if err := json.Unmarshal([]byte(line), &event); err == nil {
			// Successfully parsed JSON - extract Output field if it's an output event
			if event.Action == "output" && event.Output != "" {
				// Strip ANSI color codes for cleaner display
				cleaned := stripANSI(event.Output)
				// Trim trailing newlines but keep the content
				cleaned = strings.TrimRight(cleaned, "\n")
				if cleaned != "" {
					outputLines = append(outputLines, cleaned)
				}
			}
		} else {
			// Not JSON - include raw line
			outputLines = append(outputLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return []string{"Error scanning file: " + err.Error()}
	}

	// Return last n lines
	if len(outputLines) <= n {
		return outputLines
	}
	return outputLines[len(outputLines)-n:]
}

// stripANSI removes ANSI color codes from a string.
func stripANSI(str string) string {
	// Remove ANSI escape sequences like [33m, [0m, etc.
	result := ""
	inEscape := false
	for i := 0; i < len(str); i++ {
		if str[i] == '\x1b' && i+1 < len(str) && str[i+1] == '[' {
			inEscape = true
			i++ // Skip the '['
			continue
		}
		if inEscape {
			if (str[i] >= 'A' && str[i] <= 'Z') || (str[i] >= 'a' && str[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result += string(str[i])
	}
	return result
}

// getPackageTestType determines the test type for a package.
// If all tests have the same type, returns that type.
// If mixed or empty, returns empty string (use default Go handler).
func getPackageTestType(tests []testing.TestReference) string {
	if len(tests) == 0 {
		return ""
	}

	firstType := tests[0].Type
	for i := 1; i < len(tests); i++ {
		if tests[i].Type != firstType {
			return "" // Mixed types, use default
		}
	}

	return firstType
}

// convertToCucumberTagExpr converts godog tag format to cucumber-js tag expression.
// Godog: "@L0,@L1 && ~@skip:wip" → Cucumber: "(@L0 or @L1) and not @skip:wip".
func convertToCucumberTagExpr(godogTags string) string {
	if godogTags == "" {
		return ""
	}

	var parts []string
	// Split by && for AND conditions
	andParts := strings.Split(godogTags, " && ")
	for _, part := range andParts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "~") {
			// Negation: ~@tag → not @tag
			parts = append(parts, "not "+strings.TrimPrefix(part, "~"))
		} else if strings.Contains(part, ",") {
			// OR: @L0,@L1 → (@L0 or @L1)
			orTags := strings.Split(part, ",")
			parts = append(parts, "("+strings.Join(orTags, " or ")+")")
		} else {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " and ")
}
