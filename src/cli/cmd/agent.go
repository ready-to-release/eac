package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(AgentCmd)
}

// AgentCmd is the parent command for AI agent configuration
var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Configure AI agents for the project",
	Long:  `Manage AI provider configuration and agent settings for AI-powered commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			log.WithField("error", err).Error().Msg("Failed to show help")
		}
	},
}
