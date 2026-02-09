# fileutil

File operation utilities with atomic writes, locked writes, and platform-aware cleanup.
Provides safe file operations that handle concurrent access and platform-specific quirks.

## Key Functions

- `AtomicWrite` -- writes data to a file atomically using a temporary file and rename
- `AtomicWriteJSON` -- marshals a value to JSON and writes it atomically
- `AtomicWriteJSONWithLock` -- atomic JSON write with file-based locking for cross-process safety
- `RemoveAllWithRetry` -- removes a directory tree with platform-specific retry logic

## Patterns

- **Atomic write via rename**: writes to a temporary file in the same directory, then renames to the target path, preventing partial writes from being observed
- **Platform-specific retry**: on Windows, `RemoveAllWithRetry` retries on "access denied" errors caused by antivirus or indexing locks; on other platforms it passes through to `os.RemoveAll`
- **Build-tag separation**: Windows and non-Windows retry implementations are separated via Go build tags

## Internal Structure

| File | Purpose |
|---|---|
| `atomic.go` | `AtomicWrite` and `AtomicWriteJSON` using temp file + rename |
| `locked_write.go` | `AtomicWriteJSONWithLock` combining file locks with atomic writes |
| `retry_windows.go` | Windows-specific `RemoveAllWithRetry` with retry loop |
| `retry_other.go` | Non-Windows `RemoveAllWithRetry` passthrough |

## Dependencies

- `core/config` -- config types for JSON serialization targets

## Role in System

Provides safe file operations used throughout the CLI for writing build outputs, test results, cache manifests, and CTRF reports. The atomic write pattern prevents corrupted files when multiple processes or goroutines write to the same output directory.

## Code Health

### Tech Debt
- None identified; no TODO/FIXME markers, no mutable global state, all functions are short and focused

### Pain Points
- None identified; compact package with strong test coverage (393 test lines vs 140 source lines)

### Optimization Opportunities
- None identified; the package is well-structured with clear build-tag separation
