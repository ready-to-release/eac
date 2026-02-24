//go:build L0 && ov

package generation

import (
	"context"
	"testing"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/core/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockExecutor implements ai.GenerationExecutor for testing.
type mockExecutor struct {
	responses []string
	callCount int
}

func (m *mockExecutor) Execute(_ context.Context, _ string, _ ...ai.Option) (string, error) {
	idx := m.callCount
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	m.callCount++
	return m.responses[idx], nil
}

// mockValidator implements validation.Validator for testing.
type mockValidator struct {
	results [][]validation.ValidationError
	callIdx int
}

func (m *mockValidator) Validate(content string, _ map[string]interface{}) []validation.ValidationError {
	idx := m.callIdx
	if idx >= len(m.results) {
		idx = len(m.results) - 1
	}
	m.callIdx++
	return m.results[idx]
}

func (m *mockValidator) VerifyImplementation() []validation.ValidationError {
	return nil
}

func TestStructuredGenerator_Generate_Success(t *testing.T) {
	executor := &mockExecutor{responses: []string{`{"key": "value"}`}}
	validator := &mockValidator{results: [][]validation.ValidationError{nil}}

	gen := &StructuredGenerator{
		OutputFormat: FormatJSON,
		Executor:     executor,
		Validator:    validator,
		MaxAttempts:  3,
		Logger:       zap.NewNop(),
	}

	result, err := gen.Generate(context.Background(), "Generate JSON")
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, 1, result.Attempts)
	assert.Equal(t, `{"key": "value"}`, result.GeneratedContent)
}

func TestStructuredGenerator_Generate_RetryOnValidationFailure(t *testing.T) {
	executor := &mockExecutor{responses: []string{`bad`, `{"key": "value"}`}}
	validator := &mockValidator{
		results: [][]validation.ValidationError{
			{*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0)},
			nil,
		},
	}

	gen := &StructuredGenerator{
		OutputFormat: FormatJSON,
		Executor:     executor,
		Validator:    validator,
		MaxAttempts:  3,
		Logger:       zap.NewNop(),
	}

	result, err := gen.Generate(context.Background(), "Generate JSON")
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, 2, result.Attempts)
}

func TestStructuredGenerator_Generate_AllAttemptsExhausted(t *testing.T) {
	executor := &mockExecutor{responses: []string{`bad`}}
	validator := &mockValidator{
		results: [][]validation.ValidationError{
			{*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0)},
		},
	}

	gen := &StructuredGenerator{
		OutputFormat: FormatJSON,
		Executor:     executor,
		Validator:    validator,
		MaxAttempts:  2,
		Logger:       zap.NewNop(),
	}

	result, err := gen.Generate(context.Background(), "Generate JSON")
	require.Error(t, err)
	assert.False(t, result.Valid)
	assert.Equal(t, 2, result.Attempts)
}

func TestStructuredGenerator_CleanupOutput_StripsFences(t *testing.T) {
	gen := &StructuredGenerator{}

	tests := []struct {
		name     string
		input    string
		format   StructuredFormat
		expected string
	}{
		{
			name:     "JSON with fences",
			input:    "```json\n{\"key\": \"value\"}\n```",
			format:   FormatJSON,
			expected: `{"key": "value"}`,
		},
		{
			name:     "Gherkin passes through",
			input:    "Feature: Test\n  Scenario: S1",
			format:   FormatGherkin,
			expected: "Feature: Test\n  Scenario: S1",
		},
		{
			name:     "plain JSON passes through",
			input:    `{"key": "value"}`,
			format:   FormatJSON,
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.cleanupOutput(tt.input, tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStructuredGenerator_GetStrategy_Default(t *testing.T) {
	gen := &StructuredGenerator{}
	strategy := gen.getStrategy()
	assert.IsType(t, &StandardStrategy{}, strategy)
}

func TestStructuredGenerator_GetStrategy_Custom(t *testing.T) {
	custom := &FocusedStrategy{}
	gen := &StructuredGenerator{Strategy: custom}
	strategy := gen.getStrategy()
	assert.Equal(t, custom, strategy)
}
