// Command: template command
// Short: REPLACE: One sentence describing what this command does (50-80 characters)
// Long: REPLACE: First paragraph explaining the purpose and what problem this command solves.
// Long: REPLACE: Second paragraph describing the behavior and how it works internally.
// Long: REPLACE: Third paragraph describing inputs, outputs, and important details users need to know.
// Long: REPLACE (Optional): Additional paragraphs for default behaviors, validation requirements, or special notes.
// Flag.example: type=string, shorthand=e, default=value, usage=REPLACE: Educational description explaining what this flag does, when to use it, and why it's useful, required=false, completion=value1,value2
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode with verbose logging and intermediate file outputs for troubleshooting
package template

// INSTRUCTIONS:
// 1. Copy this file to go/eac/commands/impl/<category>/<commandname>.go
// 2. Replace all REPLACE: placeholders with actual content
// 3. Update package name to match your command
// 4. Update function names (TemplateCommand -> YourCommand)
// 5. Add/remove flags as needed
// 6. Implement your command logic in the main function
// 7. Delete this comment block
//
// For detailed guidance, see docs/command-template.md

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

func init() {
	// Register this command with the registry
	// The registry will automatically extract help metadata from file header comments
	registry.Register(TemplateCommand)
}

// Config holds configuration for this command
// Add fields for each flag you define
type Config struct {
	Example string // Corresponds to --example flag
	Debug   bool   // Corresponds to --debug flag

	// Add more fields as needed for your command
	WorkspaceRoot string // Repository root (commonly needed)
}

// TemplateCommand is the main entry point for the command
// Return 0 for success, 1 for errors
func TemplateCommand() int {
	// Phase 1: Parse configuration from command line arguments
	config, err := parseConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Phase 2: Validate prerequisites (optional)
	// Example: Check if required files/directories exist
	// Example: Validate configuration values
	if err := validatePrerequisites(config); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Phase 3: Execute command logic
	if err := executeCommand(config); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	return 0
}

// parseConfig parses command line arguments into configuration
// Handles flag parsing and validation
func parseConfig() (*Config, error) {
	config := &Config{
		// Set default values here
		Example: "value", // Default from flag definition
		Debug:   false,   // Default from flag definition
	}

	// Get repository root (if needed)
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w\n\nPlease run this command from within a git repository", err)
	}
	config.WorkspaceRoot = workspaceRoot

	// Parse command line arguments
	// os.Args[0] = program name
	// os.Args[1] = command name (e.g., "template command")
	// os.Args[2:] = flags and arguments
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--example", "-e":
			// String flag: requires a value
			if i+1 < len(args) {
				config.Example = args[i+1]
				i++ // Skip next argument (it's the value)
			} else {
				return nil, fmt.Errorf("--example requires a value")
			}

		case "--debug", "-d":
			// Boolean flag: no value needed
			config.Debug = true

		default:
			// Handle positional arguments or unknown flags
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s\n\nUse 'help template command' to see available flags", arg)
			}
			// If you need positional arguments, handle them here
			// Example: config.PositionalArg = arg
		}
	}

	// Validation: Check required flags
	// Example:
	// if config.Example == "" {
	// 	return nil, fmt.Errorf("--example is required\n\nUsage: template command --example <value>")
	// }

	// Validation: Check value constraints
	// Example:
	// validValues := []string{"value1", "value2", "value3"}
	// if !contains(validValues, config.Example) {
	// 	return nil, fmt.Errorf("invalid value for --example: %s\n\nValid values: %s", config.Example, strings.Join(validValues, ", "))
	// }

	return config, nil
}

// validatePrerequisites checks if prerequisites are met before executing
// This is optional - delete if not needed
func validatePrerequisites(config *Config) error {
	// Example: Check if a required file exists
	// requiredFile := filepath.Join(config.WorkspaceRoot, "path", "to", "file")
	// if _, err := os.Stat(requiredFile); err != nil {
	// 	return fmt.Errorf("required file not found: %s", requiredFile)
	// }

	// Example: Validate that a directory is writable
	// if info, err := os.Stat(config.OutputDir); err != nil || !info.IsDir() {
	// 	return fmt.Errorf("output directory is not accessible: %s", config.OutputDir)
	// }

	return nil
}

// executeCommand contains the main command logic
// Separated from main function for cleaner code organization
func executeCommand(config *Config) error {
	// REPLACE THIS SECTION with your command implementation

	// Example: Display a message
	log.Info("Template command executed successfully!")

	if config.Debug {
		log.Debugf("Debug: Example value = %s", config.Example)
		log.Debugf("Debug: Workspace root = %s", config.WorkspaceRoot)
	}

	// Example: Read a file
	// content, err := os.ReadFile(filepath.Join(config.WorkspaceRoot, "file.txt"))
	// if err != nil {
	// 	return fmt.Errorf("failed to read file: %w", err)
	// }

	// Example: Process data
	// result := processData(content)

	// Example: Write output
	// if err := os.WriteFile(outputPath, result, 0644); err != nil {
	// 	return fmt.Errorf("failed to write output: %w", err)
	// }

	// Example: Display results
	// fmt.Println("Results:")
	// fmt.Println(string(result))

	return nil
}

// Helper function: contains checks if a slice contains a value
// Delete if not needed
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// Additional helper functions as needed
// Keep functions small and focused (following Vibe Coding principles)
