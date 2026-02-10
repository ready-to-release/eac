# testdata

Provides shared test data preparation for `get-tests`, `show-tests`, `get-suite`, and `show-suite` commands. Handles test discovery, enrichment, OS filtering, and file-to-module mapping.

## Key Types

- **`TestData`** -- Pre-computed test data with discovery results, OS filtering, and aggregations by type/level/module

## Key Functions

- **`GetAllTests()`** -- Discover all tests and return enriched test data with aggregations
- **`FindRepoRoot()`** -- Find repository root by walking up for `.git` directory, respecting `CLIE_REPO_ROOT` env var
- **`BuildFileModuleMap()`** -- Create a mapping from file paths to module monikers using git and module contracts
- **`mapGOOSToDepTag()`** -- Convert `runtime.GOOS` to dependency tag format (e.g., `darwin` to `macos`)

## Patterns

- Unified discovery: uses `testing.DiscoverAndEnrich()` for comprehensive test discovery with inference
- OS filtering: filters tests by current platform using dependency tag matching
- File-module mapping: maps test files to their owning modules via git repository file listing
- Environment variable override: respects `CLIE_REPO_ROOT` for isolated test environments

## Internal Structure

| File | Responsibility |
| --- | --- |
| helpers.go | Repository root finding with environment variable override |
| testdata.go | Test discovery, enrichment, OS filtering, file-module mapping, and aggregation |

## Dependencies

- `core/config` -- EAC configuration and environment configuration
- `core/domain/modules` -- module registry for test enrichment
- `core/domain/reports` -- module contract loading
- `core/environments` -- environment variable constants
- `core/git` -- git repository management
- `core/logging` -- structured logging
- `core/repository` -- repository file listing with module ownership
- `core/testing` -- unified test discovery and enrichment

## Role in System

The `testdata` package provides the shared data preparation layer for all test listing commands. Both GET and SHOW variants of test and suite commands delegate to `GetAllTests()` to discover, enrich, and aggregate test data before applying their own filtering and formatting.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
