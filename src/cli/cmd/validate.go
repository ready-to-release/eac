package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/ready-to-release/eac/src/cli/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	validateStrict     bool
	validateShowSchema bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate r2r-cli.yml configuration file",
	Long: `Validate an r2r-cli.yml configuration file against the embedded schema.

This command checks your configuration file for:
  - Required fields (extensions array, name, image)
  - Valid field patterns (extension names, environment variables)
  - Valid enum values (image_pull_policy, network_mode)
  - Resource limits and port ranges
  - Duplicate extension names

Examples:
  # Validate the default configuration file
  r2r validate

  # Validate a specific configuration file
  r2r validate ./r2r-cli.local.yml

  # Use strict validation mode (warnings become errors)
  r2r validate --strict

  # Show the embedded schema version and details
  r2r validate --show-schema`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --show-schema flag
		if validateShowSchema {
			return showEmbeddedSchema()
		}

		// Determine config file to validate
		var configFile string
		if len(args) > 0 {
			configFile = args[0]
		} else {
			// Use default config file discovery
			repoRoot, err := conf.FindRepositoryRoot()
			if err != nil {
				return fmt.Errorf("no configuration file specified and could not find repository root: %w", err)
			}
			configFile = filepath.Join(repoRoot, "r2r-cli.yml")
		}

		// Check if file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("configuration file not found: %s", configFile)
		}

		fmt.Printf("Validating configuration file: %s\n", configFile)

		// Load the configuration using viper
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read configuration file: %w", err)
		}

		// Get the raw configuration as a map (preserves field names)
		configMap := viper.AllSettings()

		// Create validator
		v, err := validator.NewEmbeddedValidator()
		if err != nil {
			return fmt.Errorf("failed to initialize validator: %w", err)
		}

		// Validate the configuration using the raw map
		result, err := v.ValidateInterface(configMap)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		// Display results
		if result.IsValid() && len(result.Warnings) == 0 {
			fmt.Printf("✅ Configuration is valid (schema version: %s)\n", validator.GetEmbeddedSchemaVersion())
			return nil
		}

		// Display errors
		if len(result.Errors) > 0 {
			fmt.Println("❌ Validation errors found:")
			for _, e := range result.Errors {
				if e.Field != "" {
					fmt.Printf("  - %s: %s", e.Field, e.Message)
				} else {
					fmt.Printf("  - %s", e.Message)
				}
				if e.Expected != "" && e.Expected != e.Rule {
					fmt.Printf(" (expected: %s)", e.Expected)
				}
				fmt.Println()
			}
		}

		// Display warnings
		if len(result.Warnings) > 0 {
			fmt.Println("⚠️  Validation warnings:")
			for _, w := range result.Warnings {
				if w.Field != "" {
					fmt.Printf("  - %s: %s", w.Field, w.Message)
				} else {
					fmt.Printf("  - %s", w.Message)
				}
				fmt.Println()
			}
		}

		// In strict mode, warnings are treated as errors
		if validateStrict && len(result.Warnings) > 0 {
			return fmt.Errorf("validation failed in strict mode: %d error(s), %d warning(s)",
				len(result.Errors), len(result.Warnings))
		}

		if !result.IsValid() {
			return fmt.Errorf("validation failed: %d error(s)", len(result.Errors))
		}

		fmt.Printf("⚠️  Configuration is valid with %d warning(s)\n", len(result.Warnings))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)

	validateCmd.Flags().BoolVar(&validateStrict, "strict", false,
		"Treat warnings as errors")
	validateCmd.Flags().BoolVar(&validateShowSchema, "show-schema", false,
		"Display the embedded schema information")
}

// showEmbeddedSchema displays information about the embedded schema
func showEmbeddedSchema() error {
	fmt.Printf("Embedded Schema Information:\n")
	fmt.Printf("  Version: %s\n", validator.GetEmbeddedSchemaVersion())
	fmt.Printf("  Schema ID: r2r-cli-config/v1.0\n")

	// Get and parse the schema
	schemaStr := validator.GetEmbeddedSchema()
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		return fmt.Errorf("failed to parse embedded schema: %w", err)
	}

	// Display schema details
	if id, ok := schema["$id"].(string); ok {
		fmt.Printf("  Full ID: %s\n", id)
	}
	if title, ok := schema["title"].(string); ok {
		fmt.Printf("  Title: %s\n", title)
	}
	if desc, ok := schema["description"].(string); ok {
		fmt.Printf("  Description: %s\n", desc)
	}

	// Count properties
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		fmt.Printf("  Root Properties: %d\n", len(props))
		fmt.Printf("    - %v\n", getKeys(props))
	}

	// Show required fields
	if required, ok := schema["required"].([]interface{}); ok {
		fmt.Printf("  Required Fields: %v\n", required)
	}

	fmt.Println("\nTo see the full schema, check: schemas/r2r-cli-config/v1.0/schema.json")

	return nil
}

// getKeys returns the keys from a map as a slice
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
