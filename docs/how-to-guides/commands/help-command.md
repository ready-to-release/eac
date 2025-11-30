# Help Command

## Purpose

The `help` command provides built-in documentation for all EAC commands. It displays usage information, available flags, subcommands, and detailed descriptions to help you understand and use any command effectively.

## Quick Reference

```bash
# Show all available commands
eac help

# Show help for a specific command
eac help <command>

# Show help for a subcommand
eac help <parent> <subcommand>

# Show detailed help with additional information
eac help --verbose
eac help -v <command>
```

## Command Syntax

```text
help [command] [subcommand]
```

### Flags

| Flag        | Shorthand | Type | Description                                                  |
| ----------- | --------- | ---- | ------------------------------------------------------------ |
| `--verbose` | `-v`      | bool | Show detailed information including additional help sections |

## Usage Modes

### 1. List All Commands

Display all available commands grouped by category.

```bash
eac help
```

**Output includes:**

- Command categories (e.g., Build, Test, Validation, etc.)
- Brief description of each command
- Command aliases if available

**Example:**

```bash
eac help
```

```text
EAC - Enterprise Application Contracts

USAGE:
  eac [command] [flags]

AVAILABLE COMMANDS:
  Build Commands:
    build                 Build one or more modules by moniker

  Test Commands:
    test                  Test one or more modules by moniker
    test-suite           Run tests for a specific test suite

  Validation Commands:
    validate             Validate repository contracts and dependencies
    validate-contracts   Validate repository contracts against JSON schemas

  ...
```

### 2. Show Command Help

Display detailed help for a specific command.

```bash
eac help <command>
```

**Example:**

```bash
eac help build
```

```text
NAME:
  build - Build one or more modules by moniker

SYNOPSIS:
  eac build [monikers...] [flags]

DESCRIPTION:
  Build one or more modules by moniker. Supports parallel builds and
  dependency-aware execution.

FLAGS:
  --parallel, -p    Build modules in parallel (default: true)
  --verbose, -v     Show detailed build output

EXAMPLES:
  eac build api
  eac build api worker --parallel=false
```

### 3. Show Subcommand Help

Display help for nested subcommands.

```bash
eac help <parent> <subcommand>
```

**Example:**

```bash
eac help work create
```

```text
NAME:
  work create - Create a new workspace for parallel development

SYNOPSIS:
  eac work create <workspace-name> [flags]

DESCRIPTION:
  Create a new git worktree-based workspace for parallel development.
  This allows you to work on multiple features simultaneously without
  switching branches in your main workspace.

FLAGS:
  --branch, -b      Create from specific branch (default: main)
  --path, -p        Custom path for workspace

EXAMPLES:
  eac work create feature-auth
  eac work create bugfix-123 --branch develop
```

### 4. Verbose Mode

Show additional detailed information about a command.

```bash
eac help --verbose
eac help -v <command>
```

**Example:**

```bash
eac help -v validate
```

Verbose mode adds:

- Extended descriptions
- Additional examples
- Configuration details
- Related commands
- Troubleshooting tips

## Understanding Help Output

### Output Sections

#### NAME

The command name and a brief one-line description.

```text
NAME:
  build - Build one or more modules by moniker
```

#### SYNOPSIS

The command syntax showing required and optional parameters.

```text
SYNOPSIS:
  eac build [monikers...] [flags]
```

**Syntax notation:**

- `<required>` - Required parameter
- `[optional]` - Optional parameter
- `[item...]` - Zero or more items
- `--flag` - Boolean flag
- `--flag=value` - Flag with value

#### DESCRIPTION

A detailed explanation of what the command does and how it works.

```text
DESCRIPTION:
  Build one or more modules by moniker. Supports parallel builds and
  dependency-aware execution. When no monikers are specified, builds
  all changed modules.
```

#### COMMANDS

For parent commands, lists available subcommands.

```text
COMMANDS:
  create    Create a new workspace
  list      List all workspaces
  remove    Remove a workspace
  merge     Merge workspace changes
```

#### FLAGS

Available command-line flags with their descriptions and default values.

```text
FLAGS:
  --parallel, -p    Build modules in parallel (default: true)
  --verbose, -v     Show detailed build output
  --config string   Path to config file
```

**Flag notation:**

- Boolean flags: `--flag` or `--flag=true/false`
- Value flags: `--flag=value` or `--flag value`
- Short flags: `-p` (can be combined: `-pv`)

#### EXAMPLES

Practical examples showing common usage patterns.

```text
EXAMPLES:
  # Build a single module
  eac build api

  # Build multiple modules
  eac build api worker

  # Build with custom flags
  eac build api --parallel=false --verbose
```

#### ADDITIONAL INFORMATION

(Verbose mode only) Extra details about the command.

```text
ADDITIONAL INFORMATION:
  Configuration:
    Build settings are defined in module contracts at:
    .eac/modules/{module}/module.contract.yaml

  Related Commands:
    - test: Run tests after building
    - validate: Validate before building
    - get-modules: List all available modules

  Troubleshooting:
    If builds fail, check:
    - Module dependencies are satisfied
    - Build tools are installed
    - Paths in module contracts are correct
```

## Common Use Cases

### Discover Available Commands

When starting with EAC, see what's available:

```bash
eac help
```

Browse through categories to find relevant commands.

### Learn Command Usage

Before using a new command, check its help:

```bash
eac help validate-dependencies
```

Review the synopsis, flags, and examples.

### Explore Subcommands

For commands with subcommands, explore the hierarchy:

```bash
# See all work subcommands
eac help work

# Learn about a specific subcommand
eac help work create
```

### Get Detailed Information

Use verbose mode for comprehensive documentation:

```bash
eac help -v pipeline-run
```

Review additional information, configuration details, and troubleshooting tips.

### Quick Flag Reference

Check available flags without reading full documentation:

```bash
eac help test-suite | grep FLAGS -A 10
```

## Tips and Best Practices

### 1. Start with Help

Before using any command, always run `help` first:

```bash
eac help <command>
```

This prevents errors and helps you understand the command's purpose.

### 2. Use Tab Completion

If your shell supports completion, use it to discover commands:

```bash
eac help <tab><tab>
```

Generate completion scripts with:

```bash
eac completion bash > ~/.eac-completion.sh
```

### 3. Check Examples

The examples section often shows the most common usage patterns. Start there:

```bash
eac help build | grep EXAMPLES -A 10
```

### 4. Combine with Grep

Filter help output to find specific information:

```bash
# Find all parallel-related flags
eac help test-suite | grep -i parallel

# Find examples
eac help validate | grep -A 20 EXAMPLES
```

### 5. Use Verbose for Complex Commands

For commands you use infrequently or that are complex:

```bash
eac help -v pipeline-run
```

The additional context helps ensure correct usage.

### 6. Check Parent Command Help

If a subcommand doesn't exist, check the parent:

```bash
# If this doesn't work
eac help work unknown

# Try this
eac help work
```

## Related Commands

- `completion` - Generate shell completion scripts
- `show-config` - Display current configuration
- `get` - Retrieve repository data

## Troubleshooting

### Help Command Not Working

**Problem:** `eac help` returns an error

**Solution:**

1. Verify EAC is installed: `eac --version`
2. Check you're in an EAC repository: `ls .eac/`
3. Ensure binary has execute permissions

### Command Not Listed

**Problem:** A command doesn't appear in help output

**Solution:**

1. Update to latest version: `git pull`
2. Rebuild EAC: `eac build src-cli`
3. Check command exists: `eac <command> --help`

### Verbose Flag No Effect

**Problem:** `--verbose` doesn't show additional information

**Solution:**

- Verbose mode only adds content for commands that have additional information
- Not all commands have verbose help sections
- Try: `eac help -v <command>` instead of `eac <command> -v help`

### Help Output Truncated

**Problem:** Help output is cut off

**Solution:**

- Pipe through a pager: `eac help | less`
- Increase terminal buffer size
- Redirect to file: `eac help > help.txt`

## Configuration

The help command requires no configuration. It reads command metadata from the EAC binary itself.

### Customizing Help Output

Help output is generated from:

1. Command definitions in source code
2. Flag definitions with descriptions
3. Usage examples in command implementations

To modify help output, update the command source code and rebuild EAC.

## Notes

- Help is context-aware and shows only relevant commands for the current repository
- Some commands may have different help output based on repository configuration
- The `--help` and `-h` flags can be used with any command as an alternative to `help <command>`
- Help output follows standard CLI conventions for readability

## See Also

- [How-to Guides](../index.md) - Complete list of all command guides
- [Tutorials](../../tutorials/index.md) - Introduction to EAC
