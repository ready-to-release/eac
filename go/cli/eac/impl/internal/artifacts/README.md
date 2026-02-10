# artifacts

Provides build artifact validation using UoW (Unit of Work) manifests. Checks artifact existence, integrity (via hash comparison), and staleness (source changes since build).

## Key Types

None (function-only package; uses types from `clibase/initsummary` for results).

## Key Functions

- **`ValidateBuildArtifacts()`** -- Validate that build artifacts exist and are up-to-date for given modules (checks all manifests on disk)
- **`ValidateBuildArtifactsWithExpected()`** -- Validate build artifacts with optional filtering to expected UoWs (ignores orphaned manifests)
- **`filterToExpected()`** -- Filter UoW manifests to only those matching expected UoW IDs

## Patterns

- UoW manifest-based validation: reads manifests from `out/build/{module}/{component}-{tool}/` directories
- Staleness detection via input hashing: compares current source file hashes against manifest-recorded input hashes
- Resilient hash comparison: a module is stale only if NO manifest matches (tolerates parallel build hash divergence)
- Expected UoW filtering: prevents false positives from orphaned manifests left by removed component types
- Artifact integrity checking: validates file existence and hash correctness per manifest

## Internal Structure

| File | Responsibility |
| --- | --- |
| validation.go | UoW manifest-based artifact validation with existence, integrity, and staleness checks |

## Dependencies

- `contracts/core/0.1.0` -- action constants for UoW manifest reading
- `clibase/initsummary` -- `ArtifactValidationInfo` result type
- `clibase/utils` -- utility functions (contains check)
- `core/config` -- EAC configuration for build output paths
- `core/domain/modules` -- module registry for glob pattern expansion
- `core/hash` -- file hashing and glob pattern expansion
- `core/logging` -- structured logging
- `core/output` -- UoW manifest reading and validation
- `core/workunit` -- UoW ID types for manifest identification

## Role in System

The `artifacts` package is used by the command framework's `AfterInit` hook to validate that required build artifacts exist and are current before running test, lint, or scan commands. It prevents stale artifact usage and provides clear diagnostics when builds need to be re-run.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
