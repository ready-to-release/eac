# Show Tests

<!-- book:cmd show tests -->

Displays all test assertions in the repository with their metadata, module overview statistics, and summary breakdowns.

## Usage

```bash
eac show tests
```

## Output Sections

1. **Header** - Total assertion count.
2. **Module Overview** - Summary table with one row per module, columns: Module, Assertions, Components, L0, L1, L2, L3, L4, Types. Includes a totals row and OS-filtered count if applicable.
3. **All Assertions** - Detailed table with columns: #, Module, Component, Assertion, Type, Level, Verify, Deps. Inferred tags are marked with `*`, inferred dependencies with `~`.
4. **Summary by Type** - Assertion counts per test type (e.g., gherkin, gotest).
5. **Summary by Level** - Assertion counts per level (@L0 through @L4).

## Examples

```bash
# Show all tests in the repository
eac show tests
```

## See Also

- [get tests](../get/tests.md) - JSON output
- [test](../test/test.md) - Run tests
- [show suite](./suite.md) - Test suite details
