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

This project uses a **three-layer testing approach** with **separate files for each layer**:

- **ATDD** (acceptance.spec) → Gauge
- **BDD** (behavior.feature) → Godog
- **TDD** (unit tests) → Go test, pytest

**Full documentation**: [docs/reference/testing/specifications/](docs/reference/testing/specifications/index.md)

### Quick Reference

| Layer | Tool | Format | File | Location |
|-------|------|--------|------|----------|
| **ATDD** | Gauge | Markdown specs | `acceptance.spec` | `requirements/<module>/<feature>/` |
| **BDD** | Godog | Gherkin scenarios | `behavior.feature` | `requirements/<module>/<feature>/` |
| **TDD** | Go test, pytest | Unit tests | `*_test.go`, `test_*.py` | `src/**` or `tests/**` |

### Directory Structure

```text
requirements/<module>/<feature_name>/
├── acceptance.spec                 # Gauge (ATDD) - business requirements
├── behavior.feature                # Godog (BDD) - executable scenarios
├── acceptance_test.go              # Gauge step implementations
├── step_definitions_test.go        # Godog step definitions
└── issues.md                       # Optional: questions and blockers
```

**Modules**: `cli/`, `vscode/`, `docs/`, `mcp/`

### Creating Feature Specifications

**Step 1**: Create feature directory
```bash
mkdir -p requirements/<module>/<feature_name>
```

**Step 2**: Create acceptance.spec (Gauge ATDD)

**File**: `requirements/<module>/<feature_name>/acceptance.spec`

```markdown
# [Feature Name]

> **Feature ID**: <module>_<feature_name>
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: <Module>
> **Tags**: <tags>

## User Story

* As a [role]
* I want [capability]
* So that [business value]

## Acceptance Criteria

* [Measurable criterion 1]
* [Measurable criterion 2]

## Acceptance Tests

### AC1: [Criterion 1]
**Validated by**: behavior.feature -> @ac1 scenarios

* [Gauge step 1]
* [Gauge step 2]
* [Verification step]
```

**Step 3**: Create behavior.feature (Godog BDD)

**File**: `requirements/<module>/<feature_name>/behavior.feature`

```gherkin
# Feature ID: <module>_<feature_name>
# Acceptance Spec: acceptance.spec
# Module: <Module>

@<module> @critical @<feature_name>
Feature: [Feature Name]

  @success @ac1
  Scenario: [Happy path]
    Given [precondition]
    When [action]
    Then [observable outcome]
    And [verification]

  @error @ac1
  Scenario: [Error case]
    Given [precondition]
    When [invalid action]
    Then [error behavior]
    And [error verification]
```

### Feature ID Linkage

**Traceability** is maintained through Feature ID across all files:

```text
Feature ID: cli_init_project

Used in:
- acceptance.spec: > **Feature ID**: cli_init_project
- behavior.feature: # Feature ID: cli_init_project
- acceptance_test.go: // Feature: cli_init_project
- step_definitions_test.go: // Feature: cli_init_project
- Unit tests: // Feature: cli_init_project
```

**Example by language**:
```
Go:     // Feature: cli_init_project
Python: # Feature: cli_init_project
```

### Decision Tree for AI Assistants

| User Request | Action |
|--------------|--------|
| "Create a feature spec" | 1. Create feature directory<br>2. Create `acceptance.spec` (Gauge)<br>3. Create `behavior.feature` (Godog)<br>4. Add Feature ID to both files |
| "Implement feature X" | 1. Read `acceptance.spec` and `behavior.feature`<br>2. Create `acceptance_test.go` (Gauge steps)<br>3. Create `step_definitions_test.go` (Godog steps)<br>4. Write unit tests with Feature ID comment<br>5. Implement feature code<br>6. Run: `gauge run` and `godog run` |
| "Add acceptance criteria" | Update ATDD layer in `acceptance.spec` only |
| "Add a scenario" | Update BDD layer in `behavior.feature` only |
| "Validate feature files" | Check: both files exist, Feature ID matches, @ac tags link scenarios to criteria |
| "Run tests" | `gauge run requirements/` (ATDD)<br>`godog requirements/**/behavior.feature` (BDD)<br>`go test ./...` (TDD) |

### Common Tags

**Feature-level**: `@cli`, `@vscode`, `@io`, `@integration`, `@critical`

**Scenario-level**: `@success`, `@error`, `@flag`

**Acceptance Criteria links**: `@ac1`, `@ac2`, `@ac3`, etc.

### Running Tests

```bash
# Run ATDD acceptance tests (Gauge)
gauge run requirements/

# Run BDD behavior tests (Godog)
godog requirements/**/behavior.feature

# Run TDD unit tests
go test ./...

# Run specific feature
gauge run requirements/cli/init_project/
godog requirements/cli/init_project/behavior.feature
```

### Detailed Guides

- [ATDD Guide](docs/reference/testing/specifications/atdd.md) - Gauge specs, Example Mapping, acceptance criteria
- [BDD Guide](docs/reference/testing/specifications/bdd.md) - Godog scenarios, Gherkin syntax, step definitions
- [TDD Guide](docs/reference/testing/specifications/tdd.md) - Unit tests (Go and Python)
