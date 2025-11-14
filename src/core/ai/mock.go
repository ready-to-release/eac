// File: src/core/ai/mock.go
package ai

import "context"

// MockProvider is a test provider that returns a configured response
// Placed in ai package for easy access in tests
type MockProvider struct {
	response string
}

// NewMockProvider creates a mock provider that returns the configured response
func NewMockProvider(response string) *MockProvider {
	return &MockProvider{
		response: response,
	}
}

// Name returns "mock" for provider identification
func (p *MockProvider) Name() string {
	return "mock"
}

// Execute returns the configured mock response
func (p *MockProvider) Execute(ctx context.Context, input string, opts ...Option) (string, error) {
	return p.response, nil
}
