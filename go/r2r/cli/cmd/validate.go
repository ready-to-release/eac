package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/r2r/cli/internal/conf"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
	"github.com/ready-to-release/eac/go/r2r/cli/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	validateStrict     bool
	validateShowSchema bool
)

// validateCmd represents the validate command.
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
			configFile = filepath.Join(repoRoot, ".r2r", "r2r-cli.yml")
		}

		// Check if file exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("configuration file not found: %s", configFile)
		}

		logging.Infof("Validating configuration file: %s", configFile)

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
			logging.Infof("✅ Configuration is valid (schema version: %s)", validator.GetEmbeddedSchemaVersion())
			return nil
		}

		// Display errors
		if len(result.Errors) > 0 {
			logging.Error("❌ Validation errors found:")
			for _, e := range result.Errors {
				if e.Field != "" {
					if e.Expected != "" && e.Expected != e.Rule {
						logging.Errorf("  - %s: %s (expected: %s)", e.Field, e.Message, e.Expected)
					} else {
						logging.Errorf("  - %s: %s", e.Field, e.Message)
					}
				} else {
					if e.Expected != "" && e.Expected != e.Rule {
						logging.Errorf("  - %s (expected: %s)", e.Message, e.Expected)
					} else {
						logging.Errorf("  - %s", e.Message)
					}
				}
			}
		}

		// Display warnings
		if len(result.Warnings) > 0 {
			logging.Warn("⚠️  Validation warnings:")
			for _, w := range result.Warnings {
				if w.Field != "" {
					logging.Warnf("  - %s: %s", w.Field, w.Message)
				} else {
					logging.Warnf("  - %s", w.Message)
				}
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

		logging.Warnf("⚠️  Configuration is valid with %d warning(s)", len(result.Warnings))
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

// showEmbeddedSchema displays information about the embedded schema.
func showEmbeddedSchema() error {
	logging.Info("Embedded Schema Information:")
	logging.Infof("  Version: %s", validator.GetEmbeddedSchemaVersion())
	logging.Info("  Schema ID: r2r-cli-config/v1.0")

	// Get and parse the schema
	schemaStr := validator.GetEmbeddedSchema()
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		return fmt.Errorf("failed to parse embedded schema: %w", err)
	}

	// Display schema details
	if id, ok := schema["$id"].(string); ok {
		logging.Infof("  Full ID: %s", id)
	}
	if title, ok := schema["title"].(string); ok {
		logging.Infof("  Title: %s", title)
	}
	if desc, ok := schema["description"].(string); ok {
		logging.Infof("  Description: %s", desc)
	}

	// Count properties
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		logging.Infof("  Root Properties: %d", len(props))
		logging.Infof("    - %v", getKeys(props))
	}

	// Show required fields
	if required, ok := schema["required"].([]interface{}); ok {
		logging.Infof("  Required Fields: %v", required)
	}

	logging.Info("\nTo see the full schema, check: schemas/r2r-cli-config/v1.0/schema.json")

	return nil
}

// getKeys returns the keys from a map as a slice.
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
