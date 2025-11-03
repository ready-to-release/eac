package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/repository/definitions"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	definitionsPath string
)

func init() {
	DefinitionsCmd.Flags().StringVar(&definitionsPath, "path", "", "Path to process definitions from (defaults to repository root)")
	RootCmd.AddCommand(DefinitionsCmd)
}

var DefinitionsCmd = &cobra.Command{
	Use:   "definitions",
	Short: "Process and merge definitions.yml files from the repository",
	Long:  `Recursively finds all definitions.yml files in the repository and merges them into a single YAML structure.`,
	Run: func(cmd *cobra.Command, args []string) {
		processDefinitions(cmd)
	},
}

func processDefinitions(cmd *cobra.Command) {
	var targetPath string

	if definitionsPath != "" {
		// Use provided path
		absPath, err := filepath.Abs(definitionsPath)
		if err != nil {
			log.Error().Err(err).Str("path", definitionsPath).Msg("Failed to resolve absolute path")
			os.Exit(1)
		}

		// Check if path exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			log.Error().Str("path", absPath).Msg("Path does not exist")
			os.Exit(1)
		}

		targetPath = absPath
		log.Debug().Str("path", targetPath).Msg("Using provided path")
	} else {
		// Use repository root
		repoRoot, err := conf.FindRepositoryRoot()
		if err != nil {
			log.Error().Err(err).Msg("Failed to find repository root. Please run this command from within a git repository.")
			os.Exit(1)
		}
		targetPath = repoRoot
		log.Debug().Str("path", targetPath).Msg("Using repository root")
	}

	log.Info().Str("root", targetPath).Msg("Processing definition files")

	merged, err := definitions.ProcessDirectory(targetPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process definition files")
		os.Exit(1)
	}

	yamlOutput, err := yaml.Marshal(merged)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal merged YAML")
		os.Exit(1)
	}

	fmt.Fprint(cmd.OutOrStdout(), string(yamlOutput))
}
