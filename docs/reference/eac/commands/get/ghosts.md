# Get ghosts

<!-- book:cmd get ghosts -->

Returns structured data about all ghost (dark launch) entities discovered in the repository. Ghosts are files and directories matching the configured naming convention (default: `ghost-*` prefix).

## Usage

```bash
eac get ghosts [--type file|directory] [--module <name>] [--unowned] [--as-yaml|--as-json|--as-toml]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--type` | string | Filter by entity type: `file` or `directory` |
| `--module` | string | Filter to ghosts in a specific module |
| `--unowned` | bool | Only show ghosts not owned by any module |
| `--as-yaml` | | Output as YAML (default) |
| `--as-json` | | Output as JSON |
| `--as-toml` | | Output as TOML |

## Output

Returns a `GhostReport` containing:

- **ghosts**: List of discovered ghost entities with paths, types, and owning modules
- **summary**: Aggregate statistics (counts by type, owned vs unowned)
- **config**: The effective ghost tracking configuration

## Use Cases

Ghost entities enable:

- **Dark launching**: Code deployed but inactive
- **L4 monitoring**: Hidden observability probes
- **Feature toggles**: Without a full feature flag system

## Examples

```bash
# All ghosts
eac get ghosts

# Only ghost files
eac get ghosts --type file

# Ghosts in core module
eac get ghosts --module core

# Unowned ghosts as JSON
eac get ghosts --unowned --as-json
```

## See Also

- [show ghosts](../show/ghosts.md)
- [Ghost Tracking](../../../../explanation/continuous-delivery/workflow/ghost-tracking.md)
- [get Commands](../categories/get.md)
