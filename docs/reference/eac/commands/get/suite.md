# get suite

<!-- book:cmd get suite -->

## Output

Returns a suite report containing:

- Suite metadata (moniker, name, description)
- Test selection criteria (tags, modules, patterns)
- Suite-level configuration and settings
- List of included tests with their metadata and module associations

If an invalid suite moniker is provided, the command lists available suites on stderr.

## See Also

- [show suite](../show/suite.md) - Formatted display
- [test suite](../test/suite.md) - Run suite
- [test list-suites](../test/list-suites.md)
