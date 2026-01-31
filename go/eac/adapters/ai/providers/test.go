// File: go/eac/adapters/ai/providers/test.go
package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/adapters/ai"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// TestProvider is a provider for acceptance tests that reads mock responses from files.
//
// Mock Response Resolution Order:
//  1. File: .r2r/test/ai-mock.txt (relative to repo root)
//  2. Environment variable: R2R_TEST_AI_RESPONSE
//  3. Error if neither is available
type TestProvider struct{}

// NewTestProvider creates a test provider for acceptance testing.
func NewTestProvider() *TestProvider {
	return &TestProvider{}
}

// Name returns "test" for provider identification.
func (p *TestProvider) Name() string {
	return "test"
}

// Execute returns the mock response from file or environment variable.
func (p *TestProvider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// 1. Try file-based mock response
	// Get repo root dynamically to support isolated test environments
	if repoRoot, repoErr := repository.GetRepositoryRoot(""); repoErr == nil && repoRoot != "" {
		mockFilePath := paths.AITestMockPath(repoRoot)
		if content, err := os.ReadFile(mockFilePath); err == nil {
			return string(content), nil
		}
	}

	// 2. Fall back to environment variable
	if response := os.Getenv("R2R_TEST_AI_RESPONSE"); response != "" {
		return response, nil
	}

	// 3. Error if no mock response configured
	return "", fmt.Errorf("test provider: no mock response configured. " +
		"Set R2R_TEST_AI_RESPONSE env var or create .r2r/test/ai-mock.txt file")
}
