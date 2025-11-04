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
- **TDD** (unit tests) → Go test

**Full documentation**: [docs/reference/testing/specifications/](docs/reference/testing/specifications/index.md)

### Quick Reference

| Layer | Tool | Format | File | Location |
|-------|------|--------|------|----------|
| **ATDD** | Gauge | Markdown specs | `acceptance.spec` | `requirements/<module>/<feature>/` |
| **BDD** | Godog | Gherkin scenarios | `behavior.feature` | `requirements/<module>/<feature>/` |
| **TDD** | Go test | Unit tests | `*_test.go` | `src/**` |

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

**Example**:

```text
Go: // Feature: cli_init_project
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

### Feature File Size Best Practices

**Scenario Count Guidelines per `.feature` file**:

| Scenario Count | Status | Action |
|----------------|--------|--------|
| **10-15** | ✅ Ideal | Optimal for maintainability |
| **15-20** | ✅ Acceptable | Still manageable |
| **20-30** | ⚠️ Large | Should refactor into multiple files |
| **30+** | ❌ Too Large | Must refactor |

**When to split a feature file**:

- Scenario count exceeds 20
- Multiple distinct sub-features exist
- File has both success and error scenarios that could be separated

**How to split**:

1. Update `acceptance.spec` to list multiple feature files
2. Create focused `.feature` files (e.g., `format_validation.feature`, `completeness_validation.feature`)
3. Update "Validated by" links to point to specific files
4. Delete old monolithic `behavior.feature`

**Example**: Module Detection (40 scenarios) split into:

- `automation_module_detection.feature` (8 scenarios)
- `source_module_detection.feature` (8 scenarios)
- `infrastructure_module_detection.feature` (8 scenarios)
- `documentation_module_detection.feature` (8 scenarios)
- `module_detection_edge_cases.feature` (8 scenarios)

**Run split features**:

```bash
# Run all scenarios for a feature
godog requirements/<module>/<feature>/*.feature

# Run specific sub-feature
godog requirements/<module>/<feature>/sub_feature.feature
```

See [BDD Guide - Best Practices](docs/reference/testing/specifications/bdd.md#best-practices-feature-file-size-and-organization) for detailed splitting strategies.

### Common Tags

**Feature-level**: `@cli`, `@vscode`, `@io`, `@integration`, `@critical`

**Scenario-level**: `@success`, `@error`, `@flag`

**Acceptance Criteria links**: `@ac1`, `@ac2`, `@ac3`, etc.

**Verification tags (for implementation reports)**: `@IV` (Installation Verification), `@PV` (Performance Verification), OV is default when neither tag is present (Operational Verification)

**Risk control tags**: `@risk<ID>` (e.g., `@risk1`, `@risk2`) - Link user scenarios to risk control requirements

#### Risk Control Tags

Link user scenarios to risk control requirements defined in `requirements/risk-controls/`.

**Format**: `@risk<ID>` (e.g., `@risk1`, `@risk2`, `@risk10`)

**How it works**:

1. Risk controls are defined as Gherkin scenarios in `requirements/risk-controls/` that specify what the control requires
2. Each risk control scenario is tagged with `@risk<ID>`
3. User scenarios that implement the control are tagged with the same `@risk<ID>`

**Example**:

Risk control definition:

```gherkin
# requirements/risk-controls/authentication-controls.feature
@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated
  And authentication MUST occur before granting access
```

User scenario implementation:

```gherkin
# requirements/cli/user-authentication/behavior.feature
@success @ac1 @risk1
Scenario: Valid credentials grant access
  Given I have valid credentials
  When I run "simply login --user admin"
  Then I should be authenticated
```

**See**: [How to Link Risk Controls](docs/how-to-guides/testing/link-risk-controls.md) for detailed guide.

### Running Tests

```bash
# Run ATDD acceptance tests (Gauge)
gauge run requirements/

# Run BDD behavior tests (Godog)
godog requirements/**/behavior.feature

# Run BDD tests by verification type (for implementation reports)
godog --tags="@IV" requirements/**/behavior.feature       # Installation Verification
godog --tags="@PV" requirements/**/behavior.feature       # Performance Verification
godog --tags="~@IV && ~@PV" requirements/**/behavior.feature  # Operational Verification

# Run TDD unit tests
go test ./...

# Run specific feature
gauge run requirements/cli/init_project/
godog requirements/cli/init_project/behavior.feature
```

### Detailed Guides

- [ATDD Guide](docs/reference/testing/specifications/atdd.md) - Gauge specs, Example Mapping, acceptance criteria
- [BDD Guide](docs/reference/testing/specifications/bdd.md) - Godog scenarios, Gherkin syntax, step definitions
- [TDD Guide](docs/reference/testing/specifications/tdd.md) - Unit tests with Go
