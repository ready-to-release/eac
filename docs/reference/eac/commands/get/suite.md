# Get suite

<!-- book:cmd get suite -->

Returns a test suite definition and configuration in structured format, including metadata, test selection criteria, and the list of included tests.

## Usage

```bash
eac get suite <suite-moniker> [--as-yaml|--as-json|--as-toml]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `suite-moniker` | Yes | Test suite moniker (e.g. `unit`, `integration`, `acceptance`) |

## Flags

| Flag | Description |
|------|-------------|
| `--as-yaml` | Output as YAML (default) |
| `--as-json` | Output as JSON |
| `--as-toml` | Output as TOML |

## Output

Returns a suite report containing:

- Suite metadata (moniker, name, description)
- Test selection criteria (tags, modules, patterns)
- Suite-level configuration and settings
- List of included tests with their metadata and module associations

If an invalid suite moniker is provided, the command lists available suites on stderr.

## Examples

```bash
# Get the unit test suite definition
eac get suite unit

# Integration suite as JSON
eac get suite integration --as-json

# Acceptance suite as TOML
eac get suite acceptance --as-toml
```

## See Also

- [show suite](../show/suite.md) - Formatted display
- [test suite](../test/suite.md) - Run suite
- [test list-suites](../test/list-suites.md)
