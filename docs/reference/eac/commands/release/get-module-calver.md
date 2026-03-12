# Release get-module-calver

<!-- book:cmd release get-module-calver -->

Generates a calendar-versioned (calver) tag for a module. The format follows `prefix/YYYY.MMDD.HHMM` using the current UTC time.

By default the command only outputs the tag name. Use `--create` to actually create the git tag.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--create` | bool | `false` | Create the git tag (default: just output tag name) |
| `--push` | bool | `false` | Push the tag to remote after creation (requires `--create`) |
| `--dry-run` | bool | `false` | Show what would be done without creating/pushing |

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<prefix>` | Yes | Module name used as the tag prefix (e.g. `docs`) |

## Output

- Tag name in format `prefix/YYYY.MMDD.HHMM` (default)
- Git tag created locally if `--create` is specified
- Tag pushed to remote if `--push` is also specified

## Examples

```bash
# Output the tag name only
eac release get-module-calver docs           # Output: docs/2025.0312.1630

# Create the tag locally
eac release get-module-calver docs --create

# Create and push
eac release get-module-calver docs --create --push

# Preview
eac release get-module-calver docs --create --dry-run
```

## See Also

- [release this](./this.md)
- [release clie](./clie.md)
- [release Commands](../categories/release.md)
