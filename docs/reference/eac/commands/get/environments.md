# Get environments

<!-- book:cmd get environments -->

Returns environment definitions and configuration from the repository config. Outputs the list of configured environments with their settings.

## Usage

```bash
eac get environments [--as-yaml|--as-json|--as-toml]
```

## Flags

| Flag | Description |
|------|-------------|
| `--as-yaml` | Output as YAML (default) |
| `--as-json` | Output as JSON |
| `--as-toml` | Output as TOML |

## Output

Returns the environments list from the loaded configuration, including environment variables, deployment targets, and environment-specific settings for each defined environment.

## Examples

```bash
# List all environments
eac get environments

# As JSON
eac get environments --as-json
```

## See Also

- [show environments](../show/environments.md) - Formatted table
- [test](../test/test.md)
- [get Commands](../../categories/get.md)
