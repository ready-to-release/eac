# extension

Outputs YAML-formatted extension metadata describing the eac extension's capabilities, commands, requirements, and configuration for integration with the clie CLI.

## Key Types

- **`Metadata`** -- Top-level extension metadata structure: name, version, description, schema version, capabilities, commands, requirements, volumes, expected host images, environment variables, and author metadata
- **`Command`** -- Describes a single command with description and parameter names
- **`Requirements`** -- Extension requirements: clie version, container runtime, minimum memory and CPU
- **`ExtensionMetadata`** -- Author, repository, documentation URL, license, and tags
- **`Volume`** -- Volume mount request with name, target path, and type (cache or bind)
- **`EnvVar`** -- Environment variable definition with name, optional value, and required flag

## Key Functions

- `ExtensionMeta` -- Entry point for the `extension-meta` command; builds metadata from the command registry and outputs as YAML
- `PrintAvailableCommands` -- Prints a user-friendly list of available commands to stdout

## Patterns

- **Command registration via init()**: Uses `registry.Register(ExtensionMeta)` for automatic command discovery
- **Registry-driven metadata**: Commands map is built dynamically from the command registry, extracting flag names as parameters, so the metadata stays in sync with actual registered commands
- **Self-describing extension**: The metadata output follows a schema contract that the clie CLI uses for extension discovery and configuration

## Internal Structure

| File | Responsibility |
| --- | --- |
| meta.go | All package functionality: Metadata and related types, ExtensionMeta command implementation, PrintAvailableCommands helper |

## Dependencies

- `go/clibase/flags` -- Shared flag validation
- `go/clibase/registry` -- Command registry access for building commands map
- `go/core/logging` -- Component logger for error output
- `go/core/tool` -- Tool image resolution for expected host images

## Role in System

This package implements the `extension-meta` command, which is the discovery mechanism for the clie CLI to understand what the eac extension provides. When clie loads eac as an extension, it calls `extension-meta` to get the YAML metadata describing all available commands, required volumes, environment variables, container runtime needs, and expected Docker images. This enables clie to properly configure and proxy commands to the eac binary.

## Code Health

### Tech Debt
- No unit tests for meta.go (217 lines); relies on integration-level testing

### Pain Points
- None identified

### Optimization Opportunities
- None identified
