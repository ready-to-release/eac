# environment

Runtime environment detection for CLI commands. Determines the execution context
(local, CI, Docker) and configures TUI availability based on terminal capabilities.

## Key Types

- `Env` -- detected environment state: CI provider, Docker presence, terminal type, and TUI eligibility

## Key Functions

- `Detect` -- inspects environment variables and OS state to produce an `Env` instance
- `ShouldUseTUI` -- returns whether TUI mode should be enabled based on detected environment
- `ValidateTUI` -- checks terminal capabilities against TUI requirements and returns a reason string if TUI cannot be used

## Patterns

- **Environment sniffing**: `Detect` reads well-known environment variables (CI, GITHUB_ACTIONS, TERM, etc.) to determine the execution context
- **Defensive TUI activation**: TUI is only enabled when terminal capabilities are confirmed, preventing broken rendering in CI or piped output

## Internal Structure

| File | Purpose |
|---|---|
| `environment.go` | `Env` struct, `Detect()`, `ShouldUseTUI()`, and `ValidateTUI()` |

## Dependencies

None (leaf package within clibase).

## Role in System

Called early in command initialization to determine how the CLI should behave. CI environments get plain-text output, local terminals with sufficient capabilities get TUI mode, and Docker contexts adjust path handling. The detected environment flows into `display.Config` and flag defaults.

## Code Health

### Tech Debt
- None identified; clean 85-line file with no mutable global state

### Pain Points
- None identified

### Optimization Opportunities
- None identified; compact leaf package
