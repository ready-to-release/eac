# eac

EAC command execution adapter that routes commands through native binary
or container modes via the tool abstraction layer.

## Key Types

- **`EACPort`** -- Interface for executing EAC commands
- **`ExecConfig`** -- Configures command execution options
- **`Result`** -- Holds exit code, stdout, stderr, and duration
- **`CommandExecutorAdapter`** -- Bridges `EACPort` to legacy executor pattern
- **`MockEAC`** -- In-memory mock that records calls for testing

## Patterns

- Port interface: `EACPort` with single `Execute` method for mockability
- Factory routing: `New()` resolves native vs container via tool registry
- Adapter pattern: `CommandExecutorAdapter` bridges to `func(ctx, root, args)` signature
- Dependency injection: Callers receive `EACPort` once, pass it to consumers

## Internal Structure

| File/Sub-package    | Responsibility                                         |
| ------------------- | ------------------------------------------------------ |
| port.go             | `EACPort` interface, `ExecConfig`, and `Result` types  |
| factory.go          | `New()` factory routing to native or container adapter |
| native.go           | `nativeAdapter` executing via local binary             |
| container.go        | `containerAdapter` executing via eac-ext container     |
| command_executor.go | `CommandExecutorAdapter` and `RunCommand` convenience  |
| mock.go             | `MockEAC` for unit testing                             |

## Dependencies

- `core/tool` -- tool registry, executor, and `ToolDefinition`
- `core/paths` -- resolving commands binary path

## Role in System

The `eac-to-eac` module allows the system to invoke EAC CLI commands as
a first-class tool. It supports native execution (local binary) and
container execution (eac-ext Docker image), selected automatically based
on the tool registry configuration. This enables recursive EAC invocation
within build pipelines and test infrastructure.

## Code Health

### Tech Debt

- `New()` in factory.go calls `tool.GlobalRegistry()` and `tool.GlobalExecutor()` directly; accepting these as parameters would improve testability and remove hidden coupling

### Pain Points

- `nativeAdapter` and `containerAdapter` have nearly identical `Execute` methods (native.go:32, container.go:25); shared logic could be extracted to reduce duplication

### Optimization Opportunities

- None identified
