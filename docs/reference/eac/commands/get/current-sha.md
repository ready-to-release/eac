# get current-sha

<!-- book:cmd get current-sha -->

## Detection Order

1. `--sha` flag (explicit override) - source: `explicit`
2. `GITHUB_SHA` environment variable (GitHub Actions) - source: `ci`
3. `origin/main` HEAD after fetch (local devbox) - source: `devbox`

## Output

Default format outputs just the SHA string.

Shell format outputs:

```text
SHA="abc123def456..."
SOURCE="ci"
```

## See Also

- [pipeline await-ci](../pipeline/await-ci.md)
- [release check-ci](../release/check-ci.md)
- [get Commands](../get/index.md)
