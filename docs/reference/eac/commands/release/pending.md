# Release pending

<!-- book:cmd release pending -->

Analyzes commits since the last release tag and determines whether a module has pending changes that warrant a new version. The release decision is based on file ownership -- if any commits touch files belonging to the module, changes are pending regardless of commit message format.

Outputs JSON with version info, change counts, and the calculated next version.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--quiet` | bool | `false` | Suppress output, use exit code only (0 = has changes, 1 = no changes) |
| `--all` | bool | `false` | Check all modules with changelogs |
| `--published` | bool | `false` | Check only published modules (implies `--all`) |
| `--internal` | bool | `false` | Check only internal modules (implies `--all`) |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<module>` | Yes (unless `--all`) | Module moniker |

## Output

JSON object containing:

- `has_changes` -- whether there are releasable changes
- `current_version` -- the current released version
- `next_version` -- the calculated next version
- `version_type` -- `semver` or `calver`
- `release_type` -- `published`, `internal`, `bundle`, or `none`
- `change_summary` -- breakdown by change type (added, fixed, changed, etc.)
- `commits_total` / `commits_module` -- commit counts

When using `--all`, returns a wrapper with `modules` array and `has_any_change` flag.

## Examples

```bash
eac release pending clie              # Check single module
eac release pending clie --quiet      # Exit code only
eac release pending --all             # Check all releasable modules
eac release pending --published       # Check only published modules
```

## See Also

- [release this](./this.md)
- [release tag-pending](./tag-pending.md)
- [release Commands](../categories/release.md)
