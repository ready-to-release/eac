# Test debug

<!-- book:cmd test debug -->

Parses test results (Go test JSON and Cucumber JSON) in `out/test/` and lists all failed tests with their locations and error output.

Results are presented in a table format showing the test name, package, file location (for Cucumber), and error details. If no failures are found, displays a success message.

## Usage

```bash
eac test debug
```

This command takes no flags or arguments. It scans `out/test/**/*.json` for both Go test JSON and `.cucumber.json` files.

## Examples

```bash
# Run tests first, then inspect failures
eac test
eac test debug
```

## See Also

- [test](./test.md) - Run tests
- [show test-summary](../show/test-summary.md)
- [test Commands](../categories/test.md)
