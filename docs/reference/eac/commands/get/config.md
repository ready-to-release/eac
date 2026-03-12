# Get config

<!-- book:cmd get config -->

Returns the fully loaded repository configuration, including modules, component kinds, environments, and testing settings with all defaults applied.

## Usage

```bash
eac get config [flags]
```

## Output Structure

```yaml
modules:       # Module contracts (monikers, types, roots, dependencies)
component_kinds: # Component type definitions (build/test/deploy capabilities)
environments:  # Environment definitions
testing:       # Testing configuration (tags and suites)
```

## Examples

```bash
eac get config
eac get config --as-json
```

## See Also

- [show config](../show/config.md) - Formatted display
- [init](../init/init.md) - Configure AI provider
