# Release tag-pending

<!-- book:cmd release tag-pending -->

Scans changelog files for version entries and checks whether the corresponding git tag exists. Returns versions that need tagging. This is used by CI to detect merged releases that need tags created.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | `false` | Check all modules with changelogs |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<module>` | Yes (unless `--all`) | Module moniker |

## Output

JSON containing:

- `module` -- module name
- `version` -- latest changelog version
- `tag` -- expected git tag (e.g. `clie/0.0.14`)
- `needs_tag` -- `true` if the tag does not exist

When using `--all`, returns a report with only the modules that need tags and a `has_pending` flag.

## Examples

```bash
eac release tag-pending clie        # Check single module
eac release tag-pending --all       # Check all modules
```

## See Also

- [release pending](./pending.md)
- [release this](./this.md)
- [release Commands](../../categories/release.md)
