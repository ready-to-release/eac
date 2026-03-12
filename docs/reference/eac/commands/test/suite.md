# Test suite

<!-- book:cmd test --suite concept -->

Runs tests filtered by a named test suite. Suites are logical groupings (unit, integration, acceptance, etc.) discovered via `test list-suites`.

## Usage

```bash
eac test --suite <suite-name> [module...]
```

The `--suite` flag filters which tests are executed based on the suite definitions configured in the repository. All other `test` flags (such as `--skip-cache`, `--tui`, `--turbo`) also apply.

## Examples

```bash
# Run the integration suite across all modules
eac test --suite integration

# Run the unit suite for a specific module
eac test --suite unit eac
```

## See Also

- [test](./test.md) - Test modules directly
- [test list-suites](./list-suites.md) - List available suites
- [show suite](../show/suite.md) - Suite details
- [test Commands](../categories/test.md)
