package flags

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

// ValidateFlagsFromRegistry validates command-line flags against registry metadata.
// It automatically detects the calling command using runtime.Caller() and validates
// the provided arguments against the command's registered flag metadata.
//
// Returns an error if:
// - Unable to detect calling command
// - Command not found in registry
// - Unknown flags provided
// - Required flags missing
//
// Example usage:
//
//	func Init() int {
//	    if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
//	        log.Errorf("%v", err)
//	        return 1
//	    }
//	    // ... rest of command logic
//	}
func ValidateFlagsFromRegistry(args []string) error {
	// Auto-detect calling command using runtime.Caller()
	_, callerFile, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("failed to detect calling command")
	}

	// Extract command name from file's // Command: comment
	commandName, err := extractCommandName(callerFile)
	if err != nil {
		return fmt.Errorf("failed to extract command name: %w", err)
	}

	// Get command registration from registry
	cmd := registry.GetCommand(commandName)
	if cmd == nil {
		return fmt.Errorf("command not found in registry: %s", commandName)
	}

	// Parse args into flags map
	parsedFlags := make(map[string]string)
	var positionalArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip positional arguments
		if !strings.HasPrefix(arg, "-") {
			positionalArgs = append(positionalArgs, arg)
			continue
		}

		// Parse flag name and value
		var flagName, flagValue string
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			flagName = parts[0]
			flagValue = parts[1]
		} else {
			flagName = arg
			// Check if next arg is a value
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Look up flag metadata to determine if it takes a value
				flagMeta := findFlagMetadata(cmd.Flags, flagName)
				if flagMeta != nil && flagMeta.Type != "bool" {
					i++
					flagValue = args[i]
				}
			}
		}

		// Normalize flag name (remove leading dashes for lookup)
		normalizedName := strings.TrimLeft(flagName, "-")
		parsedFlags[normalizedName] = flagValue
	}

	// Validate parsed flags against registry metadata
	if err := validateParsedFlags(parsedFlags, cmd.Flags); err != nil {
		return err
	}

	return nil
}

// extractCommandName reads the calling file and extracts the // Command: comment.
// Handles cross-platform paths by converting Windows paths for registry lookups.
func extractCommandName(callerFile string) (string, error) {
	// Translate path for cross-compilation if needed
	translatedPath, err := translateCrossCompilePath(callerFile)
	if err != nil {
		return "", err
	}

	// Read file to find // Command: comment
	content, err := os.ReadFile(translatedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", translatedPath, err)
	}

	// Look for // Command: comment
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "// Command:") {
			commandName := strings.TrimSpace(strings.TrimPrefix(line, "// Command:"))
			return commandName, nil
		}
	}

	return "", fmt.Errorf("no // Command: comment found in file: %s", translatedPath)
}

// translateCrossCompilePath handles path translation for cross-compilation scenarios.
// Windows paths may use backslashes; ensure file exists at expected location.
func translateCrossCompilePath(path string) (string, error) {
	// Normalize path separators for current OS
	normalizedPath := filepath.FromSlash(path)

	// Check if file exists at this path
	if _, err := os.Stat(normalizedPath); err == nil {
		return normalizedPath, nil
	}

	// If not found, try with OS-specific path separators
	if filepath.Separator == '\\' {
		// On Windows, try converting forward slashes
		windowsPath := strings.ReplaceAll(path, "/", "\\")
		if _, err := os.Stat(windowsPath); err == nil {
			return windowsPath, nil
		}
	}

	// Return original path if no translation worked
	return normalizedPath, nil
}

// findFlagMetadata looks up flag metadata by name or shorthand.
func findFlagMetadata(flags []registry.FlagMetadata, flagName string) *registry.FlagMetadata {
	// Remove leading dashes
	normalizedName := strings.TrimLeft(flagName, "-")

	for i := range flags {
		flag := &flags[i]
		// Check full name
		if flag.Name == normalizedName {
			return flag
		}
		// Check shorthand (with or without dash)
		if flag.Shorthand != "" {
			normalizedShorthand := strings.TrimLeft(flag.Shorthand, "-")
			if normalizedShorthand == normalizedName {
				return flag
			}
		}
	}
	return nil
}

// validateParsedFlags validates parsed flags against registry metadata.
func validateParsedFlags(parsedFlags map[string]string, flagMetadata []registry.FlagMetadata) error {
	// Check for unknown flags
	for flagName := range parsedFlags {
		if findFlagMetadata(flagMetadata, flagName) == nil {
			// Generate helpful error with suggestions
			return buildUnknownFlagError(flagName, flagMetadata)
		}
	}

	// Check for required flags
	for _, flagMeta := range flagMetadata {
		if flagMeta.Required {
			if _, found := parsedFlags[flagMeta.Name]; !found {
				return fmt.Errorf("required flag missing: --%s", flagMeta.Name)
			}
		}
	}

	return nil
}

// buildUnknownFlagError creates a helpful error message for unknown flags.
func buildUnknownFlagError(unknownFlag string, validFlags []registry.FlagMetadata) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Unknown flag: --%s\n\n", unknownFlag))

	// List valid flags
	sb.WriteString("Valid flags:\n")
	for _, flag := range validFlags {
		flagStr := fmt.Sprintf("  --%s", flag.Name)
		if flag.Shorthand != "" {
			flagStr += fmt.Sprintf(", %s", flag.Shorthand)
		}
		if flag.Type != "bool" {
			flagStr += fmt.Sprintf(" <%s>", flag.Type)
		}
		if flag.Usage != "" {
			flagStr += fmt.Sprintf(" - %s", flag.Usage)
		}
		sb.WriteString(flagStr + "\n")
	}

	// Suggest similar flags
	suggestions := suggestSimilarFlags(unknownFlag, validFlags)
	if len(suggestions) > 0 {
		sb.WriteString("\nDid you mean:\n")
		for _, suggestion := range suggestions {
			sb.WriteString(fmt.Sprintf("  --%s\n", suggestion))
		}
	}

	return fmt.Errorf("%s", sb.String())
}

// suggestSimilarFlags suggests flags with similar names using Levenshtein distance.
func suggestSimilarFlags(unknownFlag string, validFlags []registry.FlagMetadata) []string {
	const maxDistance = 2
	var suggestions []string

	// Strip leading dashes from unknown flag for comparison
	normalizedUnknown := strings.TrimLeft(unknownFlag, "-")

	for _, flag := range validFlags {
		if distance := levenshteinDistance(normalizedUnknown, flag.Name); distance <= maxDistance {
			suggestions = append(suggestions, flag.Name)
		}
	}

	return suggestions
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(values ...int) int {
	minVal := values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}
