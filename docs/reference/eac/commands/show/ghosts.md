# Show Ghosts

<!-- book:cmd show ghosts -->

Displays discovered ghost entities (files and directories matching the ghost naming convention) in a markdown-formatted report, grouped by module with summary statistics.

## Usage

```bash
eac show ghosts [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--type <type>` | string | Filter by type (`file` or `directory`) |
| `--module <name>` | string | Filter to ghosts in a specific module |
| `--unowned` | bool | Only show ghosts not associated with any module |

## Output Sections

1. **Summary** - Table with metrics: Total Ghosts, Files, Directories, Modules with Ghosts, Unowned. Also shows the configured prefix and patterns.
2. **Ghosts by Module** - For each module that owns ghosts, a table with columns: Path, Type, Ghost Name.
3. **Unowned Ghosts** - Ghosts not associated with any module, same table format.

If no ghosts are found, prints a message with the current configuration prefix.

## Examples

```bash
# Show all ghosts
eac show ghosts

# Show ghosts in core module
eac show ghosts --module core

# Show only ghost directories
eac show ghosts --type directory

# Show only unowned ghosts
eac show ghosts --unowned
```

## See Also

- [get ghosts](../get/ghosts.md)
- [Ghost Tracking](../../../../explanation/continuous-delivery/workflow/ghost-tracking.md)
- [show Commands](../categories/show.md)
