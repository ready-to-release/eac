# Release eac-ext

<!-- book:cmd release eac-ext -->

Creates a git tag in the format `eac-ext/x.y.z` to trigger the eac-ext release workflow. The tag follows semantic versioning and triggers GitHub Actions to retag and publish the container image.

Identical in behavior to `release clie` but targets the eac-ext module. The `--tag-direct` flag is required to prevent accidental releases.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag-direct` | bool | `false` | Required flag to confirm direct tagging |
| `--dry-run` | bool | `false` | Show what would be done without creating the tag |
| `--push` | bool | `true` | Push the tag to remote after creation |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<version>` | Yes | Semver version string (x.y.z) |

## Output

- Git tag created in format `eac-ext/x.y.z`
- Tag is pushed to remote by default, triggering the release workflow

## Examples

```bash
eac release eac-ext --tag-direct 0.0.7
eac release eac-ext --dry-run 0.0.7
eac release eac-ext --tag-direct --push=false 0.0.7
```

## See Also

- [release clie](./clie.md)
- [release get-version](./get-version.md)
- [get cli-release-notes](../get/cli-release-notes.md)
- [release Commands](../categories/release.md)
