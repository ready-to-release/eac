# validate specs

<!-- book:cmd validate specs -->

## What It Checks

- **Structure** -- Proper Feature/Rule/Scenario hierarchy.
- **Tags** -- Tags conform to project conventions (unless `--no-check-tags`).
- **Step formatting** -- Given/When/Then steps follow project standards.
- **Content quality** -- Scenarios have meaningful descriptions and steps.

## Common Errors

- **File or directory not found** -- The specified path does not exist.
- **Structure error** -- Invalid Gherkin hierarchy (e.g., Scenario outside Feature).
- **Tag validation error** -- A tag does not conform to project conventions.

## See Also

- [show specs](../show/specs.md)
- [validate](./validate.md)
