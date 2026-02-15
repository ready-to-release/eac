# gource

Provides the `serve gource` command for launching Gource git repository visualization via Docker with optional video export support.

## Key Types

None (command-only package).

## Key Functions

- **`ServeGource()`** -- Entry point for the `serve gource` command; launches Gource visualization via Docker
- **`handleStop()`** -- Handle Ctrl+C signal to gracefully stop the Docker container
- **`handleFileOutput()`** -- Handle video file export from the Gource container
- **`printUsage()`** -- Display command usage and flag documentation

## Patterns

- `init()` registration: registers `ServeGource` command function with the global registry
- Docker-based tool execution: runs Gource container with repository volume mount
- Signal handling: captures Ctrl+C to cleanly stop Docker containers
- Optional file output: supports exporting visualization as video file

## Internal Structure

| File | Responsibility |
| --- | --- |
| serve.go | Gource git visualization via Docker with signal handling and video export (455 lines) |

## Dependencies

- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `gource` sub-package provides repository history visualization within the `serve` command group. It enables developers to view an animated visualization of the repository's commit history, useful for understanding codebase evolution and contributor patterns.

## Code Health

### Tech Debt
- No test for serve.go (540 lines)

### Pain Points
- serve.go (540 lines) exceeds 300-line guideline

### Optimization Opportunities
- Docker container lifecycle helpers are already well-decomposed within the file; extracting to a shared package has low benefit since gource is the only consumer (deferred)
