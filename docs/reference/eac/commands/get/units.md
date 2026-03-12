# Get Units

<!-- book:cmd get units -->

Returns structured data about units of work for a specific framework (build, test, lint, or scan), including cache status and dependencies.

## Usage

```bash
eac get units <build|test|lint|scan> [flags]
```

The framework argument is required and must be one of: `build`, `test`, `lint`, `scan`.

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--module <name>` | string | Filter to a specific module |
| `--component <name>` | string | Filter to a specific component |
| `--cached` | bool | Only show cached (up-to-date) units |
| `--stale` | bool | Only show stale (needs execution) units |
| `--container` | bool | Only show container-based units |
| `--host` | bool | Only show host-installed units |

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

## Examples

```bash
# All build units
eac get units build

# Build units for a specific module
eac get units build --module eac

# Only stale test units
eac get units test --stale

# Only cached scan units
eac get units scan --cached

# Container-based lint units
eac get units lint --container
```

## See Also

- [show units](../show/units.md) - Human-readable markdown report
