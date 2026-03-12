# Show Units

<!-- book:cmd show units -->

Displays units of work for a framework in a human-readable markdown report with cache status visualization.

## Usage

```bash
eac show units <build|test|lint|scan> [flags]
```

The framework argument is required and must be one of: `build`, `test`, `lint`, `scan`.

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--module <name>` | string | Filter to a specific module |
| `--component <name>` | string | Filter to a specific component |
| `--cached` | bool | Only show cached (up-to-date) units |
| `--stale` | bool | Only show stale units |
| `--container` | bool | Only show container-based units |
| `--host` | bool | Only show host-installed units |

## Output Sections

The report includes:

1. **Summary** - Statistics table with total units, cached/stale/new counts, and container vs host counts.
2. **Cache Status Overview** - Breakdown of cached, stale, and new units with percentages.
3. **Units by Module** - Units grouped by module, with columns: Status (icon), Component, Tool, Weight, Type (host/container), Last Run (relative time), Reason (why stale).
4. **Stale Units** - Dedicated section listing only units that need execution, with their staleness reasons. Only shown if stale units exist.
5. **Dependency Chain** - Intra-module dependencies between units. Only shown if dependencies exist.

## Examples

```bash
# All build units
eac show units build

# Test units for a specific module
eac show units test --module eac

# Only stale lint units
eac show units lint --stale

# Container-based scan units
eac show units scan --container
```

## See Also

- [get units](../get/units.md) - Structured data output
- [show components](./components.md) - Component-level view
