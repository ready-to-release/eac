# Validate version

<!-- book:cmd validate version -->

Validates that a version string matches a specified format (semver or calver).

## Usage

```bash
eac validate version <version> --type <semver|calver> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `version` | The version string to validate (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--type` | string | Version type: `semver` or `calver` (required) |
| `--format` | string | Output format: default or `shell` |

## Version Formats

- **semver** -- `MAJOR.MINOR.PATCH` (e.g., `1.0.0`). A leading `v` prefix is stripped automatically.
- **calver** -- `YYYY.MMDD.HHMM` (e.g., `2025.0116.1430`).

## Examples

```bash
# Semver validation
eac validate version 1.0.0 --type semver

# Calver validation
eac validate version 2025.0116.1430 --type calver

# Shell-parseable output for scripts
eval $(eac validate version 1.0.0 --type semver --format shell)
# Sets: VALID="true" TYPE="semver" VERSION="1.0.0"
```

## See Also

- [validate](validate.md)
- [release](../release/index.md)
