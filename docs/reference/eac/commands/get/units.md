# get units

<!-- book:cmd get units -->

## Output Structure

Returns a list of units, each containing:

- `id` - Unit identifier
- `module` - Parent module moniker
- `component` - Component name
- `tool` - Tool used for execution
- `weight` - Execution weight/priority
- `container` - Whether it runs in a container
- `host_installed` - Whether it uses host-installed tools
- `cache_status.state` - One of: `up_to_date`, `stale`, `new`
- `cache_status.reason` - Why a unit is stale (e.g., source changed, previous failure)
- `cache_status.executed_at` - Timestamp of last execution
- `dependencies` - List of unit dependencies

If any components were skipped (e.g., missing tool), a `skipped` array is included alongside `units`.

## See Also

- [show units](../show/units.md) - Human-readable markdown report
