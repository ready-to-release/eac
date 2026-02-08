# execution

Defines the cache verification interface for dependency-based work unit
execution.

## Key Types

- **`CacheVerifier`** -- Port interface for checking work unit cache status
- **`CacheResult`** -- Cache hit/miss with timestamp
- **`CacheVerifierFunc`** -- Adapter to use functions as `CacheVerifier`

## Patterns

- Port interface: commands implement `CacheVerifier` with their own cache logic
- Functional adapter: `CacheVerifierFunc` wraps plain functions as verifiers
- Context-aware: verification respects cancellation via `context.Context`

## Internal Structure

| File | Responsibility |
| --- | --- |
| doc.go | Package documentation with usage examples |
| policy.go | `CacheVerifier` interface, `CacheResult`, `CacheVerifierFunc` |

## Dependencies

- `core/workunit` -- `UnitSpec` type passed to verifier

## Role in System

The `execution` package defines the cache-verification contract between
commands and the orchestrator in `core`. Build, test, lint, and scan commands
each implement `CacheVerifier` so the orchestrator can skip cached units and
progressively update the TUI as cache hits are detected.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- Package is minimal (43-line interface + 40-line test); well-suited as a reference for port-interface design
- No changes recommended; clean leaf package
