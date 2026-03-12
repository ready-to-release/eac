# Show Modules

<!-- book:cmd show modules -->

Displays all module contracts in a formatted table, sorted by dependency depth, then group, then declaration order. Optionally includes artifact statistics.

## Usage

```bash
eac show modules [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--with-artifacts` | bool | Include artifact statistics (count, missing, overrides) |
| `--markdown` | bool | Output a pipe-separated Markdown table instead of console table |

## Output Sections

A table with columns:

| Column | Description |
|--------|-------------|
| Moniker | Module identifier |
| Group | Module group |
| Components | Component types display string |

With `--with-artifacts`, three additional columns appear:

| Column | Description |
|--------|-------------|
| Artifacts | Total artifact count for current platform |
| Missing | Number of artifacts not found on disk |
| Overrides | Number of metadata overrides applied |

## Examples

```bash
# List all modules
eac show modules

# List modules with artifact stats
eac show modules --with-artifacts

# Output as Markdown table
eac show modules --markdown
```

## See Also

- [get modules](../get/modules.md) - JSON output
- [show component-kinds](./component-kinds.md) - Group by type
- [show dependencies](./dependencies.md) - Dependency graph
