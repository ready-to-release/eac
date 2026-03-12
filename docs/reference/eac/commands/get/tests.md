# Get tests

<!-- book:cmd get tests -->

Returns all discovered tests with metadata in structured format, including test locations, types, tags, and aggregate counts.

## Usage

```bash
eac get tests [--as-yaml|--as-json|--as-toml]
```

## Flags

| Flag | Description |
|------|-------------|
| `--as-yaml` | Output as YAML (default) |
| `--as-json` | Output as JSON |
| `--as-toml` | Output as TOML |

## Output Fields

| Field | Description |
|-------|-------------|
| `total_tests` | Total number of discovered tests |
| `tests` | List of all tests with metadata |

Each test entry includes:

- Test name and location (module, package, file)
- Test type (unit, integration, e2e, etc.)
- Tags and markers

## Examples

```bash
# Get all tests as YAML
eac get tests

# As JSON for analysis
eac get tests --as-json
```

## See Also

- [show tests](../show/tests.md) - Formatted table
- [test](../test/test.md)
- [get suite](./suite.md)
