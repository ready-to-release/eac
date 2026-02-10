# specs/export/formats

Provides export formatters for manual test scenarios in CSV, JSON, and Markdown formats. Implements a factory pattern for format selection with a common `ExportFormatter` interface.

## Key Types

| Type | Purpose |
|------|---------|
| `ExportFormatter` | Interface for export format implementations (`Write` + `FileExtension`) |
| `ManualTestExport` | Root export structure containing metadata and scenarios |
| `ExportMetadata` | Export context: time, module, release version, git commit, schema version |
| `ExportedScenario` | A single manual test scenario with ID, feature, name, tags, steps, description, and file path |
| `CSVFormatter` | Exports scenarios as CSV with header row |
| `JSONFormatter` | Exports scenarios as pretty-printed JSON |
| `MarkdownFormatter` | Exports scenarios as simplified Markdown with numbered headings |

## Key Functions

| Function | Purpose |
|----------|---------|
| `GetFormatter` | Factory function returning the appropriate formatter for a format string (`json`, `csv`, `markdown`/`md`) |

## Patterns

- **Strategy pattern**: `ExportFormatter` interface with format-specific implementations
- **Factory function**: `GetFormatter` maps format strings to formatter instances
- **Minimal Markdown**: `MarkdownFormatter` intentionally produces minimal output since full documentation is handled by auto-generation

## Internal Structure

| File | Purpose |
|------|---------|
| `formatter.go` | `ExportFormatter` interface and shared types (`ManualTestExport`, `ExportMetadata`, `ExportedScenario`) |
| `factory.go` | `GetFormatter` factory function |
| `csv.go` | `CSVFormatter` implementation |
| `json.go` | `JSONFormatter` implementation |
| `markdown.go` | `MarkdownFormatter` implementation |

## Dependencies

None (uses only standard library).

## Role in System

Used by the `test export-manual` command to export manual test scenarios extracted from Gherkin feature files into portable formats for external consumption (test management tools, spreadsheets, documentation).

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: None identified.
- **Optimization Opportunities**: None identified.
