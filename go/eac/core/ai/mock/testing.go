package mock

import "github.com/ready-to-release/eac/go/eac/core/validation"

// MockAIExecutor implements validation.AIExecutor for testing.
type MockAIExecutor struct {
	Responses      []string // Queue of responses to return
	ResponseIndex  int      // Current response index
	ExecutionCount int      // Number of times Execute was called
}

func (m *MockAIExecutor) Execute(ctx interface{}, prompt string, opts ...interface{}) (string, error) {
	m.ExecutionCount++
	if m.ResponseIndex >= len(m.Responses) {
		return m.Responses[len(m.Responses)-1], nil
	}
	response := m.Responses[m.ResponseIndex]
	m.ResponseIndex++
	return response, nil
}

// MockValidator implements validation.Validator for testing.
type MockValidator struct {
	ValidationResults [][]validation.ValidationError // Queue of validation results
	ValidationIndex   int                            // Current validation index
	ValidationCount   int                            // Number of times Validate was called
}

func (m *MockValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	m.ValidationCount++
	if m.ValidationIndex >= len(m.ValidationResults) {
		return m.ValidationResults[len(m.ValidationResults)-1]
	}
	result := m.ValidationResults[m.ValidationIndex]
	m.ValidationIndex++
	return result
}

func (m *MockValidator) VerifyImplementation() []validation.ValidationError {
	return nil
}
