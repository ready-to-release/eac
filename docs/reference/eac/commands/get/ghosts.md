# get ghosts

<!-- book:cmd get ghosts -->

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

## See Also

- [show ghosts](../show/ghosts.md)
- [Ghost Tracking](../../../../explanation/continuous-delivery/workflow/ghost-tracking.md)
- [get Commands](../get/index.md)
