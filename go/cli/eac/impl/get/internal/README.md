# internal (get)

Provides shared output rendering and format parsing utilities for all `get` commands.

## Key Types

- **`OutputFormat`** -- Represents the desired output format with boolean flags for YAML, JSON, and TOML, plus an optional custom renderer name

## Key Functions

- `ExecuteGetCommand` -- Helper that wraps the common pattern for get commands: parse output format flags, execute data fetcher, render and output the result
- `ParseOutputFlags` -- Parses `--as-yaml`, `--as-json`, `--as-toml`, and `--as-<custom>` flags from command arguments, validating custom renderers against available renderers for the specific command
- `RenderAndOutput` -- Marshals data to YAML first (single source of truth), then renders in the requested format (JSON, TOML, custom, or default YAML)

## Patterns

- **YAML as single source of truth**: All data is first marshaled to YAML regardless of output format, then converted to the requested format; this ensures consistent serialization behavior
- **Command-specific custom renderers**: Custom `--as-<name>` flags are validated against renderers registered for the specific command name, preventing invalid flag usage
- **Reusable command scaffold**: `ExecuteGetCommand` provides a consistent pattern for all get commands: parse flags, fetch data, render output, with uniform error handling

## Internal Structure

| File | Responsibility |
| --- | --- |
| renderer.go | All package functionality: OutputFormat type, ParseOutputFlags, RenderAndOutput, ExecuteGetCommand helper, and getCallerCommandName utility |

## Dependencies

- `go/clibase/render` -- Rendering functions (RenderAsJSON, RenderAsTOML, RenderAsCustom, ListCustomRenderers)

## Role in System

This package is the shared foundation for all `get` commands in the eac CLI. Every `get` command (get-files, get-modules, get-dependencies, etc.) uses `ExecuteGetCommand` or `RenderAndOutput` to handle output format selection and rendering consistently. It ensures all get commands support the same `--as-yaml`, `--as-json`, `--as-toml`, and custom renderer flags with uniform behavior.

## Code Health

### Tech Debt
- renderer.go (153 lines) lacks dedicated unit tests; tested indirectly via get command tests

### Pain Points
- None identified

### Optimization Opportunities
- None identified
