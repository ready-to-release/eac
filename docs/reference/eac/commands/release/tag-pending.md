# release tag-pending

<!-- book:cmd release tag-pending -->

## Output

JSON containing:

- `module` -- module name
- `version` -- latest changelog version
- `tag` -- expected git tag (e.g. `clie/0.0.14`)
- `needs_tag` -- `true` if the tag does not exist

When using `--all`, returns a report with only the modules that need tags and a `has_pending` flag.

## See Also

- [release pending](./pending.md)
- [release this](./this.md)
- [release Commands](../release/index.md)
