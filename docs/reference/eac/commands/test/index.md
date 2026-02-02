# test Commands

Testing and test suite management for BDD specifications and unit tests.

## Commands in this Category

| Command                                       | Purpose                                       |
| --------------------------------------------- | --------------------------------------------- |
| [test](./test.md)                             | Test one or more modules                      |
| [test debug](./debug.md)                      | Parse test results and list failures          |
| [test export-manual](./export-manual.md)      | Export manual test scenarios                  |
| [test import-manual](./import-manual.md)      | Import manual test results                    |
| [test list-suites](./list-suites.md)          | List all available test suites                |
| [test merge-results](./merge-results.md)      | Merge manual results into test manifest       |
| [test suite](./suite.md)                      | Run tests for a specific test suite           |

## Quick Examples

```bash
# Test a module
eac test src-auth

# Run test suite
eac test suite integration

# Debug failures
eac test debug

# Manual testing workflow
eac test export-manual --module src-auth --release v1.2.0
eac test import-manual --input results.json --release v1.2.0
eac test merge-results --module src-auth --version v1.2.0
```

## See Also

- [Category Overview](../categories/test.md)
- [show test-summary](../show/test-summary.md)
