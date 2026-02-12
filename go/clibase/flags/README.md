# flags

Composable flag system for CLI commands. Provides reusable flag sets that commands
assemble declaratively, with parsing, validation, and documentation generation.

## Key Types

- `FlagSet` -- interface for composable flag groups that register flags and apply parsed values
- `FlagDef` -- defines a single CLI flag: name, shorthand, type, default, description, and metadata
- `Parser` -- parses command-line arguments against registered `FlagDef` definitions
- `ParsedFlags` -- holds parsed flag values with typed accessor methods
- `CommandFlagConfig` -- maps command names to their flag configurations for dispatch
- `DeclarativeFlagDef` -- extends `FlagDef` with enable/disable pairs (e.g., `--deps` / `--skip-deps`)
- `DeclarativeFlagSet` -- interface for flag sets that use declarative enable/disable semantics
- `DeclarativeState` -- tracks whether a declarative flag was explicitly enabled or disabled
- `ExecutionFlags` -- flag set for concurrency, sequential, and turbo execution controls
- `OutputFlags` -- flag set for output formatting (TUI, ASCII mode, debug, timings)
- `CacheFlags` -- flag set for fine-grained cache control
- `ModuleFlags` -- flag set for module selection and skip filters
- `DryRunFlags` -- flag set for dry-run mode
- `SharedFlags` -- aggregates common flag sets used across multiple commands
- `CommandDoc` -- structured documentation for a command's flags
- `FlagDocGenerator` -- generates flag documentation from registered metadata

## Patterns

- **Composable flag sets**: commands assemble flag groups by composing `FlagSet` implementations (e.g., `ExecutionFlags` + `ModuleFlags` + `OutputFlags`)
- **Declarative enable/disable pairs**: `DeclarativeFlagDef` supports paired flags like `--deps`/`--skip-deps` with conflict detection
- **Registry-based validation**: `ValidateFlagsFromRegistry()` validates flags against the command registry metadata
- **Typed accessors**: `ParsedFlags` provides `GetBool`, `GetString`, `GetInt` with defaults

## Internal Structure

| File | Purpose |
|---|---|
| `sets.go` | `FlagSet` interface and `FlagDef` struct definitions |
| `parser.go` | `Parser` and `ParsedFlags` for argument parsing |
| `declarative.go` | `DeclarativeFlagDef`, `DeclarativeState`, and `DeclarativeFlagSet` interface |
| `flags.go` | Utility functions: `ParseDebugFlag`, `HasFlag`, `GetFlagValue`, `GetPositionalArgs` |
| `execution.go` | `ExecutionFlags` and `ExecutionFlagSet` for concurrency control |
| `output.go` | `OutputFlags` and `OutputFlagSet` for display configuration |
| `cache.go` | `CacheFlags` and `CacheFlagSet` for cache control |
| `module.go` | `ModuleFlags` and `ModuleFlagSet` for module selection |
| `dryrun.go` | `DryRunFlags` and `DryRunFlagSet` |
| `commands.go` | `SharedFlags`, `ParseSharedFlags()`, predefined flag configs per command |
| `registry.go` | `ValidateFlagsFromRegistry()` for registry-based flag validation |
| `docs.go` | `CommandDoc` and `FlagDocGenerator` for flag documentation |

## Dependencies

- `clibase/environment` -- environment detection for TUI defaults
- `clibase/display` -- display config types
- `clibase/registry` -- command registry for flag validation
- `core/cache` -- cache config types

## Role in System

Provides the flag parsing and composition layer used by all CLI commands. Each command selects which flag sets it needs, and the framework handles parsing, validation, and documentation. The declarative flag system ensures consistent UX for enable/disable flag pairs across commands.

## Code Health

### Tech Debt

- None identified

### Pain Points

- `output_test.go` is 494 lines, significantly exceeds 300-line threshold
- `cache_test.go` is 466 lines, significantly exceeds 300-line threshold
- `parser_test.go` is 339 lines, exceeds 300-line threshold
- `commands_test.go` is 333 lines, exceeds 300-line threshold

### Optimization Opportunities

- None identified
