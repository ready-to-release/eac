# get

Parent command that retrieves repository data in structured output formats (YAML, JSON, TOML). Houses over 30 subcommands spanning configuration, modules, files, CI/CD, testing, releases, and more.

## Key Types

- **`ConfigOutput`** -- Structured output of all loaded configuration
- **`ArtifactBuildModes`** -- Artifact lists for default and all build modes
- **`moduleFilters`** -- Parsed filter flags for module queries
- **`filterOptions`** -- Parsed filter flags for file queries
- **`OutputFormat`** -- Desired output format (internal helper)

## Patterns

- Registry-based subcommand dispatch: each subcommand file calls `registry.Register` in `init()`
- Shared `ExecuteGetCommand` helper: wraps flag parsing, data fetching, and rendering
- TUI selector fallback: interactive subcommand picker when run without arguments
- Comment-driven metadata: command, flags, and grouping declared via structured comments

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| get.go | Parent command entry point with TUI and subcommand dispatch |
| internal/renderer.go | Shared output format parsing and multi-format rendering |
| config.go | Retrieve loaded repository configuration |
| modules.go | List module contracts with optional filtering |
| dependencies.go | Full dependency graph with PlantUML/Mermaid output |
| files.go | Repository files with module ownership and filters |
| changed-modules.go | Modules affected by git diff or stdin file list |
| artifacts.go | Resolved build artifacts for a module with platform support |
| tests.go | Test definitions and results retrieval |
| release-notes.go | Release notes for modules |
| specs.go | Specification retrieval |
| ci-dispatch.go | CI dispatch information |
| evidence-ci-runs.go | Evidence CI run data for compliance |

## Dependencies

- `cli/eac/help` -- help text rendering for parent command
- `cli/eac/impl/internal` -- shared artifact resolution types
- `cli/eac/impl/get/internal` -- output format parsing and rendering
- `clibase/registry` -- command registration and subcommand discovery
- `clibase/flags` -- flag validation from registry metadata
- `clibase/render` -- JSON, TOML, and custom renderer support
- `clibase/gitexec` -- git command execution
- `adapters/tui` -- TUI detection and subcommand options
- `adapters/tui/selector` -- interactive command selector
- `core/config` -- configuration loading
- `core/repository` -- repository root, dependency graphs, changed modules

## Role in System

The `get` package is the primary data retrieval interface in `eac-cli`, exposing repository structure, configuration, and CI/CD state as machine-readable output. Each subcommand follows a uniform pattern of loading data through core domain services and rendering via the shared `ExecuteGetCommand` helper, making output format selection consistent across all queries.

## Code Health

### Tech Debt
- Only 8 test files cover ~38 source files; most subcommands (modules, files, dependencies, tests, etc.) have no unit tests
- 11 subcommands do not use the shared `ExecuteGetCommand` helper (e.g., `artifacts.go`, `binary-sizes.go`, `files-by-module.go`), implementing their own output formatting

### Pain Points
- Pattern duplication across individual command files: each file repeats the same init/register, flag-parse, load-config, render cycle despite 28/39 using the shared helper
- `files.go` (236 lines) and `artifacts.go` (260 lines) are notably larger than the average subcommand (~80-100 lines), suggesting they carry extra complexity

### Optimization Opportunities
- Migrate the remaining 11 non-helper subcommands to `ExecuteGetCommand` for consistent output format handling -- moderate effort, improves uniformity
- Add table-driven tests covering the core data-fetch logic for the untested subcommands -- moderate effort, high coverage gain
