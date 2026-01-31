// Package ai provides AI provider implementations for LLM integration.
//
// This adapter module implements the ai.Provider interface for multiple
// AI backends:
//   - Anthropic Claude (API and CLI)
//   - OpenAI GPT models
//   - Google Gemini
//
// Usage:
//
//	import "github.com/ready-to-release/eac/go/eac/adapters/ai"
//	import "github.com/ready-to-release/eac/go/eac/adapters/ai/providers"
//
//	executor := ai.NewExecutor(workspaceRoot)
//	providers.RegisterBuiltIn(executor)
//
//	response, err := executor.Execute(ctx, "Your prompt here")
package ai
