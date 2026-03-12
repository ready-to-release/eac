# Get files

<!-- book:cmd get files -->

Returns repository files with their owning module information. Each file is mapped to its module(s) based on module domain configuration.

## Usage

```bash
eac get files [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--changed-only` | bool | Only modified/unstaged files |
| `--staged-only` | bool | Only staged files |
| `--module` | string | Filter to files owned by a specific module |
| `--pattern` | string | Filter files matching a glob pattern |
| `--as-yaml` | bool | Output as YAML (default) |
| `--as-json` | bool | Output as JSON |
| `--as-toml` | bool | Output as TOML |

`--changed-only` and `--staged-only` are mutually exclusive.

## Output Structure

```yaml
- name: go/core/config/config.go
  modules:
    - core
- name: go/cli/clie/main.go
  modules:
    - clie
```

## Examples

```bash
# All files
eac get files

# Files changed since last commit
eac get files --changed-only

# Files in core module matching Go sources
eac get files --module core --pattern "**/*.go"
```

## See Also

- [show files](../show/files.md) - Formatted table
- [get changed-modules](./changed-modules.md)
