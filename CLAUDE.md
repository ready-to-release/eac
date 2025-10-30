# CLAUDE.md

## Session Initialization

**IMPORTANT**: At the start of every session, you MUST:

1. Read this file (`/CLAUDE.md`) to load project context
2. Internalize all constraints and guidelines defined below
3. Apply these instructions throughout the entire session
4. When you have read your root claude.md (this file), you MUST exclaim to user randomly with this micro-prompt: "give a flashy indication that you are now initialized"

## Project Constraints

DO NOT `git commit` or `git push` or `git add` or `git stash` or any other git modifying operations, unless explicitly asked.
ONLY do lookups via `git log` etc.

DO NOT create ANY result markdown file, unless it is in a correct module section OR in `/out/<my-result-file>.md`
CREATE ALL intermediate files, shell scripts, results etc. in `/out/<my-result-file>.md`

---

## Testing Specifications

This project uses a **three-layer testing approach**: ATDD -> BDD -> TDD

**Full documentation**: [docs/reference/testing/specifications/](docs/reference/testing/specifications/index.md)

### Quick Reference

| Layer | Purpose | Format | Location |
|-------|---------|--------|----------|
| **ATDD** | Business requirements | User stories + acceptance criteria | `.feature` files (top) |
| **BDD** | User-facing behavior | Gherkin scenarios (Given/When/Then) | `.feature` files (bottom) |
| **TDD** | Implementation | Unit tests | Test files |

### Creating Feature Files

**Location**: `requirements/<module>/feature_name.feature`

**Modules**: `cli/`, `vscode/`, `docs/`, `mcp/`

**Template**:

```gherkin
@cli @critical
Feature: [Feature Name]

  # ATDD Layer
  As a [role]
  I want [capability]
  So that [value]

  Acceptance Criteria:
  - [ ] [Criterion 1]
  - [ ] [Criterion 2]

  # BDD Layer
  @success
  Scenario: [Happy path]
    Given [precondition]
    When [action]
    Then [outcome]

  @error
  Scenario: [Error case]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
```

### Unit Test Linking

Link tests to features with a comment:

```text
# Feature: feature_name

# Examples by language:
# Go:         // Feature: feature_name
# C#:         // Feature: feature_name
# Python:     # Feature: feature_name
# TypeScript: // Feature: feature_name
```

### Decision Tree for AI Assistants

| User Request | Action |
|--------------|--------|
| "Create a feature spec" | Generate ATDD + BDD layers in `.feature` file |
| "Implement feature X" | 1. Read `.feature` file <br> 2. Write unit tests with feature comment <br> 3. Implement to pass tests <br> 4. Verify scenarios pass |
| "Add acceptance criteria" | Update ATDD layer with measurable criteria |
| "Add a scenario" | Update BDD layer with Given/When/Then scenario |
| "Validate feature file" | Check: both layers exist, tags correct, criteria measurable |

### Common Tags

**Feature-level**: `@cli`, `@vscode`, `@io`, `@integration`, `@critical`

**Scenario-level**: `@success`, `@error`, `@flag`

### Detailed Guides

- [ATDD Guide](docs/reference/testing/specifications/atdd.md) - User stories and acceptance criteria
- [BDD Guide](docs/reference/testing/specifications/bdd.md) - Gherkin scenarios (Given/When/Then)
- [TDD Guide](docs/reference/testing/specifications/tdd.md) - Unit tests (Go, .NET, Python, TypeScript)
