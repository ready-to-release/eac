# Get specs unused-steps

<!-- book:cmd get specs unused-steps -->

Scans godog step definition files and compares them against feature files to detect unused step definitions. Shared steps from the internal steps file are checked against all impl/specs pairs.

## Usage

```bash
eac get specs unused-steps [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `-v`, `--verbose` | bool | Show detailed output including all scanned files |
| `-m`, `--module` | string | Only analyze a specific module |

## Output

Prints unused step definitions grouped by file, with line numbers and truncated patterns. Summary reports total unused count.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No unused steps found |
| 1 | Unused steps detected (or error) |

## Examples

```bash
# Scan all modules
eac get specs unused-steps

# Verbose output
eac get specs unused-steps --verbose

# Single module
eac get specs unused-steps --module eac
```

## See Also

- [get tests](./tests.md)
