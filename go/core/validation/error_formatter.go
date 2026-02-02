package validation

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ErrorFormatter provides utilities for creating AI-friendly error messages.
type ErrorFormatter struct{}

// NewErrorFormatter creates a new error formatter.
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{}
}

// FormatEnhancedError creates a comprehensive error message for AI generation
// following the template: What's Wrong | What AI Generated | Expected | Example | How to Fix.
func (f *ErrorFormatter) FormatEnhancedError(
	code ErrorCode,
	summary string,
	aiOutput string,
	expectedFormat string,
	correctExample string,
	fixGuidance string,
	lineNum int,
) *ValidationError {
	var sb strings.Builder

	// Section 1: What's wrong
	sb.WriteString(summary)
	sb.WriteString("\n")

	// Section 2: What AI generated (if provided)
	if aiOutput != "" {
		sb.WriteString("\nWhat AI generated:\n")
		sb.WriteString(f.indent(f.truncate(aiOutput, 10), "  "))
		sb.WriteString("\n")
	}

	// Section 3: Expected format (if provided)
	if expectedFormat != "" {
		sb.WriteString("\nExpected format:\n")
		sb.WriteString(f.indent(expectedFormat, "  "))
		sb.WriteString("\n")
	}

	// Section 4: Correct example (if provided)
	if correctExample != "" {
		sb.WriteString("\nCorrect example:\n")
		sb.WriteString(f.indent(correctExample, "  "))
		sb.WriteString("\n")
	}

	// Section 5: How to fix (if provided)
	if fixGuidance != "" {
		sb.WriteString("\nHow to fix:\n")
		sb.WriteString(f.indent(fixGuidance, "  "))
	}

	return NewValidationError(code, strings.TrimSpace(sb.String()), lineNum)
}

// FormatWithExample creates an error showing correct vs actual output.
func (f *ErrorFormatter) FormatWithExample(
	code ErrorCode,
	summary string,
	correctExample string,
	actualSnippet string,
	lineNum int,
) *ValidationError {
	var sb strings.Builder

	sb.WriteString(summary)
	sb.WriteString("\n")

	if actualSnippet != "" {
		sb.WriteString("\nAI generated:\n")
		sb.WriteString(f.indent(actualSnippet, "  "))
		sb.WriteString("\n")
	}

	if correctExample != "" {
		sb.WriteString("\nExpected:\n")
		sb.WriteString(f.indent(correctExample, "  "))
	}

	return NewValidationError(code, strings.TrimSpace(sb.String()), lineNum)
}

// FormatStructuredError creates an error showing expected structure.
func (f *ErrorFormatter) FormatStructuredError(
	code ErrorCode,
	summary string,
	structure string,
	actualFields []string,
	lineNum int,
) *ValidationError {
	var sb strings.Builder

	sb.WriteString(summary)
	sb.WriteString("\n")

	sb.WriteString("\nExpected structure:\n")
	sb.WriteString(f.indent(structure, "  "))
	sb.WriteString("\n")

	if len(actualFields) > 0 {
		sb.WriteString("\nFields AI provided:\n")
		for _, field := range actualFields {
			sb.WriteString(fmt.Sprintf("  • %s\n", field))
		}
	}

	return NewValidationError(code, strings.TrimSpace(sb.String()), lineNum)
}

// FormatMissingFieldError creates an error for missing required fields.
func (f *ErrorFormatter) FormatMissingFieldError(
	code ErrorCode,
	fieldName string,
	parentStructure string,
	actualFields []string,
	exampleValue string,
	lineNum int,
) *ValidationError {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Missing required field: %s", fieldName))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("\nThe '%s' must have a '%s' field.", parentStructure, fieldName))
	sb.WriteString("\n")

	if exampleValue != "" {
		sb.WriteString("\nExample:\n")
		sb.WriteString(f.indent(exampleValue, "  "))
		sb.WriteString("\n")
	}

	if len(actualFields) > 0 {
		sb.WriteString("\nFields currently present:\n")
		for _, field := range actualFields {
			sb.WriteString(fmt.Sprintf("  • %s\n", field))
		}
	}

	sb.WriteString(fmt.Sprintf("\nFix: Add the '%s' field with an appropriate value.", fieldName))

	return NewValidationError(code, strings.TrimSpace(sb.String()), lineNum)
}

// indent adds indentation to each line.
func (f *ErrorFormatter) indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	for i, line := range lines {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(prefix)
		sb.WriteString(line)
	}
	return sb.String()
}

// truncate limits text to maxLines.
func (f *ErrorFormatter) truncate(text string, maxLines int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}

	var sb strings.Builder
	for i := 0; i < maxLines; i++ {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(lines[i])
	}
	sb.WriteString("\n... (truncated)")
	return sb.String()
}

// TruncateJSON truncates JSON to show relevant parts.
func (f *ErrorFormatter) TruncateJSON(jsonStr string, maxLines int) string {
	// Try to pretty-print first
	var prettyJSON interface{}
	if err := json.Unmarshal([]byte(jsonStr), &prettyJSON); err == nil {
		if formatted, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
			jsonStr = string(formatted)
		}
	}

	return f.truncate(jsonStr, maxLines)
}

// ExtractJSONFields extracts top-level field names from JSON.
func (f *ErrorFormatter) ExtractJSONFields(jsonStr string) []string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return []string{}
	}

	fields := make([]string, 0, len(data))
	for key := range data {
		fields = append(fields, key)
	}
	return fields
}

// HighlightDifference shows what's different between expected and actual.
func (f *ErrorFormatter) HighlightDifference(expected, actual string) string {
	var sb strings.Builder

	sb.WriteString("Expected:\n")
	sb.WriteString(f.indent(expected, "  ✓ "))
	sb.WriteString("\n\n")

	sb.WriteString("Actual:\n")
	sb.WriteString(f.indent(actual, "  ✗ "))

	return sb.String()
}
