// Package commitmessage provides commit message validation against contract
package commitmessage

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationError represents a contract violation
type ValidationError struct {
	Code     string
	Message  string
	Line     int
	Severity string // "error" or "warning"
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%s] Line %d: %s", e.Code, e.Line, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

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
	expectedVersion := "0.1.0"
	if contract.Version != expectedVersion {
		errors = append(errors, ValidationError{
			Code:     "CONTRACT_VERSION_MISMATCH",
			Message:  fmt.Sprintf("Expected version %s, got %s", expectedVersion, contract.Version),
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
			if section.MaxLength != 72 {
				errors = append(errors, ValidationError{
					Code:     "CONTRACT_CONSTRAINT_MISMATCH",
					Message:  fmt.Sprintf("top_level_heading max_length should be 72, got %d", section.MaxLength),
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
		"max_line_length":         72,
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

	// RULE 1: First line must be top-level heading with conventional commit format
	conventionalCommitRegex := regexp.MustCompile(`^# ([a-z0-9\-]+|multi-module):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)$`)

	if !strings.HasPrefix(lines[0], "# ") {
		errors = append(errors, ValidationError{
			Code:     "MISSING_TOP_HEADING",
			Message:  "First line must start with '# '",
			Line:     1,
			Severity: "error",
		})
	} else {
		// RULE 2: Title must follow conventional commit format
		if !conventionalCommitRegex.MatchString(lines[0]) {
			errors = append(errors, ValidationError{
				Code:     "INVALID_TITLE_FORMAT",
				Message:  "Title must follow format: # <module|multi-module>: <type>: <summary>",
				Line:     1,
				Severity: "error",
			})
		}

		title := strings.TrimPrefix(lines[0], "# ")

		// RULE 3: Title max 72 characters
		if len(lines[0]) > 72 {
			errors = append(errors, ValidationError{
				Code:     "TITLE_TOO_LONG",
				Message:  fmt.Sprintf("Title exceeds 72 characters (%d chars)", len(lines[0])),
				Line:     1,
				Severity: "error",
			})
		}

		// RULE 4: No trailing period (except ellipsis "...")
		if strings.HasSuffix(title, ".") && !strings.HasSuffix(title, "...") {
			errors = append(errors, ValidationError{
				Code:     "TITLE_TRAILING_PERIOD",
				Message:  "Title must not end with period",
				Line:     1,
				Severity: "error",
			})
		}
	}

	// RULE 5: Check for top-level body (should appear after title, before first ## section)
	hasTopLevelBody := false
	hasModuleSection := false
	foundModules := make(map[string]bool) // Track which modules we found in the commit message

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Check if we have body text before any ## sections
		if i > 0 && !hasModuleSection && !strings.HasPrefix(trimmed, "##") && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			hasTopLevelBody = true
		}

		// Module sections start with ##
		if strings.HasPrefix(trimmed, "## ") {
			hasModuleSection = true
			moduleHeader := strings.TrimPrefix(trimmed, "## ")
			foundModules[moduleHeader] = true

			// RULE 6: Module headers must be plain name (no colons)
			if strings.Contains(moduleHeader, ":") {
				errors = append(errors, ValidationError{
					Code:     "MODULE_HEADER_FORMAT",
					Message:  fmt.Sprintf("Module header must be plain name only, found: '%s'", moduleHeader),
					Line:     lineNum,
					Severity: "error",
				})
			}
		}

		// RULE 7: Line length in body text (skip headers, tables, code blocks, horizontal rules)
		if trimmed != "" &&
			!strings.HasPrefix(trimmed, "#") &&
			!strings.HasPrefix(trimmed, "|") &&
			!strings.HasPrefix(trimmed, "```") &&
			trimmed != "---" &&
			!strings.HasPrefix(trimmed, "Agent:") {
			if len(trimmed) > 72 {
				errors = append(errors, ValidationError{
					Code:     "LINE_TOO_LONG",
					Message:  fmt.Sprintf("Line exceeds 72 characters (%d chars)", len(trimmed)),
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

	// RULE 9: Check for unclosed code blocks
	errors = append(errors, validateCodeBlocks(lines)...)

	return errors
}

// validateModuleSubjectLines checks that module sections have proper subject lines
func validateModuleSubjectLines(lines []string) []ValidationError {
	var errors []ValidationError

	// Regex for semantic subject line: <module>: <type>: <description>
	subjectRegex := regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)$`)

	inModuleSection := false
	currentModule := ""
	foundSubjectLine := false

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Detect module section header
		if strings.HasPrefix(trimmed, "## ") {

			// If we were in a module section and didn't find subject line
			if inModuleSection && !foundSubjectLine {
				errors = append(errors, ValidationError{
					Code:     "MISSING_SUBJECT_LINE",
					Message:  fmt.Sprintf("Module '%s' missing subject line", currentModule),
					Severity: "error",
				})
			}

			inModuleSection = true
			currentModule = strings.TrimPrefix(trimmed, "## ")
			foundSubjectLine = false
			continue
		}

		// If in module section, look for subject line
		if inModuleSection && !foundSubjectLine && trimmed != "" {
			// Skip blank lines
			if trimmed == "" {
				continue
			}

			// This should be the subject line
			if !subjectRegex.MatchString(trimmed) {
				errors = append(errors, ValidationError{
					Code:     "INVALID_SUBJECT_FORMAT",
					Message:  fmt.Sprintf("Subject line does not follow '<module>: <type>: <description>' format: %s", trimmed),
					Line:     lineNum,
					Severity: "error",
				})
			} else {
				// Validate subject line length
				if len(trimmed) > 72 {
					errors = append(errors, ValidationError{
						Code:     "SUBJECT_TOO_LONG",
						Message:  fmt.Sprintf("Subject line exceeds 72 characters (%d chars)", len(trimmed)),
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
