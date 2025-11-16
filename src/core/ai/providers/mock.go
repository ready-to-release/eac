// File: src/core/ai/providers/mock.go
package providers

import (
	"context"

	"github.com/ready-to-release/eac/src/core/ai"
)

// MockProvider is a test provider that returns a configured response
//
// Intent: Provide a predictable AI provider for testing without API calls.
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Simple struct with single field
//   - Returns configured response immediately
//   - No complex logic or state management
//
// Easy to change:
//   - Can extend with error simulation
//   - Can add call counting for verification
//   - Doesn't affect real providers
//
// Hard to break:
//   - No external dependencies (no API calls)
//   - Deterministic output
//   - Fast execution (tests run quickly)
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
// Options are accepted but ignored (for interface compatibility)
func (p *MockProvider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	return p.response, nil
}
