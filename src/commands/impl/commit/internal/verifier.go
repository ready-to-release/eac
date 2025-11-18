// Package commitmessage provides commit message validation against contract
package commitmessage

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/src/core/ai/contract"
	"gopkg.in/yaml.v3"
)

// Precompiled regular expressions for performance
var (
	conventionalCommitRegex = regexp.MustCompile(`^(feat|fix|refactor|docs|chore|test|perf|style)\(([a-z0-9\-]+|multi-module)\):\s*(.+)$`)
	moduleSubjectLineRegex  = regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)$`)
)

// ValidationError is an alias to the core contract ValidationError
type ValidationError = contract.ValidationError

// CommitMessageContract represents the structure.yml contract
type CommitMessageContract struct {
	Version     string `yaml:"version"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Structure   []struct {
		Section   string `yaml:"section"`
		Required  bool   `yaml:"required"`
		Format    string `yaml:"format"`
		MaxLength int    `yaml:"max_length,omitempty"`
	} `yaml:"structure"`
	SemanticTypes     []string         `yaml:"semantic_types"`
	SubjectLineFormat string           `yaml:"subject_line_format"`
	Constraints       map[string]any   `yaml:"constraints"`
	MarkdownRules     []map[string]any `yaml:"markdown_rules"`
	AntiCorruption    map[string]any   `yaml:"anti_corruption"`
}

// AntiCorruptionRules is an alias to the core contract AntiCorruptionRules
type AntiCorruptionRules = contract.AntiCorruptionRules

// LoadAntiCorruptionRules loads anti-corruption rules using the core framework
func LoadAntiCorruptionRules(rulesPath string) (*AntiCorruptionRules, error) {
	// Extract directory and version from path
	// rulesPath format: "workspace/contracts/commit-message/0.1.0/anti-corruption.yml"
	// We need to extract workspace root, contract dir, and version

	// For now, read directly - this is backward compatible
	data, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read anti-corruption rules file: %w", err)
	}

	var rules AntiCorruptionRules
	var rawData map[string]interface{}

	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse anti-corruption rules YAML: %w", err)
	}

	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse anti-corruption rules YAML into map: %w", err)
	}

	rules.RawData = rawData
	return &rules, nil
}

// LoadContract loads and parses the structure.yml file
func LoadContract(contractPath string) (*CommitMessageContract, error) {
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file: %w", err)
	}

	var contract CommitMessageContract
	if err := yaml.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("failed to parse contract YAML: %w", err)
	}

	return &contract, nil
}

// VerifyContractImplementation validates that verifier implements all contract rules
func VerifyContractImplementation(contractPath string) []ValidationError {
	var errors []ValidationError

	contract, err := LoadContract(contractPath)
	if err != nil {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_LOAD_ERROR",
			Message:  err.Error(),
			Severity: "error",
		})
		return errors
	}

	// Verify version matches
	if contract.Version != ContractVersion {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_VERSION_MISMATCH",
			Message:  fmt.Sprintf("Expected version %s, got %s", ContractVersion, contract.Version),
			Severity: "error",
		})
	}

	// Verify required structure sections
	requiredSections := map[string]bool{
		"top_level_heading": false,
		"top_level_body":    false,
		"module_sections":   false,
	}

	for _, section := range contract.Structure {
		if section.Required {
			if _, exists := requiredSections[section.Section]; exists {
				requiredSections[section.Section] = true
			}
		}

		// Verify top_level_heading max_length
		if section.Section == "top_level_heading" {
			if section.MaxLength != MaxHeaderLength {
				errors = append(errors, ValidationError{
					Code:     "CONTRACT_CONSTRAINT_MISMATCH",
					Message:  fmt.Sprintf("top_level_heading max_length should be %d, got %d", MaxHeaderLength, section.MaxLength),
					Severity: "error",
				})
			}
		}
	}

	for section, found := range requiredSections {
		if !found {
			errors = append(errors, ValidationError{
				Code:     "CONTRACT_MISSING_SECTION",
				Message:  fmt.Sprintf("Contract missing required section: %s", section),
				Severity: "error",
			})
		}
	}

	// Verify semantic types
	expectedTypes := map[string]bool{
		"feat": true, "fix": true, "refactor": true, "docs": true,
		"chore": true, "test": true, "perf": true, "style": true,
	}
	for _, t := range contract.SemanticTypes {
		if !expectedTypes[t] {
			errors = append(errors, ValidationError{
				Code:     "CONTRACT_UNKNOWN_TYPE",
				Message:  fmt.Sprintf("Unknown semantic type in contract: %s", t),
				Severity: "warning",
			})
		}
		delete(expectedTypes, t)
	}
	for missingType := range expectedTypes {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_MISSING_TYPE",
			Message:  fmt.Sprintf("Contract missing semantic type: %s", missingType),
			Severity: "error",
		})
	}

	// Verify subject line format
	if contract.SubjectLineFormat != "<module>: <type>: <description>" {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_FORMAT_MISMATCH",
			Message:  fmt.Sprintf("subject_line_format mismatch: %s", contract.SubjectLineFormat),
			Severity: "error",
		})
	}

	// Verify constraints
	requiredConstraints := map[string]any{
		"max_line_length":         MaxLineLength,
		"no_trailing_periods":     true,
		"code_blocks_closed":      true,
		"module_header_no_colons": true,
	}

	for key, expectedVal := range requiredConstraints {
		actualVal, exists := contract.Constraints[key]
		if !exists {
			errors = append(errors, ValidationError{
				Code:     "CONTRACT_MISSING_CONSTRAINT",
				Message:  fmt.Sprintf("Contract missing constraint: %s", key),
				Severity: "error",
			})
		} else if actualVal != expectedVal {
			errors = append(errors, ValidationError{
				Code:     "CONTRACT_CONSTRAINT_VALUE",
				Message:  fmt.Sprintf("Constraint %s should be %v, got %v", key, expectedVal, actualVal),
				Severity: "error",
			})
		}
	}

	// Verify markdown rules exist
	if len(contract.MarkdownRules) == 0 {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_MISSING_MARKDOWN_RULES",
			Message:  "Contract should define markdown_rules",
			Severity: "warning",
		})
	}

	return errors
}

// VerifyCommitMessageContract validates a commit message against contracts/commit-message/0.1.0/structure.yml
// affectedModules is the list of modules that had staged changes
func VerifyCommitMessageContract(commitMessage string, affectedModules []string) []ValidationError {
	var errors []ValidationError

	lines := strings.Split(commitMessage, "\n")
	if len(lines) == 0 {
		errors = append(errors, ValidationError{
			Code:     "EMPTY_MESSAGE",
			Message:  "Commit message is empty",
			Severity: "error",
		})
		return errors
	}

	// RULE 1: First line must be conventional commit header with scope
	// Format: <type>(<scope>): <summary>
	if !conventionalCommitRegex.MatchString(lines[0]) {
		errors = append(errors, ValidationError{
			Code:     "INVALID_HEADER_FORMAT",
			Message:  "Header must follow format: <type>(<scope>): <summary> (e.g., feat(cli): add new command)",
			Line:     1,
			Severity: "error",
		})
	}

	// RULE 2: Header max length
	if len(lines[0]) > MaxHeaderLength {
		errors = append(errors, ValidationError{
			Code:     "HEADER_TOO_LONG",
			Message:  fmt.Sprintf("Header exceeds %d characters (%d chars)", MaxHeaderLength, len(lines[0])),
			Line:     1,
			Severity: "error",
		})
	}

	// RULE 3: No trailing period (except ellipsis "...")
	if strings.HasSuffix(lines[0], ".") && !strings.HasSuffix(lines[0], "...") {
		errors = append(errors, ValidationError{
			Code:     "HEADER_TRAILING_PERIOD",
			Message:  "Header must not end with period",
			Line:     1,
			Severity: "error",
		})
	}

	// RULE 4: Check for Auditor-Summary field (should appear early, after blank line)
	hasAuditorSummary := false
	for i := 1; i < len(lines) && i < 10; i++ { // Check first 10 lines
		if strings.HasPrefix(lines[i], "Auditor-Summary:") {
			hasAuditorSummary = true
			break
		}
	}
	if !hasAuditorSummary {
		errors = append(errors, ValidationError{
			Code:     "MISSING_AUDITOR_SUMMARY",
			Message:  "Missing Auditor-Summary field after header",
			Severity: "error",
		})
	}

	// RULE 5: Check for top-level body (should appear after header and auditor summary, before module sections)
	hasTopLevelBody := false
	hasModuleSection := false
	foundModules := make(map[string]bool) // Track which modules we found in the commit message
	afterAuditorSummary := false
	inModuleSection := false

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Track when we pass the Auditor-Summary line
		if strings.HasPrefix(trimmed, "Auditor-Summary:") {
			afterAuditorSummary = true
			continue
		}

		// Check if we have body text after Auditor-Summary and before module sections
		if afterAuditorSummary && !hasModuleSection && trimmed != "" &&
			!strings.HasPrefix(trimmed, "Auditor-Summary:") &&
			!strings.HasPrefix(trimmed, "Changes:") &&
			trimmed != "---" {
			hasTopLevelBody = true
		}

		// Module sections: look for plain module name followed by dashes
		// Pattern: module-name on one line, dashes on next line
		if i < len(lines)-1 && !inModuleSection {
			nextLine := strings.TrimSpace(lines[i+1])
			// Check if current line is a module name and next line is dashes
			if isModuleName(trimmed) && isDashesLine(nextLine) {
				hasModuleSection = true
				inModuleSection = true
				foundModules[trimmed] = true
			}
		}

		// Exit module section when we hit horizontal rule or blank lines
		if trimmed == "---" || (inModuleSection && trimmed == "" && i < len(lines)-1) {
			nextTrimmed := ""
			if i < len(lines)-1 {
				nextTrimmed = strings.TrimSpace(lines[i+1])
			}
			// If next line is also blank or dashes, we're between sections
			if nextTrimmed == "" || nextTrimmed == "---" || isDashesLine(nextTrimmed) {
				inModuleSection = false
			}
		}

		// RULE 7: Line length in body text (skip special lines: module headers with dashes,
		// Auditor-Summary, Changes, tables, code blocks, horizontal rules)
		if trimmed != "" &&
			!strings.HasPrefix(trimmed, "Auditor-Summary:") &&
			!strings.HasPrefix(trimmed, "Changes:") &&
			!strings.HasPrefix(trimmed, "|") &&
			!strings.HasPrefix(trimmed, "```") &&
			!isDashesLine(trimmed) &&
			trimmed != "---" &&
			!strings.HasPrefix(trimmed, "Agent:") {
			if len(trimmed) > MaxLineLength {
				preview := trimmed
				if len(preview) > 60 {
					preview = preview[:57] + "..."
				}
				errors = append(errors, ValidationError{
					Code:     "LINE_TOO_LONG",
					Message:  fmt.Sprintf("Line exceeds %d characters (%d chars): %s", MaxLineLength, len(trimmed), preview),
					Line:     lineNum,
					Severity: "warning",
				})
			}
		}
	}

	if !hasTopLevelBody {
		errors = append(errors, ValidationError{
			Code:     "MISSING_TOP_LEVEL_BODY",
			Message:  "Missing top-level body text after title (before module sections)",
			Severity: "error",
		})
	}

	// Check if we're in a multi-module commit (more than 1 affected module)
	if len(affectedModules) > 1 {
		// Multi-module commits MUST have module sections
		if !hasModuleSection {
			moduleList := strings.Join(affectedModules, ", ")
			errors = append(errors, ValidationError{
				Code:     "MISSING_MODULE_SECTION",
				Message:  fmt.Sprintf("Multi-module commit missing module sections. Expected: %s", moduleList),
				Severity: "error",
			})
		} else {
			// Check which specific modules are missing
			var missingModules []string
			for _, expectedModule := range affectedModules {
				if !foundModules[expectedModule] {
					missingModules = append(missingModules, expectedModule)
				}
			}

			if len(missingModules) > 0 {
				moduleList := strings.Join(missingModules, ", ")
				errors = append(errors, ValidationError{
					Code:     "MISSING_MODULE_SECTION",
					Message:  fmt.Sprintf("Missing module sections for: %s", moduleList),
					Severity: "error",
				})
			}
		}
	}
	// Single-module commits don't require module sections

	// RULE 8: Validate module subject lines
	errors = append(errors, validateModuleSubjectLines(lines)...)

	// RULE 8b: Validate module section structure (plain text format)
	errors = append(errors, validateModuleSectionStructure(lines)...)

	// RULE 9: Check for unclosed code blocks
	errors = append(errors, validateCodeBlocks(lines)...)

	return errors
}

// validateModuleSubjectLines checks that module sections have proper subject lines
func validateModuleSubjectLines(lines []string) []ValidationError {
	var errors []ValidationError

	inModuleSection := false
	currentModule := ""
	foundSubjectLine := false

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Detect module section header (new format: module name + dashes)
		if isDashesLine(trimmed) && i > 0 {
			// Check if previous non-empty line is a module name
			prevLine := ""
			for j := i - 1; j >= 0; j-- {
				prevTrimmed := strings.TrimSpace(lines[j])
				if prevTrimmed != "" {
					prevLine = prevTrimmed
					break
				}
			}

			if isModuleName(prevLine) {
				// If we were in a module section and didn't find subject line
				if inModuleSection && !foundSubjectLine {
					errors = append(errors, ValidationError{
						Code:     "MISSING_SUBJECT_LINE",
						Message:  fmt.Sprintf("Module '%s' missing subject line", currentModule),
						Severity: "error",
					})
				}

				inModuleSection = true
				currentModule = prevLine
				foundSubjectLine = false
				continue
			}
		}

		// If in module section, look for subject line
		if inModuleSection && !foundSubjectLine && trimmed != "" {
			// Skip blank lines
			if trimmed == "" {
				continue
			}

			// This should be the subject line
			if !moduleSubjectLineRegex.MatchString(trimmed) {
				errors = append(errors, ValidationError{
					Code:     "INVALID_SUBJECT_FORMAT",
					Message:  fmt.Sprintf("Subject line does not follow '<module>: <type>: <description>' format: %s", trimmed),
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				// Validate subject line length
				if len(trimmed) > MaxSubjectLength {
					errors = append(errors, ValidationError{
						Code:     "SUBJECT_TOO_LONG",
						Message:  fmt.Sprintf("Subject line exceeds %d characters (%d chars)", MaxSubjectLength, len(trimmed)),
						Line:     lineNum,
						Severity: "error",
					})
				}

				// Check no trailing period (except ellipsis "...")
				if strings.HasSuffix(trimmed, ".") && !strings.HasSuffix(trimmed, "...") {
					errors = append(errors, ValidationError{
						Code:     "SUBJECT_TRAILING_PERIOD",
						Message:  "Subject line must not end with period",
						Line:     lineNum,
						Severity: "error",
					})
				}
			}

			foundSubjectLine = true
		}

		// Exit module section when we hit horizontal rule
		if trimmed == "---" {
			if inModuleSection && !foundSubjectLine {
				errors = append(errors, ValidationError{
					Code:     "MISSING_SUBJECT_LINE",
					Message:  fmt.Sprintf("Module '%s' missing subject line", currentModule),
					Severity: "error",
				})
			}
			inModuleSection = false
		}
	}

	return errors
}

// validateCodeBlocks ensures all code blocks are properly closed
func validateCodeBlocks(lines []string) []ValidationError {
	var errors []ValidationError

	codeBlockOpen := false
	codeBlockLine := 0

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if codeBlockOpen {
				// Closing block
				codeBlockOpen = false
			} else {
				// Opening block
				codeBlockOpen = true
				codeBlockLine = lineNum
			}
		}
	}

	if codeBlockOpen {
		errors = append(errors, ValidationError{
			Code:     "UNCLOSED_CODE_BLOCK",
			Message:  fmt.Sprintf("Code block opened at line %d is not closed", codeBlockLine),
			Severity: "error",
		})
	}

	return errors
}

// validateModuleSectionStructure validates that module sections have proper structure
// Format: <module-name>\n<dashes>\n<module>: <type>: <description>\n<body>
func validateModuleSectionStructure(lines []string) []ValidationError {
	var errors []ValidationError

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// Check for orphaned dashes line (dashes without module name before it)
		// Skip section separators (exactly 3 dashes) as they separate modules
		if isDashesLine(trimmed) && !isSectionSeparator(trimmed) {
			// Look back to see if previous non-empty line was a module name
			prevNonEmpty := ""
			for j := i - 1; j >= 0; j-- {
				if strings.TrimSpace(lines[j]) != "" {
					prevNonEmpty = strings.TrimSpace(lines[j])
					break
				}
			}

			// If previous line wasn't a valid module name, this is an orphaned dashes line
			if !isModuleName(prevNonEmpty) {
				errors = append(errors, ValidationError{
					Code:     "ORPHANED_DASHES_LINE",
					Message:  fmt.Sprintf("Orphaned dashes line at line %d - must be preceded by module name", i+1),
					Line:     i + 1,
					Severity: "error",
				})
			}
		}

		// Check for module subject line without proper header structure
		if moduleSubjectLineRegex.MatchString(trimmed) {
			// This is a subject line - check if it has proper module structure before it
			moduleName := moduleSubjectLineRegex.FindStringSubmatch(trimmed)[1]

			// Look back for module name + dashes
			foundModuleName := false
			foundDashes := false

			for j := i - 1; j >= 0 && j >= i-5; j-- { // Look back up to 5 lines
				prevTrimmed := strings.TrimSpace(lines[j])
				if prevTrimmed == "" {
					continue // Skip blank lines
				}

				if isDashesLine(prevTrimmed) {
					foundDashes = true
				} else if isModuleName(prevTrimmed) && prevTrimmed == moduleName {
					foundModuleName = true
				} else {
					// Hit some other content, stop searching
					break
				}
			}

			// Subject line should have module name and dashes before it
			if !foundModuleName || !foundDashes {
				if !foundModuleName && !foundDashes {
					errors = append(errors, ValidationError{
						Code:     "MALFORMED_MODULE_SECTION",
						Message:  fmt.Sprintf("Module section at line %d missing module name and dashes header", i+1),
						Line:     i + 1,
						Severity: "error",
					})
				} else if !foundModuleName {
					errors = append(errors, ValidationError{
						Code:     "MISSING_MODULE_NAME",
						Message:  fmt.Sprintf("Module section at line %d missing module name (has dashes but no name)", i+1),
						Line:     i + 1,
						Severity: "error",
					})
				} else if !foundDashes {
					errors = append(errors, ValidationError{
						Code:     "MISSING_MODULE_DASHES",
						Message:  fmt.Sprintf("Module section at line %d missing dashes separator (has name but no dashes)", i+1),
						Line:     i + 1,
						Severity: "error",
					})
				}
			}
		}
	}

	return errors
}

// isModuleName checks if a string looks like a module name
// Module names are lowercase alphanumeric with dashes or underscores
func isModuleName(s string) bool {
	if s == "" || len(s) > MaxModuleNameLength {
		return false
	}

	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return false
		}
	}

	return true
}

// isDashesLine checks if a line consists only of dashes
func isDashesLine(s string) bool {
	if len(s) < MinDashesLength {
		return false
	}

	for _, ch := range s {
		if ch != '-' {
			return false
		}
	}

	return true
}

// isSectionSeparator checks if a line is exactly a section separator (---)
// Section separators are used to divide module sections and should not be
// validated as module header underlines
func isSectionSeparator(s string) bool {
	return s == "---"
}
