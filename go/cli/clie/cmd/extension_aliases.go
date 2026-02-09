package cmd

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/cli/clie/internal/conf"
	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
	"github.com/ready-to-release/eac/go/cli/clie/internal/envconsts"
	"github.com/spf13/cobra"
)

// CreateExtensionAliases creates direct command aliases for configured extensions
// This allows users to run "clie pwsh" instead of "clie run pwsh".
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
			logging.Debugf("Skipping alias - command already exists: extension=%s", extension.Name)
			continue
		}

		// Create the alias command
		aliasCmd := &cobra.Command{
			Use:   extension.Name,
			Short: fmt.Sprintf("Run %s extension (alias for 'clie run %s')", extension.Name, extension.Name),
			Long: fmt.Sprintf(`This is a convenience alias for 'clie run %s'.

All arguments after the extension name are passed directly to the container.

Examples:
  clie %s                    # Interactive mode
  clie %s --help             # Show extension help
  clie %s echo "Hello"       # Run a command`,
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

		logging.Debugf("Created extension alias command: extension=%s", extension.Name)
	}
}

// InitializeExtensionAliases should be called after config is loaded but before command execution.
func InitializeExtensionAliases() {
	// Try to load config early for alias creation
	// This is best-effort - if it fails, we just won't have aliases
	configFile := os.Getenv(envconsts.EnvCLIEConfig)
	if configFile == "" {
		// Config location: .clie/clie.yml
		possibleConfigs := []string{
			".clie/clie.yml",
			".clie/clie.yaml",
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
			logging.Debugf("Failed to load config for aliases: config=%s error=%v", configFile, err)
		}
	}
}
