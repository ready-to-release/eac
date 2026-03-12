# Get ci-workflows

<!-- book:cmd get ci-workflows -->

Discovers all CI workflow files (`ci-*.yaml`) in `.github/workflows/` and returns the module names extracted from their filenames.

## Usage

```bash
eac get ci-workflows [--format space|list|json]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--format` | string | Output format: `space` (default), `list` (one per line), `json` (JSON array) |

Standard `--as-yaml`, `--as-json`, `--as-toml` flags are also supported for structured output.

## Examples

```bash
# Space-separated (default, for shell scripting)
eac get ci-workflows
# Output: clie core docs eac-cli eac-ext

# One module per line
eac get ci-workflows --format list

# JSON array
eac get ci-workflows --format json
# Output: ["clie","core","docs","eac-cli","eac-ext"]
```

## See Also

- [get module-ci-workflow](./module-ci-workflow.md)
- [pipeline status](../pipeline/status.md)
- [pipeline ci](../pipeline/ci.md)
- [get Commands](../categories/get.md)
