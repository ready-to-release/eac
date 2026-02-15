# help

Displays help information for commands via the `show help` command, providing both a compact list of all commands and detailed help for specific commands.

## Key Functions

- `ShowHelp` -- Entry point for the `show help` command; parses arguments to show either all commands or detailed help for a specific command
- `showAllCommands` -- Lists all commands alphabetically using the compact list renderer
- `showCommandHelp` -- Displays detailed help for a specific command with NAME, SYNOPSIS, DESCRIPTION, COMMANDS, FLAGS, and EXAMPLES sections
- `showParentHelp` -- Displays help for a command prefix that has subcommands but no direct registration

## Patterns

- **Command registration via init()**: Uses `registry.Register(ShowHelp)` for automatic command discovery
- **Man-page style output**: Formats detailed help in traditional sections (NAME, SYNOPSIS, DESCRIPTION, COMMANDS, FLAGS, EXAMPLES)
- **Compact list rendering**: Uses the `render.RenderCompactList` function for the all-commands view rather than category grouping
- **Parent command fallback**: When a command name isn't directly registered but has subcommands, displays the parent help with available subcommands

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | All package functionality: ShowHelp entry point, showAllCommands, showCommandHelp, showParentHelp, buildSynopsis, displayFlag, getSubcommands |

## Dependencies

- `go/clibase/flags` -- Shared flag validation
- `go/clibase/render` -- Compact list rendering for all-commands view
- `go/clibase/registry` -- Command registry access for listing commands and their metadata
- `go/core/logging` -- Component logger for info and error output

## Role in System

This package implements the `show help` command, providing an alternative entry point to the help system accessible via the `show` command family. While the standalone `help` command groups by category, `show help` uses a compact alphabetical list for the overview and adds an EXAMPLES section to detailed command help. Both serve the same fundamental purpose of documenting available CLI commands.

## Code Health

### Tech Debt
- No test for commands.go (277 lines)

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
