# repository

BDD test suite that validates repository-wide structural invariants using Godog. Covers Go module tidiness, module hierarchy consistency, file ownership, dependency isolation, version consistency, script placement, and release folder conventions.

## Key Types

- **`repositoryContext`** -- Scenario state for all repository validation steps
- **`moduleIsolationContext`** -- Tracks Go import violations per module
- **`testSanityContext`** -- Validates test discovery matches raw file scans
- **`commandDocsContext`** -- Checks CLI commands have reference documentation
- **`goVersionContext`** -- Verifies Go version consistency across config files
- **`nodeVersionContext`** -- Verifies Node.js version consistency across actions

## Patterns

- Godog runner: Single `TestRepositoryFeatures` entry point delegates to `RegisterSteps`
- Lazy repo root: `ensureRepoRoot()` initializes shared state on first step invocation
- Cached file lists: Uses `OriginalRepoCache` from godog adapter to avoid repeated `git ls-files`
- AST-based import scanning: Parses Go files with `go/parser` to detect forbidden cross-module imports
- Passive assertions: "if any X, I should see Y" steps are no-ops; errors from prior steps carry details

## Internal Structure

| File | Responsibility |
| --- | --- |
| godog_test.go | Test suite entry point wiring Godog options |
| steps_test.go | Core steps: module tidy, markdown, feature tags, hierarchy, file ownership |
| steps_modules_test.go | Module-level validation steps |
| steps_module_isolation_test.go | Go module import isolation enforcement |
| steps_test_sanity_test.go | Test discovery count validation against raw scans |
| steps_command_docs_test.go | CLI command documentation coverage checks |
| steps_go_version_test.go | Go version consistency across go.work and actions |
| steps_node_version_test.go | Node.js version consistency across actions |
| steps_scripts_test.go | Script file location and naming conventions |
| steps_release_test.go | Release directory structure and changelog validation |

## Dependencies

- `go/adapters/godog` -- Shared test context and scenario initializer
- `go/adapters/eac` -- EAC adapter for executing CLI commands in specs
- `go/adapters/cucumber` -- Test type descriptor registration
- `go/adapters/gotest` -- Test type descriptor registration
- `go/core/repository` -- Repository root discovery and file-module mapping
- `go/core/testing` -- Test discovery and enrichment
- `go/core/config` -- EAC configuration loading
- `go/core/docsync` -- Command documentation sync scanning
- `contracts/core/0.1.0` -- Module contract port interface

## Role in System

This package is the primary structural guard for the repository. Its Gherkin scenarios (in `specs/repository/`) enforce that Go modules remain tidy, dependency contracts are bidirectional, files have exactly one owning module, CLI commands have documentation, tool versions stay consistent, and documentation caches are up to date. Failures in these specs block CI merges, preventing structural drift.

## Code Health

### Tech Debt
- TODO at steps_test.go:488 for proper Gherkin parsing and tag conflict detection remains unimplemented
- TODO at steps_test.go:550 for link checking remains unimplemented
- steps_test.go (~593 lines) is the largest step file, mixing hierarchy, file ownership, markdown, and feature tag steps

### Pain Points
- Global mutable context variables (`cmdDocsCtx` in steps_command_docs_test.go:32, `modIsoCtx` in steps_module_isolation_test.go:30) create implicit shared state across scenarios
- Global mutable `discoveryCache` in steps_test_sanity_test.go:28 caches test discovery results as package-level state

### Optimization Opportunities
- Extract steps_test.go into focused step files (e.g., hierarchy steps, ownership steps, markdown steps) to reduce the ~593-line single file
- Implement the TODO stubs for Gherkin tag conflict detection and link checking to complete planned validation coverage
