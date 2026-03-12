# Show Help

<!-- book:cmd show help -->

Displays help information for EAC commands. Without arguments, lists all available commands. With a command name, shows detailed help including description, flags, and usage examples.

## Usage

```bash
eac show help [command...] [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--verbose`, `-v` | bool | Show detailed information including canonical name and all subcommands |

## Output Sections

**When called without arguments:** A compact list of all registered command names.

**When called with a command name:** Man-page style output with sections:

- **NAME** - Command name and short description.
- **SYNOPSIS** - Usage pattern with flags and arguments placeholders.
- **DESCRIPTION** - Long description (if available).
- **COMMANDS** - Subcommands (for parent commands like `show`, `get`, `work`).
- **FLAGS** - Each flag with type, required/optional status, default value, and usage.
- **EXAMPLES** - Example invocations.

## Examples

```bash
# List all commands
eac show help

# Help for a specific command
eac show help show modules

# Help for a parent command
eac show help build

# Verbose mode with extra metadata
eac show help show config --verbose
```

## See Also

- [help](../help/help.md) - Alternative help command
- [get commands](../get/commands.md) - Command metadata (JSON)
- [show valid-commands](./valid-commands.md)
