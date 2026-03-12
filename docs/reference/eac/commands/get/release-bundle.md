# Get release-bundle

<!-- book:cmd get release-bundle -->

Returns release bundle configuration with module groupings, versioning info, and optionally resolved versions from GitHub releases. The bundle is defined in a module's `release_bundle` configuration.

## Usage

```bash
eac get release-bundle [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--with-versions` | bool | Resolve current versions from GitHub releases |
| `--title-only` | bool | Output only the resolved release title (implies `--with-versions`) |
| `--format` | string | Output format: `markdown`, `flat`, `shell`, `table` |

## Output Formats

**Default (YAML/JSON/TOML):**

```yaml
title_format: "EAC v{eac_version} / CLIE v{clie_version}"
headline:
  eac:
    moniker: eac
    versioning: calver
    version: "2025.03"    # with --with-versions
    tag: eac/2025.03
categories:
  - name: CLI Tools
    modules:
      - moniker: clie
        versioning: semver
        version: "1.0.0"
```

**`--format shell`** (for `eval`):

```bash
RELEASE_MAP='{"clie":{"tag":"clie/1.0.0","version":"1.0.0"}}'
ALL_RELEASED="true"
```

**`--format flat`**: `moniker|version|tag|category` per line.

**`--format markdown`**: Release notes with tables per category.

**`--format table`**: Markdown summary table with status icons.

## Examples

```bash
# Bundle structure only
eac get release-bundle

# With resolved versions
eac get release-bundle --with-versions

# Release title string
eac get release-bundle --title-only

# Shell variables for CI
eval $(eac get release-bundle --format shell)
```

## See Also

- [get](get.md)
- [release](../release/index.md)
