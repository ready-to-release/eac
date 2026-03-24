# show suite

<!-- book:cmd show suite -->

## Output Sections

1. **Suite Header** - Name, moniker, description, production test count, framework test count (excluded from display), and total discovered count.
2. **Selection Criteria** - For each selector: AnyOf tags, RequireAll tags, and Exclude tags.
3. **Production Tests** - Table with columns: #, Moniker, Test Name, Type, Module, Level, Verification, System Deps.
4. **Statistics** - Counts by Type (e.g., gherkin, gotest), counts by Module, and aggregated system and module dependencies.

Validation warnings are printed to stderr if any tests have issues. Framework tests are excluded from the display.

## See Also

- [get suite](../get/suite.md) - JSON output
- [test suite](../test/suite.md) - Run suite
- [test list-suites](../test/list-suites.md)
