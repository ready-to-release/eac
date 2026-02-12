# session

Cross-platform shell session identification for container naming and tracking.

## Key Types

_No exported types. This package provides a single utility function._

## Key Functions

- **`GetIdentifier`** -- Returns a unique identifier for the current shell session by checking platform-specific environment variables with PPID fallback

## Patterns

- Cascading detection: Checks PowerShell, TERM_SESSION_ID, SSH_CLIENT, WSL, tmux, screen, Windows Terminal, TTY, then falls back to parent PID
- Prefix-based identifiers: Returns strings like `pwsh-...`, `term-...`, `ssh-...`, `wsl-...`, `tmux-...`, `wt-...`, `pid-...` for easy classification

## Internal Structure

| File       | Responsibility                                            |
| ---------- | --------------------------------------------------------- |
| session.go | GetIdentifier with cascading platform-specific env checks |

## Dependencies

_None. This is a leaf package using only the standard library._

## Role in System

The session package provides a stable identifier for the current shell session, used for container naming and tracking across multiple clie invocations within the same terminal. This helps with cleanup operations and orphan container detection.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- No test files exist for this package. `session.go` lacks corresponding unit tests for the cross-platform shell session identification logic.

### Optimization Opportunities

- None identified.
