# workspace

Detects the workspace root directory using a prioritized detection chain:
environment override, container mode, or git directory walk-up.

## Key Types

- **`Workspace`** -- Resolved root path, detection source, and container flag
- **`Options`** -- Detection mode, start path, and validation settings
- **`Mode`** -- Detection strategy: `ModeAuto`, `ModeExplicit`, `ModeGitOnly`
- **`DetectionError`** -- Structured error with operation, source, and path

## Patterns

- Cached detection: git-detected roots are cached for process lifetime
- Test isolation: `ForTesting` sets `CLIE_REPO_ROOT` with automatic cleanup
- Mode selection: `ModeExplicit` for strict isolation, `ModeGitOnly` for real root

## Internal Structure

| File | Responsibility |
| --- | --- |
| doc.go | Package documentation with detection precedence rules |
| workspace.go | `Detect`, `ForTesting`, `RequireIsolation`, git root walk |
| config.go | `Options`, `Mode` constants, `DefaultOptions` |
| errors.go | `DetectionError`, sentinel errors `ErrNotFound`/`ErrInvalidPath` |

## Dependencies

- `core/environments` -- environment variable constants
- `core/paths` -- container root and config path constants

## Role in System

The `workspace` package is the entry point for all path resolution in `core`.
Nearly every command begins by calling `workspace.Detect()` to locate the
repository root. Test infrastructure uses `ForTesting` and `RequireIsolation`
to ensure tests operate in isolated temporary directories.

## Code Health

### Tech Debt
- None identified

### Pain Points
- `workspace.go` uses process-level cache and `os.Getenv` directly, making concurrent detection tests require careful isolation
- `MustDetect()` and `RootOrPanic()` use panics; callers must be aware these are unsafe in library contexts

### Optimization Opportunities
- Replace direct `os.Getenv` calls with an injectable environment reader to simplify testing without real env vars (low effort)
- Package is well-tested (729-line test file) and well-structured; no major issues
