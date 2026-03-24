# show units

<!-- book:cmd show units -->

## Output Sections

The report includes:

1. **Summary** - Statistics table with total units, cached/stale/new counts, and container vs host counts.
2. **Cache Status Overview** - Breakdown of cached, stale, and new units with percentages.
3. **Units by Module** - Units grouped by module, with columns: Status (icon), Component, Tool, Weight, Type (host/container), Last Run (relative time), Reason (why stale).
4. **Stale Units** - Dedicated section listing only units that need execution, with their staleness reasons. Only shown if stale units exist.
5. **Dependency Chain** - Intra-module dependencies between units. Only shown if dependencies exist.

## See Also

- [get units](../get/units.md) - Structured data output
- [show components](./components.md) - Component-level view
