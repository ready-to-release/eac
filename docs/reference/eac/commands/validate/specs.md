# Validate specs

<!-- book:cmd validate specs -->

## How It Works

Validates Gherkin specifications against quality contracts:

- **Gherkin Syntax**: Validates feature file structure (Feature, Scenario, Given/When/Then)
- **Step Definitions**: Verifies all steps have matching Go step implementations
- **Quality Standards**: Checks specification follows best practices and naming conventions
- **Tag Validation**: Ensures all tags are defined in tag contract

Validation prevents untestable or malformed specifications from being committed.

## See Also

- [validate](./validate.md)
- [create spec](../create/spec.md)
- [get specs unused-steps](../get/specs-unused-steps.md)
- [validate Commands](../categories/validate.md)
