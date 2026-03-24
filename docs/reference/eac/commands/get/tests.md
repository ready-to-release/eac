# get tests

<!-- book:cmd get tests -->

## Output Fields

| Field         | Description                      |
| ------------- | -------------------------------- |
| `total_tests` | Total number of discovered tests |
| `tests`       | List of all tests with metadata  |

Each test entry includes:

- Test name and location (module, package, file)
- Test type (unit, integration, e2e, etc.)
- Tags and markers

## See Also

- [show tests](../show/tests.md) - Formatted table
- [test](../test/test.md)
- [get suite](./suite.md)
