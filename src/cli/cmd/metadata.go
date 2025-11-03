package cmd

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/docker"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(MetadataCmd)
}

var MetadataCmd = &cobra.Command{
	Use:   "metadata <extension>",
	Short: "Retrieve metadata from an extension",
	Long:  `Retrieve metadata from an extension by executing its extension-meta command.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf.InitConfig()

		// Create container host
		host, err := docker.NewContainerHost()
		if err != nil {
			cmd.PrintErrf("Error creating container host: %v\n", err)
			os.Exit(1)
		}
		defer host.Close()

		// Validate extensions
		if err := host.ValidateExtensions(); err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			os.Exit(1)
		}

		// Find extension
		extensionName := args[0]
		ext, err := host.FindExtension(extensionName)
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			os.Exit(1)
		}

		cmd.PrintErrln("Retrieving metadata from extension:", ext.Name)
		cmd.PrintErrln("Image:", ext.Image)

		// Execute metadata command
		output, err := host.ExecuteMetadataCommand(ext)
		if err != nil {
			cmd.PrintErrf("Error retrieving metadata: %v\n", err)
			os.Exit(1)
		}

		// Output the metadata to stdout
		fmt.Fprint(cmd.OutOrStdout(), output)
	},
}
