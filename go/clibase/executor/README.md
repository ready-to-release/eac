# executor

Dispatches CLI commands by name through a command registry.

## Key Types

- **`Executor`** -- looks up commands from a `core.CommandRegistryPort` and executes them, translating registry lookups and type assertions into exit codes

## Key Functions

- `New` -- constructs an `Executor` backed by the given `CommandRegistryPort`

## Patterns

- Registry-based dispatch: the executor does not know about any specific commands; it resolves them at runtime via `registry.Get(name)` and delegates to `SimpleCommandPort.Execute`
- Parent-command guard: if the resolved command does not implement `SimpleCommandPort` (i.e., it is a parent/group command), execution is refused with an error message rather than panicking
- Thin wrapper: the package intentionally has minimal logic -- it wires `os.Stdout`/`os.Stderr` into a `CommandRequest` and returns the exit code, leaving all command logic to the implementations behind the port interfaces

## Internal Structure

| File | Responsibility |
| --- | --- |
| executor.go | `Executor` struct, `New` constructor, `Execute` dispatch method (lookup, type-assert, invoke), and `Registry` accessor |

## Dependencies

- `github.com/ready-to-release/eac/contracts/core/0.1.0` -- `CommandRegistryPort`, `SimpleCommandPort`, and `CommandRequest` interfaces/types that define the command contract

## Role in System

This package sits at the boundary between CLI argument parsing and command execution. After the CLI framework parses the user's input into a command name and arguments, the `Executor` resolves the command from the central registry and invokes it. It is a small but critical glue layer that decouples the CLI shell from individual command implementations.

## Code Health

### Tech Debt
- None

### Pain Points
- None identified

### Optimization Opportunities
- None identified
