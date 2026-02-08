# Get Help with Commands

## What You'll Accomplish

Discover available EAC commands and learn how to use them effectively using the built-in help system.

## Prerequisites

### Required Knowledge

**New to eac?** Learn these concepts first:

- [Quick Start Guide](../../../../tutorials/getting-started/quick-start.md) - Install eac and understand basic commands

### Required Setup

- EAC installed and accessible via `eac`

## Steps

### 1. Show General Help

```bash
eac help
```

**What happens**: Displays all available command categories and common commands

### 2. Get Help for Specific Command

```bash
eac help build
```

**What happens**: Shows detailed syntax, options, and examples for the build command

### 3. List All Available Commands

```bash
eac show help
```

**What happens**: Displays formatted table of all commands with descriptions

### 4. Get Machine-Readable Command List

```bash
eac get commands | jq '.commands[] | {name, description}'
```

**What happens**: Returns JSON list of all commands for scripting

## Example Scenario

You want to find commands for releasing a module:

```bash
# Search for release-related commands
eac show help | grep release

# Get help for release command
eac help release

# Get help for specific subcommand
eac help release changelog
```

## Common Issues

| Problem             | Solution                                          |
| ------------------- | ------------------------------------------------- |
| "Command not found" | Ensure EAC is installed and in PATH               |
| Too much output     | Use `eac help <command>` for specific command |

## Next Steps

- [Explore Your Repository](./explore-your-repository.md) → Discover modules and files
- [Command Reference](../../../../reference/eac/commands/index.md) → Complete command documentation

## Related Commands

- [`help`](../../../../reference/eac/commands/help/help.md) - Display help information
- [`show help`](../../../../reference/eac/commands/show/help.md) - Show help in table format
- [`get commands`](../../../../reference/eac/commands/get/commands.md) - Get commands as JSON
