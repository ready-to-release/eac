// Package aitest provides test and mock AI provider adapters for acceptance testing.
package aitest

import (
	"context"
	"fmt"
	"os"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	aiproviders.Register("test", func(config *ai.ProviderConfig) (ai.Provider, error) {
		return NewTestProvider(), nil
	})
}

// TestProvider is a provider for acceptance tests that reads mock responses from files.
//
// Mock Response Resolution Order:
//  1. File: .eac/test/ai-mock.txt (relative to repo root)
//  2. Environment variable: CLIE_TEST_AI_RESPONSE
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
	if repoRoot, repoErr := repository.GetRepositoryRoot(""); repoErr == nil && repoRoot != "" {
		mockFilePath := paths.AITestMockPath(repoRoot)
		if content, err := os.ReadFile(mockFilePath); err == nil {
			return string(content), nil
		}
	}

	// 2. Fall back to environment variable
	if response := os.Getenv(environments.EnvCLIETestAIResponse); response != "" {
		return response, nil
	}

	// 3. Error if no mock response configured
	return "", fmt.Errorf("test provider: no mock response configured. " +
		"Set CLIE_TEST_AI_RESPONSE env var or create .eac/test/ai-mock.txt file")
}

// Compile-time interface check.
var _ ai.Provider = (*TestProvider)(nil)
