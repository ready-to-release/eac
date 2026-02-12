# validate

Parent command that validates repository contracts, configuration, dependencies, specifications, and compliance documents. Houses subcommands covering schema validation, code quality checks, OSCAL risk documents, and Gherkin spec quality.

## Key Types

- **`ConfigValidationResult`** -- Result of multi-phase configuration validation
- **`ConfigIssue`** -- Single validation error or warning with file and line
- **`ConfigFileInfo`** -- Metadata about a loaded config file and its layer
- **`ValidateConfig`** -- Parsed flags for spec validation (path, format, quiet)
- **`ValidationResult`** -- Validation result for a single spec file
- **`FixResult`** -- Result of auto-fix operations on spec files
- **`FixedIssue`** -- Description of a single applied fix

## Patterns

- Table-driven command registration: `commands.go` registers all subcommands via `RegisterAll()`
- TUI selector fallback: interactive subcommand picker when run without arguments
- Multi-phase validation: file checks, schema validation, cross-reference validation
- Multi-format output: text (human-readable), JSON (machine), GitHub Actions annotations
- Auto-fix support: `--fix` flag applies correctable changes and re-validates

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | Table-driven registration of all validate subcommands via `RegisterAll()` |
| validate.go | Parent command entry point with TUI and subcommand dispatch |
| config.go | Multi-phase config validation (files, schemas, cross-references) |
| contracts.go | Schema validation of repository YAML contracts |
| specs.go | Gherkin specification validation entry point |
| specs_validate.go | Gherkin specification validation logic |
| specs_format.go | Gherkin specification output formatting |
| specs_fix.go | Gherkin specification auto-fix logic |
| dependencies.go | Module dependency graph validation |
| module-hierarchy.go | Module hierarchy structure validation |
| module-files.go | Module file ownership validation |
| test-tags.go | Test tag definition validation |
| design.go | Design document validation |
| artifacts.go | Build artifact validation |
| books.go | Books configuration validation |
| release-version.go | Semver release version format validation |
| markdown.go | Markdown file validation |
| go-tidy.go | Go module tidiness validation |
| release.go | Release configuration validation |
| version.go | Version format validation |
| control-tags.go | Security control tag validation |
| validation_report.go | Shared `ValidationOutput` formatter for spec and config validation |

## Dependencies

- `cli/eac/help` -- help text rendering for parent command
- `clibase/registry` -- command registration and workspace root
- `clibase/flags` -- flag validation from registry metadata
- `adapters/tui` -- TUI detection and subcommand options
- `adapters/tui/selector` -- interactive command selector
- `core/config` -- configuration loading with schema validation
- `core/validation` -- validation error types and OSCAL validators
- `core/validation/formats/oscal` -- OSCAL catalog and profile validators
- `core/validation/formats/gherkin` -- Gherkin specification validator
- `core/domain` -- validation error formatting
- `core/repository` -- repository root discovery
- `core/paths` -- default version paths

## Role in System

The `validate` package serves as the quality gate in `eac`, ensuring repository contracts, configuration, and specifications conform to their schemas and structural rules. It is invoked both interactively during development and automatically in CI pipelines, with GitHub Actions annotation output enabling inline error reporting on pull requests.

## Code Health

### Tech Debt
- None identified

### Pain Points
- specs_test.go is 931 lines (significantly exceeds 300-line threshold)
- config.go is 463 lines (exceeds 300-line threshold)
- test-tags.go is 307 lines (exceeds 300-line threshold)

### Optimization Opportunities
- None identified
