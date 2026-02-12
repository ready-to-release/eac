# registry

Command registration and dispatch for the CLI. Provides a central registry where
commands register themselves with metadata, flags, and handler functions.

## Key Types

- `CommandRegistration` -- holds a command's name, description, handler function, flag metadata, and subcommand groups
- `FlagMetadata` -- describes a single flag for validation and documentation: name, type, default, description
- `CommandFunc` -- function signature for command handlers
- `SubcommandGroup` -- groups related subcommands under a parent command
- `DeclarativeFlagDefLike` -- interface for extracting flag metadata from declarative flag definitions

## Key Functions

- `Register` -- registers a command with its handler and metadata
- `Get` -- retrieves a command registration by name
- `All` -- returns all registered commands (snapshot copy)
- `BuildDeclarativeMetadata` -- converts declarative flag definitions into `FlagMetadata` for registry storage

## Patterns

- **Self-registration**: commands call `Register()` from their `init()` functions, populating the registry at startup
- **Metadata-driven validation**: flag metadata enables runtime validation without coupling to specific flag implementations
- **Subcommand grouping**: `SubcommandGroup` supports hierarchical command structures (e.g., `work create`, `work merge`)
- **Thread-safe access**: `sync.RWMutex` protects all reads and writes; `GetCommandRegistry()` returns a snapshot copy

## Internal Structure

| File                  | Purpose                                                                 |
| --------------------- | ----------------------------------------------------------------------- |
| `command_registry.go` | Package-level doc comment                                               |
| `types.go`            | `ErrDuplicateCommand`, global registry, `CommandRegistry` struct        |
| `registration.go`     | `Register()`, `MustRegister()`, `RegisterAll()`                         |
| `dispatch.go`         | `Get()`, `GetByCanonical()`, `All()`, `Names()`, `Subcommands()`       |

## Dependencies

- `core/workspace` -- workspace root detection for command execution context

## Role in System

Acts as the command dispatch layer between the CLI entry point and individual command implementations. The main CLI binary queries the registry to find and execute commands, validate flags, and generate help documentation.

## Code Health

### Tech Debt

- None identified

### Pain Points

- `command_registry_test.go` (349 lines) exceeds 300 lines and is significantly larger than all implementation files combined
- Implementation files are small: `dispatch.go` (66 lines), `types.go` (47 lines), `registration.go` (34 lines), `command_registry.go` (3 lines)

### Optimization Opportunities

- None identified
