// File: go/eac/ai/providers/test.go
package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/ai"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// TestProvider is a provider for acceptance tests that reads mock responses from files.
//
// Intent: Enable AI mocking in subprocess-based acceptance tests where package-level
// variables cannot be shared across process boundaries.
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Reads mock response from well-known file location
//   - Falls back to environment variable for simple cases
//   - Clear error messages when mock response not found
//
// Easy to change:
//   - File location can be customized via config
//   - Environment variable fallback is optional
//   - Can extend to support multiple mock responses
//
// Hard to break:
//   - No external API calls
//   - Deterministic output based on file/env content
//   - Fast execution for quick test feedback
//
// Mock Response Resolution Order:
//  1. File: .r2r/test/ai-mock.txt (relative to repo root)
//  2. Environment variable: R2R_TEST_AI_RESPONSE
//  3. Error if neither is available
type TestProvider struct {
	repoRoot string
}

// NewTestProvider creates a test provider for acceptance testing
func NewTestProvider() *TestProvider {
	repoRoot, _ := repository.GetRepositoryRoot("")
	return &TestProvider{
		repoRoot: repoRoot,
	}
}

// Name returns "test" for provider identification
func (p *TestProvider) Name() string {
	return "test"
}

// Execute returns the mock response from file or environment variable
func (p *TestProvider) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// 1. Try file-based mock response
	if p.repoRoot != "" {
		mockFilePath := paths.AITestMockPath(p.repoRoot)
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
