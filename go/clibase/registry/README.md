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
- `All` -- returns all registered commands
- `BuildDeclarativeMetadata` -- converts declarative flag definitions into `FlagMetadata` for registry storage

## Patterns

- **Self-registration**: commands call `Register()` from their `init()` functions, populating the registry at startup
- **Metadata-driven validation**: flag metadata enables runtime validation without coupling to specific flag implementations
- **Subcommand grouping**: `SubcommandGroup` supports hierarchical command structures (e.g., `work create`, `work merge`)

## Internal Structure

| File | Purpose |
|---|---|
| `registry.go` | `CommandRegistration`, `FlagMetadata`, `Register()`, `Get()`, dispatch functions |
| `declarative.go` | `DeclarativeFlagDefLike` interface and `BuildDeclarativeMetadata()` converter |

## Dependencies

- `core/workspace` -- workspace root detection for command execution context

## Role in System

Acts as the command dispatch layer between the CLI entry point and individual command implementations. The main CLI binary queries the registry to find and execute commands, validate flags, and generate help documentation.

## Code Health

### Tech Debt
- ~~`registry.go:79` mutable package-level `commandRegistry` map; test-time mutations could race~~ (resolved: `sync.RWMutex` protects all reads/writes; `GetCommandRegistry()` returns snapshot copy)
- ~~`registry.go:15` exported mutable `InitialWorkingDir` package-level var~~ (resolved: removed dead variable; eac main.go has its own)

### Pain Points
- `registry.go` (530 lines) combines type definitions, registration logic, and dispatch in a single file; splitting would improve readability

### Optimization Opportunities
- ~~Protect `commandRegistry` with `sync.RWMutex` for safe concurrent test access~~ (resolved)
- Split `registry.go` into `types.go`, `register.go`, and `dispatch.go` for clarity (low effort)
