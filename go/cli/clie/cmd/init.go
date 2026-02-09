package cmd

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/cli/clie/internal/conf"
	"github.com/ready-to-release/eac/go/cli/clie/internal/logging"
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
	Short: "Initialize a new clie configuration file",
	Long:  `Creates a minimal .clie/clie.yml configuration file in the repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		createConfigFile(cmd)
	},
}

func createConfigFile(cmd *cobra.Command) {
	// Check if --use-pwd-as-root flag is set
	usePwdAsRoot, _ := cmd.Flags().GetBool("use-pwd-as-root") //nolint:errcheck // flag registered in init

	if usePwdAsRoot {
		// Create .git folder in current directory if it doesn't exist
		pwd, err := os.Getwd()
		if err != nil {
			logging.Errorf("Error: Failed to get current working directory: %v", err)
			os.Exit(1)
		}

		gitPath := filepath.Join(pwd, ".git")
		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			if err := os.Mkdir(gitPath, 0o755); err != nil { //nolint:gosec // G301: .git dirs need 0755 for git compatibility
				logging.Errorf("Error: Failed to create .git directory: %v", err)
				os.Exit(1)
			}
			logging.Info("💡 Created .git folder to simulate repository root")
		}
	}

	repoRoot, err := conf.FindRepositoryRoot()
	if err != nil {
		logging.Errorf("Error: Not a git repository. %v", err)
		logging.Info("💡 To enable clie in non-git projects, use: clie init --use-pwd-as-root")
		os.Exit(1)
	}

	// Check if --delete-configs flag is set
	deleteConfigs, _ := cmd.Flags().GetBool("delete-configs") //nolint:errcheck // flag registered in init
	if deleteConfigs {
		deleteConfigFiles(cmd, repoRoot)
		return
	}

	// Ensure .clie directory exists
	clieDir := filepath.Join(repoRoot, ".clie")
	if err := os.MkdirAll(clieDir, 0o755); err != nil { //nolint:gosec // G301: config dir needs to be accessible
		logging.Errorf("Error: Failed to create .clie directory: %v", err)
		os.Exit(1)
	}

	configFile := filepath.Join(clieDir, "clie.yml")

	if _, err := os.Stat(configFile); err == nil {
		logging.Error("Error: .clie/clie.yml already exists")
		os.Exit(1)
	}

	minimalConfig := `extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
`

	err = os.WriteFile(configFile, []byte(minimalConfig), 0o644)
	if err != nil {
		logging.Errorf("Error: Failed to create .clie/clie.yml: %v", err)
		os.Exit(1)
	}

	logging.Infof("Created %s", configFile)
}

func deleteConfigFiles(cmd *cobra.Command, repoRoot string) {
	// List of config files in .clie directory including overrides (but not examples)
	clieDir := filepath.Join(repoRoot, ".clie")
	configFiles := []string{
		"clie.yml",
		"clie.local.yml",
		"clie.personal.yml",
		"clie.dev.yml",
	}

	deletedCount := 0
	for _, configFile := range configFiles {
		configPath := filepath.Join(clieDir, configFile)

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
		logging.Infof("No configuration files found to delete in %s", clieDir)
	} else {
		logging.Infof("Deleted %d configuration file(s) from %s", deletedCount, clieDir)
	}
}
