# release pending

<!-- book:cmd release pending -->

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

## See Also

- [release this](./this.md)
- [release tag-pending](./tag-pending.md)
- [release Commands](../release/index.md)
