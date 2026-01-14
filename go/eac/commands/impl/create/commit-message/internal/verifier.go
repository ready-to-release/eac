// Package commitmessage provides commit message validation against contract
package commitmessage

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

// Lazy-loaded regular expressions for performance (compiled only when needed)
var (
	conventionalCommitRegex     *regexp.Regexp
	moduleSubjectLineRegex      *regexp.Regexp
	conventionalCommitRegexOnce sync.Once
	moduleSubjectLineRegexOnce  sync.Once
)

// getConventionalCommitRegex returns the conventional commit regex, compiling it only once
func getConventionalCommitRegex() *regexp.Regexp {
	conventionalCommitRegexOnce.Do(func() {
		conventionalCommitRegex = regexp.MustCompile(`^(feat|fix|refactor|docs|chore|test|perf|style)\(([a-z0-9\-]+|multi-module)\):\s*(.+)$`)
	})
	return conventionalCommitRegex
}

// getModuleSubjectLineRegex returns the module subject line regex, compiling it only once
func getModuleSubjectLineRegex() *regexp.Regexp {
	moduleSubjectLineRegexOnce.Do(func() {
		moduleSubjectLineRegex = regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)$`)
	})
	return moduleSubjectLineRegex
}

// ValidationError is an alias to the core contract ValidationError
type ValidationError = contracts.ValidationError

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
}

// LoadContractFromConfig loads commit-message contract from ai-config.yml
// Note: Validation is now done via JSON schema. This function returns minimal contract info.
func LoadContractFromConfig(workspaceRoot string) (*CommitMessageContract, error) {
	loader := ai.NewAIConfigLoader(workspaceRoot)
	typeConfig, err := loader.GetType("commit-message")
	if err != nil {
		return nil, fmt.Errorf("failed to load commit-message config: %w", err)
	}

	// Build minimal CommitMessageContract with standard defaults
	// Actual validation is done via JSON schema (commit-message.schema.json)
	contract := &CommitMessageContract{
		Version:           "0.1.0",
		Name:              typeConfig.Name,
		Description:       typeConfig.Description,
		SubjectLineFormat: "type(scope): subject",
		SemanticTypes:     []string{"feat", "fix", "refactor", "docs", "chore", "test", "perf", "style", "ci", "build"},
		Constraints: map[string]any{
			"max_line_length":     72,
			"scope_required":      true,
			"no_trailing_periods": true,
		},
		Structure: []struct {
			Section   string `yaml:"section"`
			Required  bool   `yaml:"required"`
			Format    string `yaml:"format"`
			MaxLength int    `yaml:"max_length,omitempty"`
		}{
			{Section: "header", Required: true, Format: "conventional"},
			{Section: "auditor-summary", Required: false, Format: "single-line"},
			{Section: "body", Required: false, Format: "markdown"},
			{Section: "modules", Required: false, Format: "module-sections"},
		},
	}

	return contract, nil
}

// VerifyCommitMessageContract validates a commit message against contracts/commit-message/0.1.0/structure.yml
// affectedModules is the list of modules that had staged changes
func VerifyCommitMessageContract(commitMessage string, affectedModules []string) []ValidationError {
	var errors []ValidationError

	lines := strings.Split(commitMessage, "\n")
	if len(lines) == 0 {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"EMPTY_MESSAGE",
			"Commit message is empty",
			0,
			"error",
		))
		return errors
	}

	// Validate different aspects
	errors = append(errors, validateHeader(lines)...)
	errors = append(errors, validateTopLevelBody(lines)...)
	errors = append(errors, validateModuleSections(lines, affectedModules)...)
	errors = append(errors, validateModuleSubjectLines(lines)...)
	errors = append(errors, validateModuleSectionStructure(lines)...)
	errors = append(errors, validateCodeBlocks(lines)...)
	errors = append(errors, validateLineLength(lines)...)

	return errors
}

// validateHeader validates the commit message header (first line)
func validateHeader(lines []string) []ValidationError {
	var errors []ValidationError

	// RULE 1: First line must be conventional commit header with scope
	// Format: <type>(<scope>): <summary>
	if !getConventionalCommitRegex().MatchString(lines[0]) {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"INVALID_HEADER_FORMAT",
			"Header must follow format: <type>(<scope>): <summary> (e.g., feat(cli): add new command)",
			1,
			"error",
		))
	}

	// RULE 2: Header max length
	if len(lines[0]) > MaxHeaderLength {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"HEADER_TOO_LONG",
			fmt.Sprintf("Header exceeds %d characters (%d chars)", MaxHeaderLength, len(lines[0])),
			1,
			"error",
		))
	}

	// RULE 3: No trailing period (except ellipsis "...")
	if strings.HasSuffix(lines[0], ".") && !strings.HasSuffix(lines[0], "...") {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"HEADER_TRAILING_PERIOD",
			"Header must not end with period",
			1,
			"error",
		))
	}

	return errors
}

// validateTopLevelBody validates the presence of Auditor-Summary and top-level body text
func validateTopLevelBody(lines []string) []ValidationError {
	var errors []ValidationError

	// RULE 4: Check for Auditor-Summary field (should appear early, after blank line)
	hasAuditorSummary := false
	for i := 1; i < len(lines) && i < 10; i++ { // Check first 10 lines
		if strings.HasPrefix(lines[i], "Auditor-Summary:") {
			hasAuditorSummary = true
			break
		}
	}
	if !hasAuditorSummary {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"MISSING_AUDITOR_SUMMARY",
			"Missing Auditor-Summary field after header",
			0,
			"error",
		))
	}

	// RULE 5: Check for top-level body text
	hasTopLevelBody := false
	afterAuditorSummary := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Auditor-Summary:") {
			afterAuditorSummary = true
			continue
		}

		// Check for body text after Auditor-Summary and before module sections
		if afterAuditorSummary && trimmed != "" &&
			!strings.HasPrefix(trimmed, "Auditor-Summary:") &&
			!strings.HasPrefix(trimmed, "Changes:") &&
			trimmed != "---" &&
			!isDashesLine(trimmed) &&
			!isModuleName(trimmed) {
			hasTopLevelBody = true
			break
		}
	}

	if !hasTopLevelBody {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"MISSING_TOP_LEVEL_BODY",
			"Missing top-level body text after title (before module sections)",
			0,
			"error",
		))
	}

	return errors
}

// validateModuleSections validates that multi-module commits have module sections
func validateModuleSections(lines []string, affectedModules []string) []ValidationError {
	var errors []ValidationError

	// Single-module commits don't require module sections
	if len(affectedModules) <= 1 {
		return errors
	}

	// Find which modules have sections
	foundModules := make(map[string]bool)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for module name followed by dashes
		if i < len(lines)-1 {
			nextLine := strings.TrimSpace(lines[i+1])
			if isModuleName(trimmed) && isDashesLine(nextLine) {
				foundModules[trimmed] = true
			}
		}
	}

	// Multi-module commits MUST have module sections
	if len(foundModules) == 0 {
		moduleList := strings.Join(affectedModules, ", ")
		errors = append(errors, *contracts.NewLegacyValidationError(
			"MISSING_MODULE_SECTION",
			fmt.Sprintf("Multi-module commit missing module sections. Expected: %s", moduleList),
			0,
			"error",
		))
		return errors
	}

	// Check which specific modules are missing
	var missingModules []string
	for _, expectedModule := range affectedModules {
		if !foundModules[expectedModule] {
			missingModules = append(missingModules, expectedModule)
		}
	}

	if len(missingModules) > 0 {
		moduleList := strings.Join(missingModules, ", ")
		errors = append(errors, *contracts.NewLegacyValidationError(
			"MISSING_MODULE_SECTION",
			fmt.Sprintf("Missing module sections for: %s", moduleList),
			0,
			"error",
		))
	}

	return errors
}

// validateLineLength validates that body text lines don't exceed max length
func validateLineLength(lines []string) []ValidationError {
	var errors []ValidationError

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip special lines that don't count toward line length
		if trimmed == "" ||
			strings.HasPrefix(trimmed, "Auditor-Summary:") ||
			strings.HasPrefix(trimmed, "Changes:") ||
			strings.HasPrefix(trimmed, "|") ||
			strings.HasPrefix(trimmed, "```") ||
			isDashesLine(trimmed) ||
			trimmed == "---" ||
			strings.HasPrefix(trimmed, "Agent:") {
			continue
		}

		if len(trimmed) > MaxLineLength {
			preview := trimmed
			if len(preview) > 60 {
				preview = preview[:57] + "..."
			}
			errors = append(errors, *contracts.NewLegacyValidationError(
				"LINE_TOO_LONG",
				fmt.Sprintf("Line exceeds %d characters (%d chars): %s", MaxLineLength, len(trimmed), preview),
				lineNum,
				"warning",
			))
		}
	}

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
					errors = append(errors, *contracts.NewLegacyValidationError(
						"MISSING_SUBJECT_LINE",
						fmt.Sprintf("Module '%s' missing subject line", currentModule),
						lineNum,
						"error",
					))
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
			if !getModuleSubjectLineRegex().MatchString(trimmed) {
				errors = append(errors, *contracts.NewLegacyValidationError(
					"INVALID_SUBJECT_FORMAT",
					fmt.Sprintf("Subject line does not follow '<module>: <type>: <description>' format: %s", trimmed),
					0,
					"error",
				))
			} else {
				// Validate subject line length
				if len(trimmed) > MaxSubjectLength {
					errors = append(errors, *contracts.NewLegacyValidationError(
						"SUBJECT_TOO_LONG",
						fmt.Sprintf("Subject line exceeds %d characters (%d chars)", MaxSubjectLength, len(trimmed)),
						0,
						"error",
					))
				}

				// Check no trailing period (except ellipsis "...")
				if strings.HasSuffix(trimmed, ".") && !strings.HasSuffix(trimmed, "...") {
					errors = append(errors, *contracts.NewLegacyValidationError(
						"SUBJECT_TRAILING_PERIOD",
						"Subject line must not end with period",
						0,
						"error",
					))
				}
			}

			foundSubjectLine = true
		}

		// Exit module section when we hit horizontal rule
		if trimmed == "---" {
			if inModuleSection && !foundSubjectLine {
				errors = append(errors, *contracts.NewLegacyValidationError(
					"MISSING_SUBJECT_LINE",
					fmt.Sprintf("Module '%s' missing subject line", currentModule),
					0,
					"error",
				))
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
		errors = append(errors, *contracts.NewLegacyValidationError(
			"UNCLOSED_CODE_BLOCK",
			fmt.Sprintf("Code block opened at line %d is not closed", codeBlockLine),
			0,
			"error",
		))
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
				errors = append(errors, *contracts.NewLegacyValidationError(
					"ORPHANED_DASHES_LINE",
					fmt.Sprintf("Orphaned dashes line at line %d - must be preceded by module name", i+1),
					0,
					"error",
				))
			}
		}

		// Check for module subject line without proper header structure
		if getModuleSubjectLineRegex().MatchString(trimmed) {
			// This is a subject line - check if it has proper module structure before it
			moduleName := getModuleSubjectLineRegex().FindStringSubmatch(trimmed)[1]

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
					errors = append(errors, *contracts.NewLegacyValidationError(
						"MALFORMED_MODULE_SECTION",
						fmt.Sprintf("Module section at line %d missing module name and dashes header", i+1),
						0,
						"error",
					))
				} else if !foundModuleName {
					errors = append(errors, *contracts.NewLegacyValidationError(
						"MISSING_MODULE_NAME",
						fmt.Sprintf("Module section at line %d missing module name (has dashes but no name)", i+1),
						0,
						"error",
					))
				} else if !foundDashes {
					errors = append(errors, *contracts.NewLegacyValidationError(
						"MISSING_MODULE_DASHES",
						fmt.Sprintf("Module section at line %d missing dashes separator (has name but no dashes)", i+1),
						0,
						"error",
					))
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
