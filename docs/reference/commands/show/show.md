# Show

<!-- book:cmd show -->

## Subcommands

All show commands return formatted tables and text for interactive terminal use.

See the [show Commands Category](../categories/show.md) for complete list of subcommands.

## Common Patterns

```bash
# Filter output with grep
r2r eac show modules | grep "src-auth"

# Count items
r2r eac show modules | wc -l

# View with pager
r2r eac show dependencies | less
```

## See Also

- [get Commands](../categories/get.md) - JSON output for automation
- [show Commands Category](../categories/show.md) - All show commands
