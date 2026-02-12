# describe

Outputs structured command information in JSON or YAML format for building shell integrations, completion scripts, and tooling.

## Key Types

- **`CommandInfo`** -- Structured information about a single command: name, parts, description, parent, leaf status, and argument type
- **`CommandTree`** -- Hierarchical command structure containing all commands, a parent-to-children tree mapping, and available module monikers for completion

## Key Functions

- `GetCommands` -- Entry point for the `get commands` command; parses format flag and outputs the command tree as JSON or YAML

## Patterns

- **Command registration via init()**: Uses `registry.Register(GetCommands)` for automatic command discovery
- **Registry introspection**: Builds the command tree by iterating over the command registry, extracting parent-child relationships from space-separated command names
- **Machine-readable output**: Outputs structured data (JSON/YAML) for consumption by shell integrations rather than human-readable text

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | All package functionality: CommandInfo/CommandTree types, GetCommands entry point, buildCommandTree, format flag parsing, and loadModuleMonikers |

## Dependencies

- `go/clibase/flags` -- Shared flag validation
- `go/clibase/registry` -- Command registry access for introspecting all registered commands
- `go/core/domain/reports` -- Module contract reports for loading module monikers
- `go/core/logging` -- Component logger for error output
- `go/core/repository` -- Repository root discovery for module loading

## Role in System

This package provides machine-readable command metadata for the eac CLI. It enables shell completion scripts and external tooling (like the clie CLI) to discover available commands, their hierarchical structure, argument types, and module names without parsing help text. The output feeds into completion engines and integration frameworks.

## Code Health

### Tech Debt
- No unit tests for commands.go (186 lines); relies on BDD-level testing

### Pain Points
- None identified

### Optimization Opportunities
- None identified
