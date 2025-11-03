package cmd

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/spf13/cobra"
)

// CreateExtensionAliases creates direct command aliases for configured extensions
// This allows users to run "r2r pwsh" instead of "r2r run pwsh"
func CreateExtensionAliases() {
	// Only create aliases if config is loaded successfully
	if len(conf.Global.Extensions) == 0 {
		return
	}

	// Create an alias command for each configured extension
	for _, ext := range conf.Global.Extensions {
		// Create a local copy to avoid closure issues
		extension := ext

		// Check if a command with this name already exists
		existingCmd, _, _ := RootCmd.Find([]string{extension.Name})
		if existingCmd != nil && existingCmd != RootCmd {
			// Command already exists, skip creating alias
			log.Debug().Str("extension", extension.Name).Msg("Skipping alias - command already exists")
			continue
		}

		// Create the alias command
		aliasCmd := &cobra.Command{
			Use:   extension.Name,
			Short: fmt.Sprintf("Run %s extension (alias for 'r2r run %s')", extension.Name, extension.Name),
			Long: fmt.Sprintf(`This is a convenience alias for 'r2r run %s'.

All arguments after the extension name are passed directly to the container.

Examples:
  r2r %s                    # Interactive mode
  r2r %s --help             # Show extension help
  r2r %s echo "Hello"       # Run a command`,
				extension.Name, extension.Name, extension.Name, extension.Name),
			DisableFlagParsing: true, // Pass all flags to the extension
			Run: func(cmd *cobra.Command, args []string) {
				// Build the full run command arguments
				runArgs := append([]string{extension.Name}, args...)

				// Call the Run command's function directly
				RunCmd.Run(cmd, runArgs)
			},
		}

		// Add the alias command to root
		RootCmd.AddCommand(aliasCmd)

		log.Debug().Str("extension", extension.Name).Msg("Created extension alias command")
	}
}

// InitializeExtensionAliases should be called after config is loaded but before command execution
func InitializeExtensionAliases() {
	// Try to load config early for alias creation
	// This is best-effort - if it fails, we just won't have aliases
	configFile := os.Getenv("R2R_CONFIG")
	if configFile == "" {
		// Try to find config in default locations
		possibleConfigs := []string{
			"r2r-cli.yml",
			"r2r-cli.yaml",
			".r2r-cli.yml",
			".r2r-cli.yaml",
		}

		for _, cfg := range possibleConfigs {
			if _, err := os.Stat(cfg); err == nil {
				configFile = cfg
				break
			}
		}
	}

	if configFile != "" {
		// Try to load the config
		if err := conf.LoadConfig(configFile); err == nil {
			// Config loaded successfully, create aliases
			CreateExtensionAliases()
		} else {
			log.Debug().Err(err).Str("config", configFile).Msg("Failed to load config for aliases")
		}
	}
}
