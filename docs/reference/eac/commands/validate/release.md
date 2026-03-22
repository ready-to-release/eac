# Validate release

<!-- book:cmd validate release -->

Validates that changelog files follow the Keep a Changelog format and contain valid version entries.

## Usage

```bash
eac validate release <module> [flags]
eac validate release --all [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `module` | Module moniker to validate (required unless `--all`) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--all` | bool | Validate all modules with changelogs |
| `--json` | bool | Output results in JSON format |

## What It Checks

- **File exists** -- Changelog exists at the path defined in the module contract.
- **Valid header** -- Title follows Keep a Changelog conventions.
- **Version format** -- Each entry uses valid semver (`x.y.z`) or calver (`YYYY.MM.DD`) format.
- **No duplicates** -- No two entries share the same version number.
- **Descending order** -- Versions are ordered newest-first (checked by date).
- **Non-empty versions** -- Each version has at least one entry (warning).

## Examples

```bash
# Single module
eac validate release clie

# All modules
eac validate release --all

# JSON output for CI
eac validate release clie --json
```

## Common Errors

- **changelog not found** -- The file does not exist at the contract-specified path.
- **invalid version format** -- A version does not match semver or calver patterns.
- **duplicate version** -- The same version number appears more than once.

## See Also

- [release changelog](../release/changelog.md)
- [release this](../release/this.md)
- [validate release-version](./release-version.md)
- [validate Commands](../../categories/validate.md)
