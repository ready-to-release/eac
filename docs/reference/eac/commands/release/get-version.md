# Release get-version

<!-- book:cmd release get-version -->

Reads the `CHANGELOG.md` file for a module and outputs the latest version. The changelog is the source of truth for versioning, making this command useful in CI/CD pipelines.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | bool | `false` | Output as git tag format (`module/version`) |
| `--json` | bool | `false` | Output in JSON format |
| `--path` | string | | Override changelog path (default: `release/<module>/CHANGELOG.md`) |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<module>` | Yes | Module moniker |

## Output

- Plain version string by default (e.g. `0.0.14`)
- Tag format with `--tag` (e.g. `clie/0.0.14`)
- JSON object with `--json` containing module, version, tag, date, and version_type fields

## Examples

```bash
eac release get-version clie              # Output: 0.0.14
eac release get-version clie --tag        # Output: clie/0.0.14
eac release get-version clie --json
```

## See Also

- [release changelog](./changelog.md)
- [release this](./this.md)
- [release Commands](../categories/release.md)
