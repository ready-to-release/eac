// Package help provides unified help printing for CLI commands.
package help

import (
	"fmt"
	"io"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

// PrintHelp prints help for a command based on its registration metadata.
// For parent commands (IsParent=true), it prints grouped subcommands.
// For leaf commands, it prints flags and long description.
// subRegs provides optional subcommand registrations for looking up descriptions.
func PrintHelp(w io.Writer, reg *registry.CommandRegistration, subRegs map[string]*registry.CommandRegistration) {
	if reg == nil {
		fmt.Fprintln(w, "No help available.")
		return
	}

	// Print command name and short description
	fmt.Fprintf(w, "%s - %s\n", reg.ActualCommand, reg.Short)
	fmt.Fprintln(w)

	// Print usage
	if reg.IsParent {
		fmt.Fprintf(w, "Usage: r2r %s <subcommand> [args...]\n", reg.ActualCommand)
	} else {
		fmt.Fprintf(w, "Usage: r2r %s [flags]\n", reg.ActualCommand)
	}
	fmt.Fprintln(w)

	// Print long description if available (for leaf commands)
	if reg.Long != "" {
		fmt.Fprintln(w, reg.Long)
		fmt.Fprintln(w)
	}

	// For parent commands, print grouped subcommands
	if reg.IsParent && len(reg.Subcommands) > 0 {
		for _, group := range reg.Subcommands {
			fmt.Fprintf(w, "%s:\n", group.Name)
			for _, sub := range group.Subcommands {
				// Look up subcommand description
				desc := ""
				if subRegs != nil {
					subKey := reg.ActualCommand + " " + sub
					if subReg, ok := subRegs[subKey]; ok && subReg != nil {
						desc = subReg.Short
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
	if len(reg.Flags) > 0 {
		fmt.Fprintln(w, "Flags:")
		for _, flag := range reg.Flags {
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

	// Print examples if any
	if len(reg.Examples) > 0 {
		fmt.Fprintln(w, "Examples:")
		for _, ex := range reg.Examples {
			fmt.Fprintf(w, "  %s\n", ex)
		}
		fmt.Fprintln(w)
	}

	// Print footer for parent commands
	if reg.IsParent {
		fmt.Fprintf(w, "Use 'r2r %s <subcommand> --help' for more information about a command.\n", reg.ActualCommand)
	}
}
