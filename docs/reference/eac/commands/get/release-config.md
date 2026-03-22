# Get release-config

<!-- book:cmd get release-config -->

Derives release configuration for a module from the core config system. All values are computed from `repository.yml` and `blueprints.yml`.

## Usage

```bash
eac get release-config --module <moniker> [--format shell|github-output]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--module` | string | Module moniker (required) |
| `--format` | string | Output format: `shell`, `github-output`, or standard get formats (yaml/json/toml) |

## Output Variables

| Variable | Description |
|----------|-------------|
| `RELEASE_TYPE` | `cli-binary`, `container`, `docs-site`, `bundle`, or `none` |
| `VERSION_TYPE` | `semver`, `calver`, or `none` |
| `HAS_EVIDENCE` | Whether the module has evidence-book components |
| `AWAIT_MODULE_RELEASES` | `true` for bundle releases (awaits dependency releases) |

## Release Type Resolution

| Condition | Release Type |
|-----------|-------------|
| `bundle` release type | `bundle` |
| `published` + dockerfile with push | `container` |
| `published` + docs-site component | `docs-site` |
| `published` + go component with binary_name | `cli-binary` |
| `internal` or `none` | `none` |

## Examples

```bash
# Get release config as YAML
eac get release-config --module eac-cli

# Shell format for CI scripts
eval $(eac get release-config --module core --format shell)
echo "Release type: $RELEASE_TYPE"

# GitHub Actions output
eac get release-config --module docs --format github-output >> $GITHUB_OUTPUT
```

## See Also

- [get release-status](./release-status.md)
- [get release-notes](./release-notes.md)
- [get Commands](../../categories/get.md)
