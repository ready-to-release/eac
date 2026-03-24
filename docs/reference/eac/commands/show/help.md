# show help

<!-- book:cmd show help -->

## Output Sections

**When called without arguments:** A compact list of all registered command names.

**When called with a command name:** Man-page style output with sections:

- **NAME** - Command name and short description.
- **SYNOPSIS** - Usage pattern with flags and arguments placeholders.
- **DESCRIPTION** - Long description (if available).
- **COMMANDS** - Subcommands (for parent commands like `show`, `get`, `work`).
- **FLAGS** - Each flag with type, required/optional status, default value, and usage.
- **EXAMPLES** - Example invocations.

## See Also

- [help](../help/help.md) - Alternative help command
- [get commands](../get/commands.md) - Command metadata (JSON)
- [show valid-commands](./valid-commands.md)
