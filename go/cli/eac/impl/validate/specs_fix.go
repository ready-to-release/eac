package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain"
)

// ============================================================================
// Fix functionality for --fix flag
// ============================================================================

// FixResult holds the result of fixing a single file.
type FixResult struct {
	Path  string
	Fixes []FixedIssue
	Error error
}

// FixedIssue represents a single fix applied.
type FixedIssue struct {
	Line        int
	Code        string
	Description string
}

// FixCount returns the number of fixes applied.
func (r *FixResult) FixCount() int {
	return len(r.Fixes)
}

// fixGherkinFile attempts to fix issues in a feature file
// Supports fixing:
// - MISSING_VERIFICATION_TAG: adds @ov tag before scenario
// - INVALID_FEATURE_NAMING: renames feature to <module>_<kebab-name> format.
func fixGherkinFile(filePath string, errors []domain.ValidationError) (*FixResult, error) {
	result := &FixResult{
		Path:  filePath,
		Fixes: []FixedIssue{},
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file: %w", err)
		return result, result.Error
	}

	lines := strings.Split(string(content), "\n")
	modified := false

	// First pass: fix feature naming (do this first as it doesn't change line numbers)
	for _, e := range errors {
		if e.GetCode() == "INVALID_FEATURE_NAMING" && e.Line > 0 {
			idx := e.Line - 1
			if idx >= 0 && idx < len(lines) {
				newFeatureName := generateFeatureName(filePath, lines[idx])
				if newFeatureName != "" {
					oldLine := lines[idx]
					lines[idx] = "Feature: " + newFeatureName
					result.Fixes = append(result.Fixes, FixedIssue{
						Line:        e.Line,
						Code:        "INVALID_FEATURE_NAMING",
						Description: fmt.Sprintf("Renamed feature to '%s'", newFeatureName),
					})
					modified = true
					_ = oldLine // suppress unused warning
				}
			}
		}
	}

	// Second pass: collect lines needing @ov insertion
	linesToFix := []int{}
	for _, e := range errors {
		if e.GetCode() == "MISSING_VERIFICATION_TAG" && e.Line > 0 {
			linesToFix = append(linesToFix, e.Line)
		}
	}

	// Sort descending to insert from bottom up (preserves line numbers)
	if len(linesToFix) > 0 {
		sort.Sort(sort.Reverse(sort.IntSlice(linesToFix)))

		for _, lineNum := range linesToFix {
			idx := lineNum - 1 // 0-based index
			if idx >= 0 && idx < len(lines) {
				// Get indentation from scenario line
				indent := getIndentation(lines[idx])
				// Insert @ov tag before scenario
				newLine := indent + "@ov"
				lines = insertLine(lines, idx, newLine)
				result.Fixes = append(result.Fixes, FixedIssue{
					Line:        lineNum,
					Code:        "MISSING_VERIFICATION_TAG",
					Description: "Added @ov tag before Scenario",
				})
				modified = true
			}
		}
	}

	// Write back if modified
	if modified {
		err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
		if err != nil {
			result.Error = fmt.Errorf("failed to write file: %w", err)
			return result, result.Error
		}
	}

	return result, nil
}

// generateFeatureName creates a valid feature name from file path and current feature line
// Format: <module>_<kebab-case-description>.
func generateFeatureName(filePath, featureLine string) string {
	// Extract module from file path (e.g., specs/core/logging/spec.feature -> core)
	module := extractModuleFromPath(filePath)
	if module == "" {
		return ""
	}

	// Extract current feature name/description
	currentName := strings.TrimPrefix(strings.TrimSpace(featureLine), "Feature:")
	currentName = strings.TrimSpace(currentName)

	// Convert to kebab-case
	kebabName := toKebabCase(currentName)
	if kebabName == "" {
		return ""
	}

	return module + "_" + kebabName
}

// extractModuleFromPath extracts the module name from a spec file path
// e.g., specs/core/logging/specification.feature -> core
// e.g., specs/eac/commit/specification.feature -> eac.
func extractModuleFromPath(filePath string) string {
	// Normalize path separators
	normalized := filepath.ToSlash(filePath)

	// Find "specs/" in the path
	specsIdx := strings.Index(normalized, "specs/")
	if specsIdx == -1 {
		return ""
	}

	// Get the part after "specs/"
	afterSpecs := normalized[specsIdx+6:]

	// Split by "/" and get the first component (module name)
	parts := strings.Split(afterSpecs, "/")
	if len(parts) < 1 {
		return ""
	}

	return parts[0]
}

// toKebabCase converts a string to kebab-case
// e.g., "Dual-output logging with configurable routing" -> "dual-output-logging-with-configurable-routing".
func toKebabCase(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	prevHyphen := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
			prevHyphen = false
		} else if r == '-' && !prevHyphen && result.Len() > 0 {
			result.WriteRune('-')
			prevHyphen = true
		}
	}

	// Remove trailing hyphen
	resultStr := result.String()
	resultStr = strings.TrimSuffix(resultStr, "-")

	return resultStr
}

// getIndentation returns the leading whitespace from a line.
func getIndentation(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	return line[:len(line)-len(trimmed)]
}

// insertLine inserts a new line at the given index.
func insertLine(lines []string, idx int, newLine string) []string {
	lines = append(lines, "")
	copy(lines[idx+1:], lines[idx:])
	lines[idx] = newLine
	return lines
}

// formatFixResult formats a fix result for display.
func formatFixResult(result *FixResult, repoRoot string) string {
	if result.FixCount() == 0 {
		return ""
	}

	var output strings.Builder
	relPath := relativePath(result.Path, repoRoot)
	output.WriteString(fmt.Sprintf("🔧 Fixed %d issue(s) in %s:\n", result.FixCount(), relPath))

	// Show fixes in order (feature naming first, then tags in ascending line order)
	// First show non-tag fixes
	for _, fix := range result.Fixes {
		if fix.Code != "MISSING_VERIFICATION_TAG" {
			output.WriteString(fmt.Sprintf("   - Line %d: %s\n", fix.Line, fix.Description))
		}
	}

	// Then show tag fixes in reverse order (they were added bottom-up)
	for i := len(result.Fixes) - 1; i >= 0; i-- {
		fix := result.Fixes[i]
		if fix.Code == "MISSING_VERIFICATION_TAG" {
			output.WriteString(fmt.Sprintf("   - Line %d: %s\n", fix.Line, fix.Description))
		}
	}

	return output.String()
}
