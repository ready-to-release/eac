# Test list-suites

<!-- book:cmd test list-suites -->

Lists all available test suites defined in the repository, showing each suite's moniker, display name, and description.

Test suites are logical groupings of tests (unit, integration, acceptance, etc.) that can be run selectively via `test --suite`.

## Usage

```bash
eac test list-suites
```

This command takes no flags or arguments.

## Examples

```bash
eac test list-suites
```

## See Also

- [test suite](./suite.md) - Run suite
- [show suite](../show/suite.md) - Suite details
- [test Commands](../../categories/test.md)
