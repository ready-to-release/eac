package main

import (
	"os"

	"github.com/ready-to-release/eac/src/cli/cmd"
)

func main() {
	// Filter out spurious "2" argument that comes from bash redirect "2>&1"
	// This is a known issue where shell redirects get parsed as arguments
	originalArgs := os.Args
	filteredArgs := filterSpuriousArguments(originalArgs)

	// Log if we fixed a bad call
	if len(originalArgs) != len(filteredArgs) {
		// We'll log this after logger is initialized in cmd.Execute()
		os.Setenv("R2R_FIXED_REDIRECT", "true")
		os.Setenv("R2R_ORIGINAL_ARGS", argsToString(originalArgs))
		os.Setenv("R2R_FILTERED_ARGS", argsToString(filteredArgs))
	}

	os.Args = filteredArgs
	cmd.Execute()
}

func argsToString(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

// filterSpuriousArguments removes the spurious "2" that appears when
// commands are executed with bash-style stderr redirection (2>&1).
// This prevents the "2" from being passed to ex. PowerShell scripts
func filterSpuriousArguments(args []string) []string {
	if len(args) == 0 {
		return args
	}

	filtered := make([]string, 0, len(args))

	// Check if this looks like a PowerShell command with redirect pollution
	isPwshCommand := false
	for _, arg := range args {
		if arg == "run" || arg == "pwsh" || arg == "powershell" {
			isPwshCommand = true
			break
		}
	}

	for i, arg := range args {
		// Skip standalone "2" that appears after PowerShell commands
		// This is the telltale sign of "2>&1" redirect pollution
		if arg == "2" && isPwshCommand {
			// Check if this "2" is likely from redirect pollution:
			// 1. It's at the end of arguments
			// 2. OR it appears after common PowerShell flags like -c, -Command, --debug
			if i == len(args)-1 {
				continue
			}
			if i > 0 {
				prevArg := args[i-1]
				// Check if previous arg suggests this "2" is spurious
				if prevArg == "-c" || prevArg == "-Command" || prevArg == "--debug" ||
					prevArg == "-File" || prevArg == "-f" || prevArg == "-q" || prevArg == "--quiet" {
					// This "2" is likely spurious if it's not followed by "&1"
					if i+1 >= len(args) || args[i+1] != "&1" {
						continue
					}
				}
			}
		}

		// Also filter out standalone "&1" that might appear after "2"
		if arg == "&1" && i > 0 && args[i-1] == "2" && isPwshCommand {
			continue
		}

		filtered = append(filtered, arg)
	}

	return filtered
}
