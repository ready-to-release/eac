# get release-bundle

<!-- book:cmd get release-bundle -->

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

## See Also

- [get](get.md)
- [release](../release/index.md)
