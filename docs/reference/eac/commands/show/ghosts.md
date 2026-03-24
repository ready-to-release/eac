# show ghosts

<!-- book:cmd show ghosts -->

## Output Sections

1. **Summary** - Table with metrics: Total Ghosts, Files, Directories, Modules with Ghosts, Unowned. Also shows the configured prefix and patterns.
2. **Ghosts by Module** - For each module that owns ghosts, a table with columns: Path, Type, Ghost Name.
3. **Unowned Ghosts** - Ghosts not associated with any module, same table format.

If no ghosts are found, prints a message with the current configuration prefix.

## See Also

- [get ghosts](../get/ghosts.md)
- [Ghost Tracking](../../../../explanation/continuous-delivery/workflow/ghost-tracking.md)
- [show Commands](../show/index.md)
