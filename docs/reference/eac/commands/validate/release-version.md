# Validate release-version

<!-- book:cmd validate release-version -->

Validates that a version string is a valid semantic version (semver) format. Silent on success, prints an error on failure.

## Usage

```bash
eac validate release-version <version>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `version` | The version string to validate (required) |

## Validation Rules

- Must match `MAJOR.MINOR.PATCH` format (e.g., `1.2.3`).
- Must not have a `v` or `V` prefix.
- No leading zeros (except `0` itself).

## Examples

```bash
# Valid
eac validate release-version 1.2.3
eac validate release-version 0.1.0

# Invalid
eac validate release-version v1.2.3    # 'v' prefix
eac validate release-version 1.2       # missing patch
eac validate release-version 01.2.3    # leading zero
```

## See Also

- [validate release](./release.md)
- [release get-version](../release/get-version.md)
- [validate Commands](../categories/validate.md)
