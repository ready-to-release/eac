# Release execute-layers

<!-- book:cmd release execute-layers -->

Executes pending releases in dependency order, processing one layer at a time. Each layer contains modules that can be released in parallel, and layers are processed sequentially to respect dependency constraints.

For each module in a layer, the command dispatches the corresponding `release-{module}.yaml` workflow. Semver releases pass a `version` input; calver releases auto-generate versions. After dispatching all modules in a layer, it waits for completion before proceeding to the next layer.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--layers` | string | | JSON array of release layers |
| `--layers-file` | string | | File containing JSON array of release layers |
| `--timeout` | int | `900` | Timeout per release in seconds |
| `--dry-run` | bool | `false` | Preview without dispatching workflows |

## Layers JSON Format

```json
[[{"module":"docs","version":"2025.0116.1430","tag":"docs/2025.0116.1430","type":"calver"}],
 [{"module":"clie","version":"1.0.0","tag":"clie/1.0.0","type":"semver"}]]
```

Each inner array is a layer. Layers are processed in order. Modules within a layer are dispatched concurrently.

## Output

- Progress messages for each module dispatch and completion
- Exit code 0 if all releases succeed
- Exit code 1 if any release fails

## Examples

```bash
# Execute layers from inline JSON
eac release execute-layers --layers '[[{"module":"docs","version":"2025.0116.1430","tag":"docs/2025.0116.1430","type":"calver"}]]'

# Execute layers from a file
eac release execute-layers --layers-file layers.json

# Preview without dispatching
eac release execute-layers --layers-file layers.json --dry-run
```

## See Also

- [release pending](./pending.md)
- [release](../release/index.md)
