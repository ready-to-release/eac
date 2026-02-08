# environment

Runtime environment detection for CLI commands. Determines the execution context
(local, CI, Docker) and configures TUI availability based on terminal capabilities.

## Key Types

- `Env` -- detected runtime environment with CI provider, terminal capabilities, and context metadata

## Key Functions

- `Detect` -- probes the environment and returns a populated `Env` with CI detection, terminal type, and color support
- `ShouldUseTUI` -- determines whether TUI mode should be enabled based on terminal capabilities and environment
- `ValidateTUI` -- validates that the current environment supports TUI rendering (terminal size, type)
- `ContextName` -- returns a human-readable name for the current execution context

## Patterns

- **Probe-based detection**: `Detect` checks environment variables, terminal properties, and OS signals to build a comprehensive environment snapshot
- **TUI eligibility**: separates "should use TUI" (preference) from "can use TUI" (capability), allowing graceful fallback to console mode

## Internal Structure

| File | Purpose |
|---|---|
| `environment.go` | `Env` struct, `Detect()`, TUI eligibility and validation functions |

## Dependencies

- `core/environments` -- environment type constants and CI provider detection
- `core/logging` -- structured logging for detection diagnostics

## Role in System

Called early in command initialization to determine how the CLI should behave. The detected environment influences TUI mode selection, color output, concurrency defaults, and CI-specific behaviors like artifact paths.

## Code Health

### Tech Debt
- None identified; no TODO/FIXME markers, no mutable global state

### Pain Points
- None identified; single-file package (84 lines) with good test coverage (140 lines)

### Optimization Opportunities
- None identified; the package is minimal and focused
