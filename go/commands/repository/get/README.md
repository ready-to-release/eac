# get

Parent command that retrieves repository data in structured output formats (YAML, JSON, TOML). Houses over 30 subcommands spanning configuration, modules, files, CI/CD, testing, releases, and more.

## Key Types

- **`ConfigOutput`** -- Structured output of all loaded configuration
- **`ArtifactBuildModes`** -- Artifact lists for default and all build modes
- **`moduleFilters`** -- Parsed filter flags for module queries
- **`filterOptions`** -- Parsed filter flags for file queries
- **`OutputFormat`** -- Desired output format (internal helper)

## Patterns

- Table-driven command registration: `commands.go` registers all subcommands via `RegisterAll()`
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

The `get` package is the primary data retrieval interface in `eac`, exposing repository structure, configuration, and CI/CD state as machine-readable output. Each subcommand follows a uniform pattern of loading data through core domain services and rendering via the shared `ExecuteGetCommand` helper, making output format selection consistent across all queries.

## Code Health

### Tech Debt
- 8 subcommands bypass `ExecuteGetCommand` due to legitimate special output requirements (see below)

### Subcommands Not Using ExecuteGetCommand

The following subcommands intentionally bypass the shared helper. Each has a documented reason:

| Subcommand | Reason |
| --- | --- |
| `binary-sizes.go` | Default output is shell variables (`SIZE_X="12.3"`) for `eval`; also supports markdown table format -- none of these map to YAML/JSON/TOML |
| `book-description.go` | Returns a single plain-text string (book title); no structured data to render |
| `cli-release-notes.go` | Generates freeform markdown release notes; output is a text document, not structured data |
| `current_sha.go` | Returns a single SHA string or shell variables (`SHA="..." SOURCE="..."`); no structured data |
| `files-by-module.go` | Parses JSON from environment variable and outputs file lists or shell variables; data flow is inverse (JSON in, plain text out) |
| `module-ci-workflow.go` | Returns a single filename string; no structured data to render |
| `module-trigger-reason.go` | Returns a single human-readable reason string; no structured data to render |
| `token-size.go` | Uses threshold-based exit codes (returns 1 when files exceed limit); default output is `file: N tokens` text; `ExecuteGetCommand` always returns 0 on success |

### Subcommands With Dual-Mode Output

These subcommands use `ExecuteGetCommand` for their default structured output path but have additional special-purpose formats:

| Subcommand | Special Formats | Helper Used For |
| --- | --- | --- |
| `ci-config.go` | `--format shell`, `--format github-output` | Default (YAML/JSON/TOML) |
| `release-config.go` | `--format shell`, `--format github-output` | Default (YAML/JSON/TOML) |
| `ci-workflows.go` | `--format space`, `--format list` | Default (YAML/JSON/TOML) |
| `release-status.go` | `--format shell` | Default (YAML/JSON/TOML) |

### Pain Points
- No files over 300 lines; largest production files are files.go and artifacts.go at ~240 lines each

### Optimization Opportunities
- Most commands now have corresponding _test.go files; remaining untested commands are validated via BDD scenarios
