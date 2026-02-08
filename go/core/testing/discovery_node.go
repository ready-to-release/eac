package testing

import (
	"os"
	"slices"
	"strings"
)

// parseNodeTestFile extracts describe blocks with tags from TypeScript/JavaScript test file.
// Pattern: describe('@L0 ComponentName', ...) or describe('@L0 @deps:foo ComponentName', ...)
// Tags must be at the start of the describe name, space-separated before the actual name.
// testFramework determines the Type field (e.g., "mocha", "jest").
// Note: Module dependency is attached by the caller (discoverTestsInFile) based on module contract.
func parseNodeTestFile(filePath, testFramework string) ([]TestReference, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	refs := []TestReference{}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for describe( patterns
		if !strings.Contains(trimmed, "describe(") {
			continue
		}

		// Extract the string inside describe('...' or describe("..."
		descName := extractDescribeName(trimmed)
		if descName == "" {
			continue
		}

		// Parse tags from the describe name
		// Format: "@L0 @deps:foo ComponentName" -> tags=[@L0, @deps:foo], name=ComponentName
		tags, testName := parseDescribeTags(descName)

		if len(tags) == 0 {
			// No tags in this describe, skip it (we only track tagged tests)
			continue
		}

		ref := TestReference{
			FilePath: filePath,
			Type:     testFramework,
			TestName: testName,
			Tags:     tags,
		}

		// Set execution control fields
		ref.IsIgnored, ref.SkipReason = extractSkipReason(ref.Tags)
		ref.IsManual = slices.Contains(ref.Tags, "@Manual")
		ref.IsSequential = slices.Contains(ref.Tags, "@sequential")
		ref.SystemDependencies = extractSystemDependencies(ref.Tags)
		ref.ModuleDependencies = extractModuleDependencies(ref.Tags)

		refs = append(refs, ref)
	}

	return refs, nil
}

// extractDescribeName extracts the string content from describe('...' or describe("...".
func extractDescribeName(line string) string {
	// Find describe( and then the opening quote
	idx := strings.Index(line, "describe(")
	if idx == -1 {
		return ""
	}

	rest := line[idx+len("describe("):]
	rest = strings.TrimSpace(rest)

	// Find the quote type (single or double)
	if len(rest) == 0 {
		return ""
	}

	quote := rest[0]
	if quote != '\'' && quote != '"' && quote != '`' {
		return ""
	}

	// Find the closing quote
	rest = rest[1:]
	endIdx := strings.IndexByte(rest, quote)
	if endIdx == -1 {
		return ""
	}

	return rest[:endIdx]
}

// parseDescribeTags parses tags from the start of a describe name.
// Input: "@L0 @deps:foo ComponentName"
// Output: tags=["@L0", "@deps:foo"], name="ComponentName".
func parseDescribeTags(descName string) ([]string, string) {
	parts := strings.Fields(descName)
	if len(parts) == 0 {
		return nil, ""
	}

	var tags []string
	var nameStart int

	for i, part := range parts {
		if strings.HasPrefix(part, "@") {
			tags = append(tags, part)
		} else {
			// First non-tag part starts the name
			nameStart = i
			break
		}
	}

	// If all parts were tags, no name found
	if nameStart == 0 && len(tags) == len(parts) {
		return tags, ""
	}

	// Reconstruct the name from remaining parts
	name := strings.Join(parts[nameStart:], " ")
	return tags, name
}
