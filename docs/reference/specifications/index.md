# Specifications Reference

Quick reference for specification formats, commands, and syntax.

---

## Architecture Overview

This project maintains clear separation between specifications and implementations:

| Concern        | Location   | Contents                      |
| -------------- | ---------- | ----------------------------- |
| Specifications | `specs/`   | Gherkin `.feature` files      |
| Implementations| `src/`     | Go test code, step definitions|

See [Gherkin Format](./gherkin-format.md) for specification syntax.
See [Godog Commands](./godog-commands.md) for running tests from `src/`.

---

## Specification Format References

### [Gherkin Format](gherkin-format.md)

specification.feature file structure with Rule blocks for ATDD and Scenario blocks for BDD.

**Quick lookup:**

- Template structure with Rules and Scenarios
- Feature metadata and tags
- Feature naming convention (kebab-case)
- Rule blocks (ATDD - acceptance criteria)
- Scenario blocks (BDD - executable examples)
- Given/When/Then syntax
- Tag system (@ac1, @success, @error, @IV, @PV, @risk)
- Background sections
- Feature name traceability
- Complete examples

### [TDD Format](tdd-format.md)

Unit test structure and Go test conventions.

**Quick lookup:**

- Test file patterns and naming
- Arrange-Act-Assert pattern
- Table-driven tests
- Error handling and assertions
- Feature name traceability
- Running tests and coverage

---

## Command References

### [Godog Commands](godog-commands.md)

Running Godog tests with go test (executes both ATDD and BDD layers).

---

## Scenario Classification

### [Verification Tags](verification-tags.md)

Scenario classification for implementation reports (IV/OV/PV).

---

## Related Documents

- **Understanding-oriented?** See [Specifications Explanation](../../explanation/specifications/index.md)
- **Need to perform tasks?** See [Specifications How-to Guides](../../how-to-guides/specifications/index.md)
