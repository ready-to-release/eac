# platform

Platform-specific abstractions for command execution and console output across
Unix and Windows.

## Key Types

- **`LineEnding`** -- Platform-specific line ending constant (\n or \r\n)
- **`WrapCommand`** -- Adapts command name and args for the current OS

## Patterns

- Build tags: Separate files per OS using `//go:build windows` and `//go:build !windows`
- Transparent wrapping: `WrapCommand` returns arguments unchanged, allowing future platform-specific command adaptation
- Compile-time selection: Only the matching platform file is included in the binary
- Paired convention: Each abstraction has one file per supported platform

## Internal Structure

| File | Responsibility |
| --- | --- |
| command_unix.go | Unix command wrapping (pass-through) |
| command_windows.go | Windows command wrapping (pass-through) |
| newline_unix.go | Unix line ending constant (\n) |
| newline_windows.go | Windows line ending constant (\r\n) |

## Dependencies

No internal dependencies. This package depends only on the Go standard library.

## Role in System

`platform` provides OS-level abstractions consumed by TUI rendering and command
execution throughout the `core` module. Its build-tag approach isolates all
platform-specific behavior into a single leaf package, ensuring that
cross-platform differences are handled in one place rather than scattered across
the codebase. The `LineEnding` constant is used by console output code to
produce correct line endings on each OS, while `WrapCommand` provides an
extension point for any future platform-specific command invocation needs.
Both are designed as stable, minimal APIs that other packages can depend on
without risk of churn.

## Code Health

### Tech Debt
- No test files exist; both `WrapCommand` and `LineEnding` are untested
- `WrapCommand` has identical pass-through implementations on both platforms (command_unix.go, command_windows.go); the abstraction exists but adds no current value

### Pain Points
- None identified

### Optimization Opportunities
- Consider adding a simple platform_test.go with build-tag-aware assertions for `LineEnding` to prevent accidental regression (trivial effort)
- If `WrapCommand` remains a pass-through long-term, it could be replaced by a single-file no-op to reduce maintenance surface (low priority, deferred until an actual platform divergence arises)
