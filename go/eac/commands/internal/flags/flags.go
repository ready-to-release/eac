// Package flags provides shared flag parsing utilities for commands.
package flags

// ParseDebugFlag checks if the debug flag is present in the arguments.
// It accepts both --debug and -d shorthand.
//
// Example:
//
//	args := os.Args[2:] // Skip program name and command
//	debug := flags.ParseDebugFlag(args)
func ParseDebugFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--debug" || arg == "-d" {
			return true
		}
	}
	return false
}

// HasFlag checks if a specific flag is present in the arguments.
// It accepts both the full flag name and optional shorthand.
//
// Example:
//
//	hasAll := flags.HasFlag(args, "--all", "-a")
func HasFlag(args []string, flag string, shorthand string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
		if shorthand != "" && arg == shorthand {
			return true
		}
	}
	return false
}

// GetFlagValue retrieves the value for a flag in --flag=value format.
// Returns empty string if not found.
//
// Example:
//
//	target := flags.GetFlagValue(args, "--target")
func GetFlagValue(args []string, prefix string) string {
	for _, arg := range args {
		if len(arg) > len(prefix) && arg[:len(prefix)] == prefix && arg[len(prefix)] == '=' {
			return arg[len(prefix)+1:]
		}
	}
	return ""
}

// GetPositionalArgs returns non-flag arguments.
// Arguments starting with -- or - are considered flags.
//
// Example:
//
//	positional := flags.GetPositionalArgs(args)
func GetPositionalArgs(args []string) []string {
	var positional []string
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}
		// Skip flags
		if arg[0] == '-' {
			continue
		}
		positional = append(positional, arg)
	}
	return positional
}
