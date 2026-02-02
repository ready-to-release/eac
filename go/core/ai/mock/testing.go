// Package mock provides mock implementations for testing AI and validation components.
package mock

import "github.com/ready-to-release/eac/go/core/validation"

// AIExecutor implements validation.AIExecutor for testing.
type AIExecutor struct {
	Responses      []string // Queue of responses to return
	ResponseIndex  int      // Current response index
	ExecutionCount int      // Number of times Execute was called
}

// Execute returns the next response from the queue.
func (m *AIExecutor) Execute(_ interface{}, _ string, _ ...interface{}) (string, error) {
	m.ExecutionCount++
	if m.ResponseIndex >= len(m.Responses) {
		return m.Responses[len(m.Responses)-1], nil
	}
	response := m.Responses[m.ResponseIndex]
	m.ResponseIndex++
	return response, nil
}

// Validator implements validation.Validator for testing.
type Validator struct {
	ValidationResults [][]validation.ValidationError // Queue of validation results
	ValidationIndex   int                            // Current validation index
	ValidationCount   int                            // Number of times Validate was called
}

// Validate returns the next validation result from the queue.
func (m *Validator) Validate(_ string, _ map[string]interface{}) []validation.ValidationError {
	m.ValidationCount++
	if m.ValidationIndex >= len(m.ValidationResults) {
		return m.ValidationResults[len(m.ValidationResults)-1]
	}
	result := m.ValidationResults[m.ValidationIndex]
	m.ValidationIndex++
	return result
}

// VerifyImplementation always returns nil for the mock.
func (m *Validator) VerifyImplementation() []validation.ValidationError {
	return nil
}
