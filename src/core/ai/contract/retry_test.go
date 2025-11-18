package contract

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Intent: Test generic retry framework with various scenarios
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with clear test case names
//   - Mock implementations are simple and focused
//   - Each test focuses on a single behavior
//
// Easy to change:
//   - Helper functions isolate test infrastructure
//   - Test data is clearly separated from test logic
//   - Mock implementations can be reused
//
// Hard to break:
//   - Tests cover happy path and error cases
//   - Edge cases are explicitly tested
//   - Mock executors and validators are deterministic

// MockAIExecutor implements AIExecutor for testing
type MockAIExecutor struct {
	responses  []string // Responses to return in sequence
	callCount  int      // Track how many times Execute was called
	lastPrompt string   // Last prompt received
}

func (m *MockAIExecutor) Execute(ctx context.Context, prompt string, opts ...interface{}) (string, error) {
	m.lastPrompt = prompt

	if m.callCount >= len(m.responses) {
		return "", nil
	}

	response := m.responses[m.callCount]
	m.callCount++
	return response, nil
}

// MockValidator implements SpecValidator for testing
type MockValidator struct {
	validationResults [][]ValidationError // Validation results to return in sequence
	callCount         int                 // Track how many times Validate was called
}

func (m *MockValidator) Validate(output string, context map[string]interface{}) []ValidationError {
	if m.callCount >= len(m.validationResults) {
		return []ValidationError{}
	}

	result := m.validationResults[m.callCount]
	m.callCount++
	return result
}

func (m *MockValidator) VerifyImplementation() []ValidationError {
	return []ValidationError{}
}

// MockRetryPromptBuilder implements RetryPromptBuilder for testing
type MockRetryPromptBuilder struct {
	buildCalled  bool
	lastOriginal string
	lastInvalid  string
	lastErrors   []ValidationError
}

func (m *MockRetryPromptBuilder) BuildRetryPrompt(original string, invalid string, errors []ValidationError) string {
	m.buildCalled = true
	m.lastOriginal = original
	m.lastInvalid = invalid
	m.lastErrors = errors
	return original + "\n\nRETRY WITH ERRORS"
}

func TestGenerateWithRetry_SuccessFirstAttempt(t *testing.T) {
	// Setup: Mock that returns valid output on first attempt
	executor := &MockAIExecutor{
		responses: []string{"valid output"},
	}

	validator := &MockValidator{
		validationResults: [][]ValidationError{
			{}, // No errors on first attempt
		},
	}

	cfg := &RetryConfig{
		Executor:    executor,
		Validator:   validator,
		MaxAttempts: 2,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "test prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "valid output" {
		t.Errorf("output = %q, want %q", result.Output, "valid output")
	}

	if result.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", result.Attempts)
	}

	if result.RetriedWithErrors {
		t.Error("RetriedWithErrors = true, want false (no retry should happen)")
	}

	if len(result.ValidationErrors) != 0 {
		t.Errorf("validation errors = %d, want 0", len(result.ValidationErrors))
	}

	if executor.callCount != 1 {
		t.Errorf("executor called %d times, want 1", executor.callCount)
	}

	if validator.callCount != 1 {
		t.Errorf("validator called %d times, want 1", validator.callCount)
	}
}

func TestGenerateWithRetry_FailFirstAttempt_SuccessSecondAttempt(t *testing.T) {
	// Setup: Mock that fails first, succeeds second
	executor := &MockAIExecutor{
		responses: []string{
			"invalid output",
			"valid output",
		},
	}

	validator := &MockValidator{
		validationResults: [][]ValidationError{
			// First attempt: has errors
			{
				{Code: "TEST_ERROR", Message: "test error", Severity: "error"},
			},
			// Second attempt: no errors
			{},
		},
	}

	cfg := &RetryConfig{
		Executor:    executor,
		Validator:   validator,
		MaxAttempts: 2,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "test prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "valid output" {
		t.Errorf("output = %q, want %q", result.Output, "valid output")
	}

	if result.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", result.Attempts)
	}

	if !result.RetriedWithErrors {
		t.Error("RetriedWithErrors = false, want true")
	}

	if len(result.ValidationErrors) != 0 {
		t.Errorf("validation errors = %d, want 0", len(result.ValidationErrors))
	}

	if executor.callCount != 2 {
		t.Errorf("executor called %d times, want 2", executor.callCount)
	}

	if validator.callCount != 2 {
		t.Errorf("validator called %d times, want 2", validator.callCount)
	}
}

func TestGenerateWithRetry_BothAttemptsFail(t *testing.T) {
	// Setup: Mock that fails both attempts
	executor := &MockAIExecutor{
		responses: []string{
			"invalid output 1",
			"invalid output 2",
		},
	}

	validator := &MockValidator{
		validationResults: [][]ValidationError{
			// First attempt: has errors
			{
				{Code: "ERROR_1", Message: "first error", Severity: "error"},
			},
			// Second attempt: different errors
			{
				{Code: "ERROR_2", Message: "second error", Severity: "error"},
			},
		},
	}

	cfg := &RetryConfig{
		Executor:    executor,
		Validator:   validator,
		MaxAttempts: 2,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "test prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", result.Attempts)
	}

	if !result.RetriedWithErrors {
		t.Error("RetriedWithErrors = false, want true")
	}

	// Should return errors from last attempt
	if len(result.ValidationErrors) != 1 {
		t.Errorf("validation errors = %d, want 1", len(result.ValidationErrors))
	}

	if result.ValidationErrors[0].Code != "ERROR_2" {
		t.Errorf("error code = %q, want ERROR_2", result.ValidationErrors[0].Code)
	}

	if executor.callCount != 2 {
		t.Errorf("executor called %d times, want 2", executor.callCount)
	}
}

func TestGenerateWithRetry_CustomPromptBuilder(t *testing.T) {
	// Setup
	executor := &MockAIExecutor{
		responses: []string{
			"invalid output",
			"valid output",
		},
	}

	validator := &MockValidator{
		validationResults: [][]ValidationError{
			{{Code: "TEST_ERROR", Message: "test error", Severity: "error"}},
			{},
		},
	}

	promptBuilder := &MockRetryPromptBuilder{}

	cfg := &RetryConfig{
		Executor:      executor,
		Validator:     validator,
		PromptBuilder: promptBuilder,
		MaxAttempts:   2,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "original prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !promptBuilder.buildCalled {
		t.Error("custom prompt builder was not called")
	}

	if promptBuilder.lastOriginal != "original prompt" {
		t.Errorf("prompt builder received original = %q, want %q", promptBuilder.lastOriginal, "original prompt")
	}

	if promptBuilder.lastInvalid != "invalid output" {
		t.Errorf("prompt builder received invalid = %q, want %q", promptBuilder.lastInvalid, "invalid output")
	}

	if len(promptBuilder.lastErrors) != 1 {
		t.Errorf("prompt builder received %d errors, want 1", len(promptBuilder.lastErrors))
	}

	// Verify retry prompt was used
	if !strings.Contains(executor.lastPrompt, "RETRY WITH ERRORS") {
		t.Errorf("executor did not receive retry prompt, got: %s", executor.lastPrompt)
	}

	if result.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", result.Attempts)
	}
}

func TestGenerateWithRetry_NoValidator(t *testing.T) {
	// Setup: No validator provided
	executor := &MockAIExecutor{
		responses: []string{"output"},
	}

	cfg := &RetryConfig{
		Executor:    executor,
		Validator:   nil, // No validator
		MaxAttempts: 2,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "test prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", result.Attempts)
	}

	if len(result.ValidationErrors) != 0 {
		t.Errorf("validation errors = %d, want 0 (no validator)", len(result.ValidationErrors))
	}
}

func TestGenerateWithRetry_AntiCorruptionApplied(t *testing.T) {
	// Setup
	executor := &MockAIExecutor{
		responses: []string{"```\nclean output\n```"},
	}

	validator := &MockValidator{
		validationResults: [][]ValidationError{
			{}, // Valid
		},
	}

	antiCorruption := &AntiCorruptionRules{
		// This will be used by ApplyWithFallback
	}

	cfg := &RetryConfig{
		Executor:       executor,
		Validator:      validator,
		AntiCorruption: antiCorruption,
		ContentMarker:  "",
		MaxAttempts:    2,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "test prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Anti-corruption should strip code fences
	if strings.Contains(result.Output, "```") {
		t.Errorf("output still contains code fences, anti-corruption not applied: %s", result.Output)
	}

	if !strings.Contains(result.Output, "clean output") {
		t.Errorf("output = %q, should contain 'clean output'", result.Output)
	}
}

func TestGenerateWithRetry_DebugMode(t *testing.T) {
	// Setup temp directory for debug files
	tmpDir := t.TempDir()

	executor := &MockAIExecutor{
		responses: []string{
			"invalid output",
			"valid output",
		},
	}

	validator := &MockValidator{
		validationResults: [][]ValidationError{
			{{Code: "TEST_ERROR", Message: "test error", Severity: "error"}},
			{},
		},
	}

	cfg := &RetryConfig{
		Executor:       executor,
		Validator:      validator,
		MaxAttempts:    2,
		Debug:          true,
		DebugOutputDir: tmpDir,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "test prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", result.Attempts)
	}

	// Check that debug files were created
	expectedFiles := []string{
		"attempt-1-raw.txt",
		"attempt-1-cleaned.txt",
		"attempt-1-errors.txt",
		"attempt-2-raw.txt",
		"attempt-2-cleaned.txt",
	}

	for _, fileName := range expectedFiles {
		filePath := filepath.Join(tmpDir, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("expected debug file not created: %s", fileName)
		}
	}
}

func TestGenerateWithRetry_InvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *RetryConfig
		wantError string
	}{
		{
			name:      "nil config",
			cfg:       nil,
			wantError: "config is nil",
		},
		{
			name: "missing executor",
			cfg: &RetryConfig{
				Validator: &MockValidator{},
			},
			wantError: "executor is required",
		},
		{
			name: "debug without output dir",
			cfg: &RetryConfig{
				Executor: &MockAIExecutor{responses: []string{"test"}},
				Debug:    true,
			},
			wantError: "debug mode requires DebugOutputDir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateWithRetry(context.Background(), tt.cfg, "test prompt")

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

func TestDefaultRetryPromptBuilder_BuildRetryPrompt(t *testing.T) {
	builder := &DefaultRetryPromptBuilder{}

	originalPrompt := "Generate something"
	invalidOutput := "bad output"
	errors := []ValidationError{
		{Code: "ERROR_1", Message: "first error", Line: 1, Severity: "error"},
		{Code: "ERROR_2", Message: "second error", Line: 2, Severity: "warning"},
	}

	result := builder.BuildRetryPrompt(originalPrompt, invalidOutput, errors)

	// Check that result contains key elements
	if !strings.Contains(result, originalPrompt) {
		t.Error("retry prompt should contain original prompt")
	}

	if !strings.Contains(result, "VALIDATION ERRORS") {
		t.Error("retry prompt should mention validation errors")
	}

	if !strings.Contains(result, "ERROR_1") {
		t.Error("retry prompt should include error codes")
	}

	if !strings.Contains(result, "first error") {
		t.Error("retry prompt should include error messages")
	}

	if !strings.Contains(result, "2 validation error(s)") {
		t.Error("retry prompt should show error count")
	}

	if !strings.Contains(result, "1 critical") {
		t.Error("retry prompt should show critical error count")
	}
}

func TestGenerateWithRetry_WarningsOnly(t *testing.T) {
	// Setup: Validator returns warnings (not errors)
	executor := &MockAIExecutor{
		responses: []string{"output with warnings"},
	}

	validator := &MockValidator{
		validationResults: [][]ValidationError{
			{
				{Code: "WARNING_1", Message: "just a warning", Severity: "warning"},
			},
		},
	}

	cfg := &RetryConfig{
		Executor:    executor,
		Validator:   validator,
		MaxAttempts: 2,
	}

	// Execute
	result, err := GenerateWithRetry(context.Background(), cfg, "test prompt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Warnings should not trigger retry
	if result.Attempts != 1 {
		t.Errorf("attempts = %d, want 1 (warnings should not trigger retry)", result.Attempts)
	}

	// Warnings should still be returned
	if len(result.ValidationErrors) != 1 {
		t.Errorf("validation errors = %d, want 1 (warning)", len(result.ValidationErrors))
	}
}
