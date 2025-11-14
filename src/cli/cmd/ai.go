package cmd

import (
	"context"
	"fmt"

	"github.com/ready-to-release/eac/src/core/ai"
	"github.com/ready-to-release/eac/src/core/ai/providers"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(AICmd)
	AICmd.AddCommand(AIAskCmd)

	// Add flags
	AIAskCmd.Flags().StringP("model", "m", "", "Override model (e.g., opus, sonnet, haiku)")
	AIAskCmd.Flags().Float64P("temperature", "t", 0.3, "Temperature (0.0-1.0)")
}

// AICmd is the parent command for AI operations
var AICmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered operations",
	Long:  `Execute AI-powered operations using the configured provider.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.WithField("error", err).Error().Msg("Failed to show help")
		}
	},
}

// AIAskCmd executes a simple AI prompt
var AIAskCmd = &cobra.Command{
	Use:   "ask <prompt>",
	Short: "Ask AI a question",
	Long: `Ask the configured AI provider a question.

This is a demonstration command showing how to use the AI executor.
The executor will use the provider configured in .r2r/agent-config.yml
or fall back to claude-cli if no configuration exists.

Examples:
  r2r ai ask "What is the capital of France?"
  r2r ai ask "Write a haiku about coding" --model opus
  r2r ai ask "Explain recursion" --temperature 0.7`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get workspace root
		workspaceRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to find repository root: %w", err)
		}

		// Create executor
		executor := ai.NewExecutor(workspaceRoot)
		providers.RegisterBuiltIn(executor)

		// Set up logging
		logger := ai.NewFileLogger(workspaceRoot)
		executor.SetLogger(logger)

		// Get prompt from args
		prompt := args[0]

		// Build options
		var opts []ai.Option
		if model, _ := cmd.Flags().GetString("model"); model != "" {
			opts = append(opts, ai.WithModel(model))
		}
		if temp, _ := cmd.Flags().GetFloat64("temperature"); cmd.Flags().Changed("temperature") {
			opts = append(opts, ai.WithTemperature(temp))
		}

		// Execute
		fmt.Printf("🤖 Asking AI: %s\n\n", prompt)

		ctx := context.Background()
		response, err := executor.Execute(ctx, prompt, opts...)
		if err != nil {
			return fmt.Errorf("AI execution failed: %w", err)
		}

		// Display response
		fmt.Printf("💬 Response:\n%s\n\n", response)

		// Show provider used
		if provider := executor.GetLastUsedProvider(); provider != nil {
			fmt.Printf("ℹ️  Provider: %s\n", provider.Name())
		}

		return nil
	},
}
