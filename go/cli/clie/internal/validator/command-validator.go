// Package validator provides validation for clie commands
// It uses the command parser to parse arguments and then validates them
// against business rules and the EBNF schema.
package validator

import (
	"fmt"
	"log/slog"
	"strings"

	parser "github.com/ready-to-release/eac/go/cli/clie/internal/command-parser"
)

// CommandValidator validates parsed commands against business rules.
type CommandValidator struct {
	parser *parser.Parser
}

// CommandValidationResult contains validation results.
type CommandValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string

	// Parsed command structure
	ParsedCommand *parser.ParsedCommand
}

// NewCommandValidator creates a new command validator.
func NewCommandValidator() *CommandValidator {
	return &CommandValidator{
		parser: parser.NewParser(),
	}
}

// ValidateCommand validates a command-line against schema and business rules.
func (cv *CommandValidator) ValidateCommand(args []string) *CommandValidationResult {
	result := &CommandValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// 1. Check for empty command
	if len(args) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "No command provided")
		return result
	}

	// 2. Parse the command
	parsed := cv.parser.Parse(args)
	result.ParsedCommand = parsed

	// 3. Validate binary name
	if !cv.parser.IsValidBinaryName(parsed.BinaryName) {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("Invalid binary name: %s", parsed.BinaryName))
		return result
	}

	// 4. Validate global flags for conflicts
	if cv.parser.HasConflictingGlobalFlags(parsed.GlobalFlags) {
		result.Valid = false
		result.Errors = append(result.Errors,
			"Cannot use both --clie-debug and --clie-quiet flags")
		return result
	}

	// 5. Validate subcommand if present
	if parsed.Subcommand != "" && !cv.parser.IsValidSubcommand(parsed.Subcommand) {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("Invalid subcommand: %s", parsed.Subcommand))
		return result
	}

	// 6. Validate extension name requirements
	if cv.parser.RequiresExtension(parsed.Subcommand) {
		if parsed.ExtensionName == "" {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s command requires an extension name", parsed.Subcommand))
			return result
		}

		if !cv.parser.IsValidExtensionName(parsed.ExtensionName) {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("Invalid extension name format: %s", parsed.ExtensionName))
			return result
		}
	}

	// 7. Add warnings for potential issues
	cv.addWarnings(result, parsed, args)

	return result
}

// addWarnings adds non-fatal warnings to the validation result.
func (cv *CommandValidator) addWarnings(result *CommandValidationResult, parsed *parser.ParsedCommand, _ []string) {
	// Warn if global flags appear in container args (might be unintentional)
	for _, arg := range parsed.ContainerArgs {
		if cv.parser.IsGlobalFlag(arg) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Global flag '%s' found in container arguments - will be passed to container, not processed by clie", arg))
		}
	}

	// Warn about unrecognized flags in Viper args that are not known global flags.
	// The EBNF grammar does not define subcommand-specific flags, so any flag in
	// Viper args that is not a global flag is flagged as unrecognized.
	for _, arg := range parsed.ViperArgs {
		if strings.HasPrefix(arg, "-") && !cv.parser.IsGlobalFlag(arg) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Unrecognized flag '%s' in CLI arguments - will be passed through to Cobra", arg))
		}
	}

	// Warn if no subcommand (will show help)
	if parsed.Subcommand == "" && len(parsed.GlobalFlags) == 0 {
		result.Warnings = append(result.Warnings,
			"No subcommand specified - help will be displayed")
	}
}

// GetViperArguments returns arguments that should be processed by Viper.
func (cv *CommandValidator) GetViperArguments(args []string) []string {
	parsed := cv.parser.Parse(args)
	return parsed.ViperArgs
}

// GetContainerArguments returns arguments that should be passed to container.
func (cv *CommandValidator) GetContainerArguments(args []string) []string {
	parsed := cv.parser.Parse(args)
	return parsed.ContainerArgs
}

// GetArgumentBoundary returns the index where container args start.
func (cv *CommandValidator) GetArgumentBoundary(args []string) int {
	parsed := cv.parser.Parse(args)
	return parsed.ArgumentBoundary
}

// IsViperArgument checks if an argument at position should be processed by Viper.
func (cv *CommandValidator) IsViperArgument(args []string, position int) bool {
	boundary := cv.GetArgumentBoundary(args)
	if boundary == -1 {
		// No container args, everything is Viper
		return true
	}
	return position < boundary
}

// ValidateExtensionName validates an extension name format.
func (cv *CommandValidator) ValidateExtensionName(name string) error {
	if !cv.parser.IsValidExtensionName(name) {
		return fmt.Errorf("invalid extension name format: %s", name)
	}
	return nil
}

// ValidateForRun performs specific validation for run command.
func (cv *CommandValidator) ValidateForRun(args []string) *CommandValidationResult {
	result := cv.ValidateCommand(args)

	if result.Valid && result.ParsedCommand.Subcommand == "run" {
		// Additional run-specific validation
		if result.ParsedCommand.ExtensionName == "" {
			result.Valid = false
			result.Errors = append(result.Errors,
				"run command requires an extension name")
		}

		// Log the argument split for debugging
		slog.Debug("run command argument split",
			"extension", result.ParsedCommand.ExtensionName,
			"viper_args", result.ParsedCommand.ViperArgs,
			"container_args", result.ParsedCommand.ContainerArgs,
			"boundary", result.ParsedCommand.ArgumentBoundary,
		)
	}

	return result
}

// ValidateForMetadata performs specific validation for metadata command.
func (cv *CommandValidator) ValidateForMetadata(args []string) *CommandValidationResult {
	result := cv.ValidateCommand(args)

	if result.Valid && result.ParsedCommand.Subcommand == "metadata" {
		// Additional metadata-specific validation
		if result.ParsedCommand.ExtensionName == "" {
			result.Valid = false
			result.Errors = append(result.Errors,
				"metadata command requires an extension name")
		}
	}

	return result
}

// Summary returns a human-readable summary of validation.
func (result *CommandValidationResult) Summary() string {
	if result.Valid {
		if len(result.Warnings) > 0 {
			return fmt.Sprintf("Command valid with %d warning(s)", len(result.Warnings))
		}
		return "Command valid"
	}
	return fmt.Sprintf("Command invalid: %d error(s), %d warning(s)",
		len(result.Errors), len(result.Warnings))
}
