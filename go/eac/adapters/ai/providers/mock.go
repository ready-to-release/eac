// File: go/eac/adapters/ai/providers/mock.go
package providers

import (
	"context"

	"github.com/ready-to-release/eac/go/eac/adapters/ai"
)

// MockProvider is a test provider that returns a configured response.
type MockProvider struct {
	response string
}

// NewMockProvider creates a mock provider that returns the configured response.
func NewMockProvider(response string) *MockProvider {
	return &MockProvider{
		response: response,
	}
}

// Name returns "mock" for provider identification.
func (p *MockProvider) Name() string {
	return "mock"
}

// Execute returns the configured mock response
// Options are accepted but ignored (for interface compatibility).
func (p *MockProvider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	return p.response, nil
}
