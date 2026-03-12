# Validate specs

<!-- book:cmd validate specs -->

Validates Gherkin specification files against quality contracts, checking structure, tags, step formatting, and content quality.

## Usage

```bash
eac validate specs <path> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `path` | File or directory to validate (required). Directories are scanned recursively. |

## Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--quiet` | `-q` | bool | Show only errors and warnings |
| `--verbose` | `-v` | bool | Show detailed output with metadata |
| `--format` | `-f` | string | Output format: `text` (default) or `json` |
| `--fix` | | bool | Auto-fix correctable issues (single file only) |
| `--no-check-tags` | | bool | Disable tag validation |

## What It Checks

- **Structure** -- Proper Feature/Rule/Scenario hierarchy.
- **Tags** -- Tags conform to project conventions (unless `--no-check-tags`).
- **Step formatting** -- Given/When/Then steps follow project standards.
- **Content quality** -- Scenarios have meaningful descriptions and steps.

## Examples

```bash
# Validate all specs
eac validate specs specs/

# Validate a single file
eac validate specs specs/eac/create/specification.feature

# JSON output for CI
eac validate specs specs/ --format json

# Auto-fix correctable issues
eac validate specs specs/eac/create/specification.feature --fix
```

## Common Errors

- **File or directory not found** -- The specified path does not exist.
- **Structure error** -- Invalid Gherkin hierarchy (e.g., Scenario outside Feature).
- **Tag validation error** -- A tag does not conform to project conventions.

## See Also

- [show specs](../show/specs.md)
- [validate](./validate.md)
