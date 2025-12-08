<!-- EDITOR
# Editor: reference/specifications/gherkin-limits.md

## Soul

Size and complexity limits for feature files with action thresholds for splitting.

## Sections

1. Rule Count Per Feature
2. Scenario Count Per Feature
3. Scenarios Per Rule
4. Related
-->

# Gherkin Limits Reference

Guidelines for feature file size and complexity.

## Rule Count Per Feature

| Rule Count | Assessment    | Action             |
| ---------- | ------------- | ------------------ |
| **2-4**    | ✅ Ideal      | Good feature scope |
| **5-6**    | ✅ Acceptable | Monitor complexity |
| **7-10**   | ⚠️ Large      | Consider splitting |
| **>10**    | ❌ Too large  | Must split feature |

### Why Limit Rules?

- **Cognitive load**: More than 6 acceptance criteria becomes hard to reason about
- **Feature scope**: Large Rule count suggests feature does too much
- **Maintainability**: Smaller features are easier to review and update
- **Testing**: Fewer Rules = faster, more focused test execution

## Scenario Count Per Feature

| Scenario Count | Assessment    | Action              |
| -------------- | ------------- | ------------------- |
| **10-15**      | ✅ Ideal      | Optimal readability |
| **15-20**      | ✅ Acceptable | Still manageable    |
| **20-30**      | ⚠️ Large      | Should split        |
| **>30**        | ❌ Too large  | Must split          |

### Why Limit Scenarios?

- **Readability**: Files over 20 scenarios become hard to navigate
- **Execution time**: Large files take longer to run
- **Merge conflicts**: Multiple developers editing same large file
- **Focus**: Many scenarios suggest feature does too much

## Scenarios Per Rule

**Target**: 2-4 scenarios per Rule

Each Rule should have:

1. **Happy path** - Primary use case
2. **Error cases** - How system handles failures
3. **Edge cases** - Boundary conditions, special states

## Related

- [Feature Naming](feature-naming.md)
- [Gherkin Concepts](../../explanation/specifications/gherkin-concepts.md)
