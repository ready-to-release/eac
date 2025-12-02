package cmd

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/r2r/cli/internal/conf"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
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
	Long:  `Creates a minimal .r2r/r2r-cli.yml configuration file in the repository.`,
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
			logging.Errorf("Error: Failed to get current working directory: %v", err)
			os.Exit(1)
		}

		gitPath := filepath.Join(pwd, ".git")
		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			if err := os.Mkdir(gitPath, 0755); err != nil {
				logging.Errorf("Error: Failed to create .git directory: %v", err)
				os.Exit(1)
			}
			logging.Info("💡 Created .git folder to simulate repository root")
		}
	}

	repoRoot, err := conf.FindRepositoryRoot()
	if err != nil {
		logging.Errorf("Error: Not a git repository. %v", err)
		logging.Info("💡 To enable r2r-cli in non-git projects, use: r2r init --use-pwd-as-root")
		os.Exit(1)
	}

	// Check if --delete-configs flag is set
	deleteConfigs, _ := cmd.Flags().GetBool("delete-configs")
	if deleteConfigs {
		deleteConfigFiles(cmd, repoRoot)
		return
	}

	// Ensure .r2r directory exists
	r2rDir := filepath.Join(repoRoot, ".r2r")
	if err := os.MkdirAll(r2rDir, 0755); err != nil {
		logging.Errorf("Error: Failed to create .r2r directory: %v", err)
		os.Exit(1)
	}

	configFile := filepath.Join(r2rDir, "r2r-cli.yml")

	if _, err := os.Stat(configFile); err == nil {
		logging.Error("Error: .r2r/r2r-cli.yml already exists")
		os.Exit(1)
	}

	minimalConfig := `extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/ext-eac:latest'
`

	err = os.WriteFile(configFile, []byte(minimalConfig), 0644)
	if err != nil {
		logging.Errorf("Error: Failed to create .r2r/r2r-cli.yml: %v", err)
		os.Exit(1)
	}

	logging.Infof("Created %s", configFile)
}

func deleteConfigFiles(cmd *cobra.Command, repoRoot string) {
	// List of config files in .r2r directory including overrides (but not examples)
	r2rDir := filepath.Join(repoRoot, ".r2r")
	configFiles := []string{
		"r2r-cli.yml",
		"r2r-cli.local.yml",
		"r2r-cli.personal.yml",
		"r2r-cli.dev.yml",
	}

	deletedCount := 0
	for _, configFile := range configFiles {
		configPath := filepath.Join(r2rDir, configFile)

		if _, err := os.Stat(configPath); err == nil {
			// File exists, delete it
			if err := os.Remove(configPath); err != nil {
				logging.Warnf("Warning: Failed to delete %s: %v", configFile, err)
				continue
			}
			logging.Infof("Deleted %s", configPath)
			deletedCount++
		}
	}

	if deletedCount == 0 {
		logging.Infof("No configuration files found to delete in %s", r2rDir)
	} else {
		logging.Infof("Deleted %d configuration file(s) from %s", deletedCount, r2rDir)
	}
}
