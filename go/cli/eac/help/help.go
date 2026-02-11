// Package help provides unified help printing for CLI commands.
package help

import (
	"fmt"
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// PrintHelp prints help for a command based on its CommandPort metadata.
// For parent commands (IsParent=true), it prints grouped subcommands.
// For leaf commands, it prints flags and long description.
// reg provides the command registry for looking up subcommand descriptions.
func PrintHelp(w io.Writer, cmd core.CommandPort, reg core.CommandRegistryPort) {
	if cmd == nil {
		fmt.Fprintln(w, "No help available.")
		return
	}

	meta := cmd.Metadata()

	// Print command name and short description
	fmt.Fprintf(w, "%s - %s\n", cmd.Name(), meta.Short)
	fmt.Fprintln(w)

	// Print usage
	if meta.IsParent {
		fmt.Fprintf(w, "Usage: clie %s <subcommand> [args...]\n", cmd.Name())
	} else {
		fmt.Fprintf(w, "Usage: clie %s [flags]\n", cmd.Name())
	}
	fmt.Fprintln(w)

	// Print long description if available (for leaf commands)
	if meta.Long != "" {
		fmt.Fprintln(w, meta.Long)
		fmt.Fprintln(w)
	}

	// For parent commands, print grouped subcommands
	if meta.IsParent && len(meta.SubcommandGroups) > 0 {
		for _, group := range meta.SubcommandGroups {
			fmt.Fprintf(w, "%s:\n", group.Name)
			for _, sub := range group.Subcommands {
				// Look up subcommand description
				desc := ""
				if reg != nil {
					subKey := cmd.Name() + " " + sub
					if subCmd, ok := reg.Get(subKey); ok {
						desc = subCmd.Metadata().Short
					}
				}
				if desc != "" {
					fmt.Fprintf(w, "  %-24s %s\n", sub, desc)
				} else {
					fmt.Fprintf(w, "  %s\n", sub)
				}
			}
			fmt.Fprintln(w)
		}
	}

	// Print flags if any
	if len(meta.Flags) > 0 {
		// Separate declarative behavior flags from regular flags
		regularFlags, behaviorGroups := CategorizeFlags(meta.Flags)

		// Print behavior flags first (grouped by behavior)
		PrintBehaviorFlags(w, behaviorGroups)

		// Print regular flags
		if len(regularFlags) > 0 {
			fmt.Fprintln(w, "Flags:")
			for _, flag := range regularFlags {
				shorthand := ""
				if flag.Shorthand != "" {
					shorthand = fmt.Sprintf("-%s, ", flag.Shorthand)
				}
				defaultVal := ""
				if flag.DefaultValue != "" {
					defaultVal = fmt.Sprintf(" (default: %s)", flag.DefaultValue)
				}
				required := ""
				if flag.Required {
					required = " [required]"
				}
				fmt.Fprintf(w, "  %s--%s%s%s\n", shorthand, flag.Name, defaultVal, required)
				if flag.Usage != "" {
					fmt.Fprintf(w, "      %s\n", flag.Usage)
				}
			}
			fmt.Fprintln(w)
		}
	}

	// Print examples if any
	if len(meta.Examples) > 0 {
		fmt.Fprintln(w, "Examples:")
		for _, ex := range meta.Examples {
			fmt.Fprintf(w, "  %s\n", ex)
		}
		fmt.Fprintln(w)
	}

	// Print footer for parent commands
	if meta.IsParent {
		fmt.Fprintf(w, "Use 'clie %s <subcommand> --help' for more information about a command.\n", cmd.Name())
	}
}
