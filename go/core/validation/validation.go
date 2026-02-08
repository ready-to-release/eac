package validation

import "fmt"

// ValidationError represents a validation error with structured metadata.
type ValidationError struct {
	Code    *ErrorCode // Structured error code
	Message string
	Line    int
	Context string // Additional context for debugging
}

// NewValidationError creates a validation error with structured error code.
func NewValidationError(code ErrorCode, message string, line int) *ValidationError {
	return &ValidationError{
		Code:    &code,
		Message: message,
		Line:    line,
	}
}

// NewValidationErrorWithContext creates a validation error with additional context.
func NewValidationErrorWithContext(code ErrorCode, message string, line int, context string) *ValidationError {
	return &ValidationError{
		Code:    &code,
		Message: message,
		Line:    line,
		Context: context,
	}
}

// GetCode returns the error code string.
func (e *ValidationError) GetCode() string {
	if e.Code != nil {
		return e.Code.Code
	}
	return ""
}

// IsCritical returns true if error is critical (non-retriable)
// Critical errors should stop retry attempts immediately.
func (e *ValidationError) IsCritical() bool {
	if e.Code != nil {
		return !e.Code.Retriable
	}
	return false
}

// IsWarning returns true if error is a warning (not a critical error).
func (e *ValidationError) IsWarning() bool {
	if e.Code != nil {
		return e.Code.Severity == SeverityWarning
	}
	return false
}

// GetCategory returns the error category.
func (e *ValidationError) GetCategory() ErrorCategory {
	if e.Code != nil {
		return e.Code.Category
	}
	return ""
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	code := e.GetCode()
	if e.Line > 0 {
		return fmt.Sprintf("[%s] Line %d: %s", code, e.Line, e.Message)
	}
	return fmt.Sprintf("[%s] %s", code, e.Message)
}

// Validator interface must be implemented by specific contract validators.
type Validator interface {
	// Validate validates output against the contract
	Validate(output string, context map[string]interface{}) []ValidationError

	// VerifyImplementation verifies that the validator implements all contract rules
	VerifyImplementation() []ValidationError
}

// NoOpValidator is a validator that always returns valid (no validation)
// Used for formats that don't require validation or when validation is handled externally.
type NoOpValidator struct{}

// Validate always returns no errors (everything is valid).
func (v *NoOpValidator) Validate(_ string, _ map[string]interface{}) []ValidationError {
	return nil
}

// VerifyImplementation always returns no errors (no rules to verify).
func (v *NoOpValidator) VerifyImplementation() []ValidationError {
	return nil
}

// AIExecutor interface abstracts AI execution for contract-based generation.
type AIExecutor interface {
	Execute(ctx interface{}, prompt string, opts ...interface{}) (string, error)
}

// AIExecutorWithProviderInfo extends AIExecutor with provider information retrieval.
// Implementations can optionally implement this interface to provide metadata about
// the AI provider used for generation.
type AIExecutorWithProviderInfo interface {
	AIExecutor
	// GetProviderName returns the name of the AI provider used for the last execution.
	// Returns empty string if no execution has occurred or provider info is unavailable.
	GetProviderName() string
}
