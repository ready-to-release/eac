# terminal

Cross-platform terminal size detection and TTY identification.

## Key Types

_No exported types. This package provides utility functions with platform-specific implementations._

## Key Functions

- **`GetSize`** -- Returns terminal width and height using platform-specific detection
- **`GetWidth`** -- Returns terminal width in columns, defaulting to 80 on failure
- **`IsTerminal`** -- Returns true if stdout is a character device (TTY)

## Patterns

- Build-tag dispatch: `getTerminalSize()` implemented per platform via build tags (`_windows.go`, `_unix.go`, `_fallback.go`)
- Windows dual strategy: Primary detection via kernel32.dll `GetConsoleScreenBufferInfo`, fallback to parsing `mode con` output
- Graceful fallback: Returns 80x24 on unsupported platforms or detection failure

## Internal Structure

| File                        | Responsibility                                       |
| --------------------------- | ---------------------------------------------------- |
| terminal.go                 | Public API: GetSize, GetWidth, IsTerminal            |
| terminal_windows.go         | Windows implementation via kernel32.dll syscall       |
| terminal_windows_console.go | Windows fallback via `mode con` command parsing       |
| terminal_unix.go            | Unix implementation via IoctlGetWinsize               |
| terminal_fallback.go        | Returns 80x24 for unsupported platforms               |

## Dependencies

_None. This is a leaf package using only the standard library and platform syscalls._

## Role in System

The terminal package provides terminal dimension information used by the docker package's `BuildEnvironmentVars` to pass terminal width and height to containers as environment variables. This enables extensions to render output correctly sized to the user's terminal. The `IsTerminal` function is used to decide whether to enable TTY mode for container I/O.

## Code Health

### Tech Debt

_None identified._

### Pain Points

_None identified._

### Optimization Opportunities

_None identified._
