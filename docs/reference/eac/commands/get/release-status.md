# Get release-status

<!-- book:cmd get release-status -->

Checks GitHub releases for a list of modules and returns their release status, including latest version and tag. Tags follow the `{moniker}/{version}` convention.

## Usage

```bash
eac get release-status --modules "<module-list>" [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--modules` | string | Space-separated list of modules to check (required) |
| `--format` | string | Output format: `json` (default) or `shell` |

## Output Structure

**Default (JSON/YAML):**

```yaml
modules:
  clie:
    tag: clie/1.0.0
    version: "1.0.0"
    released: true
  docs:
    tag: ""
    version: ""
    released: false
released: [clie]
missing: [docs]
```

**`--format shell`:**

```bash
RELEASED="clie eac-ext"
MISSING="docs"
ALL_RELEASED="false"
```

## Examples

```bash
eac get release-status --modules "clie eac-ext docs"
eac get release-status --modules "clie eac-ext" --format shell
```

## See Also

- [get](get.md)
- [release](../release/index.md)
