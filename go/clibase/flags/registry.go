package flags

import (
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// globalRegistry is set by the binary's main() to enable flag validation.
var globalRegistry core.CommandRegistryPort

// SetRegistry sets the global command registry used by ValidateFlagsFromRegistry.
func SetRegistry(reg core.CommandRegistryPort) {
	globalRegistry = reg
}

// GlobalFlags returns flags that are valid for all commands.
// Returns a fresh copy each call so callers cannot mutate shared state.
func GlobalFlags() []core.FlagSpec {
	return []core.FlagSpec{
		{Name: "help", Shorthand: "h", Type: "bool", Usage: "Show help for command"},
	}
}

// ValidateFlagsFromRegistry validates command-line flags against the global command registry.
// It resolves the current command from os.Args using longest-match, looks up its
// flag metadata, and validates the provided arguments.
// Global flags (--help, -h) are always accepted.
func ValidateFlagsFromRegistry(args []string) error {
	if globalRegistry == nil {
		return nil // No registry configured, skip validation
	}

	// Resolve command name from os.Args using longest-match
	cmdName := resolveCommandFromArgs()
	if cmdName == "" {
		return nil // Could not determine command, skip validation
	}

	return ValidateFlagsForCommand(args, cmdName)
}

// ValidateFlagsForCommand validates command-line flags against the registry using
// an explicit command name. This avoids reliance on os.Args for command resolution.
func ValidateFlagsForCommand(args []string, commandName string) error {
	if globalRegistry == nil {
		return nil // No registry configured, skip validation
	}

	cmd, ok := globalRegistry.Get(commandName)
	if !ok {
		return nil // Command not in registry, skip validation
	}

	return ValidateFlags(args, cmd.Metadata().Flags)
}

// ValidateFlags validates command-line args against the given flag specs.
func ValidateFlags(args []string, specs []core.FlagSpec) error {
	allFlags := append(specs, GlobalFlags()...)

	// Parse args into flags map
	parsedFlags := make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip positional arguments
		if !strings.HasPrefix(arg, "-") {
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
				flagMeta := findFlag(allFlags, flagName)
				if flagMeta != nil && flagMeta.Type != "bool" {
					i++
					flagValue = args[i]
				}
			}
		}

		normalizedName := strings.TrimLeft(flagName, "-")
		parsedFlags[normalizedName] = flagValue
	}

	return validateParsedFlags(parsedFlags, allFlags)
}

// resolveCommandFromArgs finds the longest matching command name from os.Args.
func resolveCommandFromArgs() string {
	if len(os.Args) < 2 {
		return ""
	}
	for count := len(os.Args) - 1; count >= 1; count-- {
		name := strings.Join(os.Args[1:count+1], " ")
		if _, ok := globalRegistry.Get(name); ok {
			return name
		}
	}
	return ""
}

// findFlag looks up a flag spec by name or shorthand.
func findFlag(flags []core.FlagSpec, flagName string) *core.FlagSpec {
	normalizedName := strings.TrimLeft(flagName, "-")

	for i := range flags {
		flag := &flags[i]
		if flag.Name == normalizedName {
			return flag
		}
		if flag.Shorthand != "" {
			normalizedShorthand := strings.TrimLeft(flag.Shorthand, "-")
			if normalizedShorthand == normalizedName {
				return flag
			}
		}
	}
	return nil
}

// validateParsedFlags validates parsed flags against flag specs.
func validateParsedFlags(parsedFlags map[string]string, specs []core.FlagSpec) error {
	for flagName := range parsedFlags {
		if findFlag(specs, flagName) == nil {
			return buildUnknownFlagError(flagName, specs)
		}
	}

	for _, spec := range specs {
		if spec.Required {
			if _, found := parsedFlags[spec.Name]; !found {
				return fmt.Errorf("required flag missing: --%s", spec.Name)
			}
		}
	}

	return nil
}

// buildUnknownFlagError creates a helpful error message for unknown flags.
func buildUnknownFlagError(unknownFlag string, validFlags []core.FlagSpec) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Unknown flag: --%s\n\n", unknownFlag))

	sb.WriteString("Valid flags:\n")
	for _, flag := range validFlags {
		flagStr := fmt.Sprintf("  --%s", flag.Name)
		if flag.Shorthand != "" {
			flagStr += fmt.Sprintf(", -%s", flag.Shorthand)
		}
		if flag.Type != "bool" {
			flagStr += fmt.Sprintf(" <%s>", flag.Type)
		}
		if flag.Usage != "" {
			flagStr += fmt.Sprintf(" - %s", flag.Usage)
		}
		sb.WriteString(flagStr + "\n")
	}

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
func suggestSimilarFlags(unknownFlag string, validFlags []core.FlagSpec) []string {
	const maxDistance = 2
	var suggestions []string

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
	if a == "" {
		return len(b)
	}
	if b == "" {
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
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
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
