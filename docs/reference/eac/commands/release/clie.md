# Release clie

<!-- book:cmd release clie -->

Creates a git tag in the format `clie/x.y.z` to trigger the clie release workflow. The tag follows semantic versioning and automatically triggers GitHub Actions to build and publish binaries for multiple platforms.

The `--tag-direct` flag is required to prevent accidental releases. The preferred flow is `release this` followed by commit and push, which lets the workflow create the tag. Use `--tag-direct` only when tagging directly from devbox.

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

- Git tag created in format `clie/x.y.z`
- Tag is pushed to remote by default, triggering the release workflow

## Examples

```bash
# Create and push a release tag
eac release clie --tag-direct 1.0.0

# Preview without creating
eac release clie --dry-run 1.0.0

# Create tag without pushing
eac release clie --tag-direct --push=false 1.0.0
```

## See Also

- [release this](./this.md)
- [release get-module-calver](./get-module-calver.md)
- [release Commands](../../categories/release.md)
