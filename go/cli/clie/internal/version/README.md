# version

Semantic version parsing, comparison, validation, and build-flag metadata.

## Key Types

- **`Info`** -- Version metadata: Version, Timestamp, Commit, BuildTime, Modified

## Key Functions

- **`GetInfo`** -- Returns current version information (thread-safe read)
- **`SetVersion`** -- Sets version fields from build flags and debug.BuildInfo (thread-safe write)
- **`Validate`** -- Validates version format (semver) and range constraints
- **`CompareVersions`** -- Compares two semver strings returning -1, 0, or 1
- **`SplitVersion`** -- Splits version into main and prerelease components
- **`IsValid`** -- Checks if a version string matches the semver pattern

## Patterns

- Thread-safe globals: `sync.RWMutex` protects all version variables for concurrent access
- Build-flag injection: Version, Commit, Timestamp set via ldflags at compile time
- Range validation: Version checked against configurable min/max range
- Prerelease ordering: Release versions rank higher than prerelease versions in comparison

## Internal Structure

| File       | Responsibility                                                    |
| ---------- | ----------------------------------------------------------------- |
| version.go | Info type, version vars, GetInfo, SetVersion, Validate, CompareVersions |

## Dependencies

- `internal/logging` -- Debug logging during validation
- `internal/envconsts` -- CLIE_NO_UPDATE_CHECK environment variable name

## Role in System

The version package provides version metadata for the `clie version` command and version validation during startup. The `Validate` function runs in root command's `init()` to catch invalid or out-of-range versions early. The `CompareVersions` function is used by the update self command to determine if an update is available.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- None identified.

### Optimization Opportunities

- None identified.
