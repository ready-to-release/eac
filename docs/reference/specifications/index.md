<!-- EDITOR
# Editor: reference/specifications/index.md

## Soul
Technical reference for Gherkin specification standards, TDD workflows, tag conventions, and quality checklists.

## Sections
1. Overview
2. In This Section
3. Diataxis
-->

# Specification Reference

This section provides technical reference documentation for writing and organizing Gherkin specifications. Find TDD workflows, naming conventions, tag references, quality standards, and structural patterns.

## In This Section

### Implementation Guides

Language-specific implementation details for BDD specifications and testing.

| Language | Reference | Description |
| -------- | --------- | ----------- |
| **Go** | [Go Implementation Guide](./go-implementation-guide.md) | Go/Godog implementation: file organization, build tags, step definitions, test execution |
| TypeScript | _Planned_ | TypeScript/Cucumber-js implementation guide |
| Python | _Planned_ | Python/behave or pytest-bdd implementation guide |
| Java | _Planned_ | Java/Cucumber implementation guide |

**Note**: Explanation documents are technology-agnostic. Consult your language implementation guide for syntax, file organization, and execution details.

### Standards and Conventions

| Reference                                                               | Description                                           |
| ----------------------------------------------------------------------- | ----------------------------------------------------- |
| [Canon TDD Workflow](./canon-tdd-workflow.md)                           | Kent Beck's five-step test-driven development process |
| [Example Mapping Cards](./example-mapping-cards.md)                     | Collaborative specification discovery technique       |
| [Feature Naming](./feature-naming.md)                                   | Naming conventions for feature files                  |
| [Gherkin Limits](./gherkin-limits.md)                                   | Structural constraints and best practices             |
| [Spec Quality Checklist](./spec-quality-checklist.md)                   | Quality criteria for specifications                   |
| [Tag Reference](./tag-reference.md)                                     | Standard tags for organizing and categorizing specs   |
| [Three-Layered Testing Structure](./three-layered-testing-structure.md) | Integration, functional, and system test organization |

{{ diataxis_footer() }}