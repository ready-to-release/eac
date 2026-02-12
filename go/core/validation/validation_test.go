package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name    string
		code    ErrorCode
		message string
		line    int
	}{
		{
			name:    "creates error with all fields",
			code:    ErrEmptyOutput,
			message: "output was empty",
			line:    5,
		},
		{
			name:    "creates error with zero line",
			code:    ErrMissingFeature,
			message: "no feature found",
			line:    0,
		},
		{
			name:    "creates error with warning severity",
			code:    ErrUnknownTag,
			message: "unknown tag @foo",
			line:    12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := NewValidationError(tt.code, tt.message, tt.line)

			require.NotNil(t, ve)
			require.NotNil(t, ve.Code)
			assert.Equal(t, tt.code.Code, ve.Code.Code)
			assert.Equal(t, tt.message, ve.Message)
			assert.Equal(t, tt.line, ve.Line)
			assert.Empty(t, ve.Context)
		})
	}
}

func TestNewValidationErrorWithContext(t *testing.T) {
	tests := []struct {
		name    string
		code    ErrorCode
		message string
		line    int
		context string
	}{
		{
			name:    "creates error with context",
			code:    ErrInvalidJSON,
			message: "invalid json",
			line:    1,
			context: "unexpected token at position 42",
		},
		{
			name:    "creates error with empty context",
			code:    ErrEmptyOutput,
			message: "empty",
			line:    0,
			context: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ve := NewValidationErrorWithContext(tt.code, tt.message, tt.line, tt.context)

			require.NotNil(t, ve)
			require.NotNil(t, ve.Code)
			assert.Equal(t, tt.code.Code, ve.Code.Code)
			assert.Equal(t, tt.message, ve.Message)
			assert.Equal(t, tt.line, ve.Line)
			assert.Equal(t, tt.context, ve.Context)
		})
	}
}

func TestValidationError_GetCode(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "returns code when Code is set",
			err:  NewValidationError(ErrEmptyOutput, "empty", 0),
			want: "EMPTY_OUTPUT",
		},
		{
			name: "returns code for semantic error",
			err:  NewValidationError(ErrMissingFeature, "no feature", 1),
			want: "MISSING_FEATURE",
		},
		{
			name: "returns empty string when Code is nil",
			err:  &ValidationError{Code: nil, Message: "orphan"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.GetCode())
		})
	}
}

func TestValidationError_IsCritical(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want bool
	}{
		{
			name: "non-retriable structure error is critical",
			err:  NewValidationError(ErrEmptyOutput, "empty", 0),
			want: true,
		},
		{
			name: "non-retriable setup error is critical",
			err:  NewValidationError(ErrSetupError, "setup failed", 0),
			want: true,
		},
		{
			name: "retriable error is not critical",
			err:  NewValidationError(ErrInvalidJSON, "bad json", 1),
			want: false,
		},
		{
			name: "retriable semantic error is not critical",
			err:  NewValidationError(ErrMissingFeature, "no feature", 0),
			want: false,
		},
		{
			name: "nil Code returns false",
			err:  &ValidationError{Code: nil, Message: "orphan"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.IsCritical())
		})
	}
}

func TestValidationError_IsWarning(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want bool
	}{
		{
			name: "warning severity returns true",
			err:  NewValidationError(ErrUnknownTag, "unknown @foo", 1),
			want: true,
		},
		{
			name: "warning quick validation failed",
			err:  NewValidationError(ErrQuickValidationFailed, "skipped", 0),
			want: true,
		},
		{
			name: "error severity returns false",
			err:  NewValidationError(ErrEmptyOutput, "empty", 0),
			want: false,
		},
		{
			name: "semantic error severity returns false",
			err:  NewValidationError(ErrMissingFeature, "missing", 0),
			want: false,
		},
		{
			name: "nil Code returns false",
			err:  &ValidationError{Code: nil, Message: "orphan"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.IsWarning())
		})
	}
}

func TestValidationError_GetCategory(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want ErrorCategory
	}{
		{
			name: "structure category",
			err:  NewValidationError(ErrEmptyOutput, "empty", 0),
			want: CategoryStructure,
		},
		{
			name: "semantic category",
			err:  NewValidationError(ErrMissingFeature, "missing", 0),
			want: CategorySemantic,
		},
		{
			name: "format category",
			err:  NewValidationError(ErrInvalidFeatureNaming, "bad name", 0),
			want: CategoryFormat,
		},
		{
			name: "constraint category",
			err:  NewValidationError(ErrTooManyRules, "too many", 0),
			want: CategoryConstraint,
		},
		{
			name: "corruption category",
			err:  NewValidationError(ErrForbiddenPrefix, "forbidden", 0),
			want: CategoryCorruption,
		},
		{
			name: "nil Code returns empty string",
			err:  &ValidationError{Code: nil, Message: "orphan"},
			want: ErrorCategory(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.GetCategory())
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  ValidationError
		want string
	}{
		{
			name: "formats with line number",
			err:  *NewValidationError(ErrEmptyOutput, "output was empty", 5),
			want: "[EMPTY_OUTPUT] Line 5: output was empty",
		},
		{
			name: "formats without line number when zero",
			err:  *NewValidationError(ErrMissingFeature, "no feature found", 0),
			want: "[MISSING_FEATURE] no feature found",
		},
		{
			name: "formats without line number when negative",
			err: ValidationError{
				Code:    &ErrorCode{Code: "TEST_CODE"},
				Message: "test message",
				Line:    -1,
			},
			want: "[TEST_CODE] test message",
		},
		{
			name: "formats with nil code",
			err:  ValidationError{Code: nil, Message: "orphan message", Line: 3},
			want: "[] Line 3: orphan message",
		},
		{
			name: "formats with nil code and no line",
			err:  ValidationError{Code: nil, Message: "orphan message", Line: 0},
			want: "[] orphan message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestValidationError_ImplementsErrorInterface(t *testing.T) {
	ve := NewValidationError(ErrEmptyOutput, "test", 1)

	// ValidationError (value receiver) should implement error interface
	var err error = *ve
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "EMPTY_OUTPUT")
}

func TestNoOpValidator_Validate(t *testing.T) {
	v := &NoOpValidator{}

	tests := []struct {
		name    string
		output  string
		context map[string]interface{}
	}{
		{
			name:    "empty output returns nil",
			output:  "",
			context: nil,
		},
		{
			name:    "non-empty output returns nil",
			output:  "some output",
			context: map[string]interface{}{"key": "value"},
		},
		{
			name:    "nil context returns nil",
			output:  "test",
			context: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(tt.output, tt.context)
			assert.Nil(t, errs)
		})
	}
}

func TestNoOpValidator_VerifyImplementation(t *testing.T) {
	v := &NoOpValidator{}
	errs := v.VerifyImplementation()
	assert.Nil(t, errs)
}

func TestNoOpValidator_ImplementsValidatorInterface(t *testing.T) {
	var v Validator = &NoOpValidator{}
	assert.NotNil(t, v)
}
