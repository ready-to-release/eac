package validator

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field    string      // Field path (e.g., "extensions[0].name")
	Rule     string      // Rule violated (e.g., "required", "pattern", "enum")
	Message  string      // Human-readable error message
	Value    interface{} // Actual value that failed validation
	Expected string      // Expected format/value description
}

// Error implements the error interface
func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationResult contains all validation errors
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError // For lenient mode
}

// IsValid returns true if there are no errors
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(field, rule, message string, value interface{}, expected string) {
	r.Errors = append(r.Errors, ValidationError{
		Field:    field,
		Rule:     rule,
		Message:  message,
		Value:    value,
		Expected: expected,
	})
}

// AddWarning adds a validation warning (for lenient mode)
func (r *ValidationResult) AddWarning(field, rule, message string, value interface{}, expected string) {
	r.Warnings = append(r.Warnings, ValidationError{
		Field:    field,
		Rule:     rule,
		Message:  message,
		Value:    value,
		Expected: expected,
	})
}

// Error returns a combined error message
func (r *ValidationResult) Error() string {
	if r.IsValid() {
		return ""
	}

	var messages []string
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("validation failed with %d error(s):\n%s",
		len(r.Errors), strings.Join(messages, "\n"))
}

// Summary returns a summary of validation results
func (r *ValidationResult) Summary() string {
	if r.IsValid() {
		if len(r.Warnings) > 0 {
			return fmt.Sprintf("validation passed with %d warning(s)", len(r.Warnings))
		}
		return "validation passed"
	}
	return fmt.Sprintf("validation failed: %d error(s), %d warning(s)",
		len(r.Errors), len(r.Warnings))
}

// Validation rule constants
const (
	RuleRequired       = "required"
	RulePattern        = "pattern"
	RuleEnum           = "enum"
	RuleMinimum        = "minimum"
	RuleMaximum        = "maximum"
	RuleMinItems       = "minItems"
	RuleMaxItems       = "maxItems"
	RuleFormat         = "format"
	RuleType           = "type"
	RuleUnique         = "unique"
	RuleDeprecated     = "deprecated"
	RuleNotImplemented = "not_implemented"
)

// Severity levels for validation
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)
