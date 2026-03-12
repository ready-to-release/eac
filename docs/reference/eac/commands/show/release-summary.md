# Show release-summary

<!-- book:cmd show release-summary -->

Generate a release summary from layers JSON, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Parses a JSON array of release layers (as produced by the release pipeline) and outputs a formatted summary showing which modules are being released, their versions, and the layer ordering.

## Usage

```
eac show release-summary --layers <json>
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--layers` | string | JSON array of release layers (required) |

The `--layers` value is a JSON array of arrays, where each inner array is a release layer containing module objects with `module`, `version`, `tag`, and `type` fields.

## Output Sections

- **Layer list**: modules grouped by release layer with version and type
- **Tag summary table**: all modules with version, tag, and type columns

## Examples

```bash
# Single-layer release
eac show release-summary \
  --layers '[[{"module":"docs","version":"2025.0116","tag":"docs/v2025.0116","type":"calver"}]]'

# Multi-layer release (dependencies first)
eac show release-summary --layers "$LAYERS_JSON"

# Redirect to GitHub Actions step summary
eac show release-summary --layers "$LAYERS_JSON" >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show](show.md)
- [release](../release/index.md)
