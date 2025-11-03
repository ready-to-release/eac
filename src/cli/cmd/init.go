package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(InitCmd)

	// Add --delete-configs flag
	InitCmd.Flags().Bool("delete-configs", false, "Delete all configuration files including overrides from repository root")
	// Add --use-pwd-as-root flag
	InitCmd.Flags().Bool("use-pwd-as-root", false, "Use current directory as repository root (creates .git folder if needed)")
}

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new r2r-cli configuration file",
	Long:  `Creates a minimal r2r-cli.yml configuration file in the repository root.`,
	Run: func(cmd *cobra.Command, args []string) {
		createConfigFile(cmd)
	},
}

func createConfigFile(cmd *cobra.Command) {
	// Check if --use-pwd-as-root flag is set
	usePwdAsRoot, _ := cmd.Flags().GetBool("use-pwd-as-root")

	if usePwdAsRoot {
		// Create .git folder in current directory if it doesn't exist
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get current working directory: %v\n", err)
			os.Exit(1)
		}

		gitPath := filepath.Join(pwd, ".git")
		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			if err := os.Mkdir(gitPath, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to create .git directory: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "💡 Created .git folder to simulate repository root\n")
		}
	}

	repoRoot, err := conf.FindRepositoryRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Not a git repository. %v\n", err)
		fmt.Fprintf(os.Stderr, "💡 To enable r2r-cli in non-git projects, use: r2r init --use-pwd-as-root\n")
		os.Exit(1)
	}

	// Check if --delete-configs flag is set
	deleteConfigs, _ := cmd.Flags().GetBool("delete-configs")
	if deleteConfigs {
		deleteConfigFiles(cmd, repoRoot)
		return
	}

	configFile := filepath.Join(repoRoot, "r2r-cli.yml")

	if _, err := os.Stat(configFile); err == nil {
		fmt.Fprintf(os.Stderr, "Error: r2r-cli.yml already exists in the repository root\n")
		os.Exit(1)
	}

	minimalConfig := `extensions:
  - name: 'pwsh'
    image: 'ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:latest'
`

	err = os.WriteFile(configFile, []byte(minimalConfig), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create r2r-cli.yml: %v\n", err)
		os.Exit(1)
	}

	cmd.Printf("Created %s\n", configFile)
}

func deleteConfigFiles(cmd *cobra.Command, repoRoot string) {
	// List of config files including overrides (but not examples)
	configFiles := []string{
		"r2r-cli.yml",
		"r2r-cli.local.yml",
		"r2r-cli.personal.yml",
		"r2r-cli.dev.yml",
	}

	deletedCount := 0
	for _, configFile := range configFiles {
		configPath := filepath.Join(repoRoot, configFile)

		if _, err := os.Stat(configPath); err == nil {
			// File exists, delete it
			if err := os.Remove(configPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to delete %s: %v\n", configFile, err)
				continue
			}
			cmd.Printf("Deleted %s\n", configPath)
			deletedCount++
		}
	}

	if deletedCount == 0 {
		cmd.Printf("No configuration files found to delete in %s\n", repoRoot)
	} else {
		cmd.Printf("Deleted %d configuration file(s) from %s\n", deletedCount, repoRoot)
	}
}
