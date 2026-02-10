# help

Displays help information for commands, either listing all available commands grouped by category or showing detailed help for a specific command.

## Key Functions

- `Help` -- Entry point for the `help` command; parses arguments to show either all commands or detailed help for a specific command
- `showAllCommands` -- Lists all commands grouped by category (first word), with standalone commands in a separate section
- `showCommandHelp` -- Displays detailed help for a specific command with NAME, SYNOPSIS, DESCRIPTION, COMMANDS (subcommands), and FLAGS sections

## Patterns

- **Command registration via init()**: Uses `registry.Register(Help)` for automatic command discovery
- **Man-page style output**: Formats help in traditional Unix man-page sections (NAME, SYNOPSIS, DESCRIPTION, COMMANDS, FLAGS)
- **Category grouping**: Groups commands by their first word (e.g., all "show" commands together, all "create" commands together)
- **Subcommand discovery**: When a command name has no direct registration but has subcommands, displays the category with its available subcommands
- **Word wrapping**: Flag usage text is wrapped at 76 characters for readable terminal output

## Internal Structure

| File | Responsibility |
| --- | --- |
| help.go | All package functionality: Help entry point, showAllCommands, showCommandHelp, buildSynopsis, displayFlag, wrapText, getSubcommands |

## Dependencies

- `go/clibase/flags` -- Shared flag validation
- `go/clibase/registry` -- Command registry access for listing commands and their metadata
- `go/core/logging` -- Component logger for info and error output

## Role in System

This package implements the `help` command, providing the primary user-facing documentation system for the eac CLI. It serves as both a command directory (listing all available commands) and a detailed reference (showing flags, descriptions, and subcommands for specific commands). It reads all metadata from the command registry, ensuring help output always reflects the current set of registered commands.

## Code Health

### Tech Debt
- None

### Pain Points
- None identified

### Optimization Opportunities
- None identified
