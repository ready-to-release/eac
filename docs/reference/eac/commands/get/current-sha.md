# Get current-sha

<!-- book:cmd get current-sha -->

Returns the current commit SHA using smart auto-detection. Useful in CI scripts and local development to get a consistent SHA reference.

## Usage

```bash
eac get current-sha [--sha <override>] [--format shell]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--sha` | string | Explicit SHA override (skips detection) |
| `--format` | string | `shell` outputs eval-friendly variables; default outputs just the SHA |

## Detection Order

1. `--sha` flag (explicit override) - source: `explicit`
2. `GITHUB_SHA` environment variable (GitHub Actions) - source: `ci`
3. `origin/main` HEAD after fetch (local devbox) - source: `devbox`

## Output

Default format outputs just the SHA string.

Shell format outputs:

```
SHA="abc123def456..."
SOURCE="ci"
```

## Examples

```bash
# Auto-detect SHA
eac get current-sha

# Explicit override
eac get current-sha --sha abc123

# Shell format for eval
eval $(eac get current-sha --format shell)
echo "SHA: $SHA from $SOURCE"
```

## See Also

- [pipeline await-ci](../pipeline/await-ci.md)
- [release check-ci](../release/check-ci.md)
- [get Commands](../categories/get.md)
