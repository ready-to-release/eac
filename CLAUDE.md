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

This project uses a **two-layer testing approach**:

- **BDD** (behavior.feature) → Godog
- **TDD** (unit tests) → Go test

**Full documentation**: [docs/reference/testing/specifications/](docs/reference/testing/specifications/index.md)

### Quick Reference

| Layer    | Tool    | Format            | File               | Location                           |
| -------- | ------- | ----------------- | ------------------ | ---------------------------------- |
| **BDD**  | Godog   | Gherkin scenarios | `behavior.feature` | `specs/<module>/<feature>/`        |
| **TDD**  | Go test | Unit tests        | `*_test.go`        | `src/**`                           |

### Directory Structure

**Specifications** (in `specs/`) and **Test Implementations** (in `src/`) are **separate**:

```text
specs/<module>/<feature_name>/
├── behavior.feature                # Godog (BDD) - executable scenarios
└── issues.md                       # Optional: questions and blockers

src/<module>/
├── *.go                            # Production code
├── *_test.go                       # TDD unit tests (co-located with code)
└── tests/
    └── steps_test.go               # Godog step definitions (BDD)
```

**Key Principles**:
- **Specifications live in `specs/`** - Gherkin feature files (WHAT to test)
- **Test implementations live in `src/`** - Go test code (HOW to test)
- **Unit tests co-locate with production code** - `*_test.go` files alongside the code they test
- **BDD step definitions in `tests/` subdirectory** - Godog implementations

**Modules**: `cli/`, `vscode/`, `docs/`, `mcp/`

### Creating Feature Specifications

**Step 1**: Create feature directory

```bash
mkdir -p specs/<module>/<feature_name>
```

**Step 2**: Create behavior.feature (Godog BDD) with acceptance criteria

**File**: `specs/<module>/<feature_name>/behavior.feature`

```gherkin
# Feature ID: <module>_<feature_name>
# Module: <Module>

@<module> @critical @<feature_name>
Feature: [Feature Name]

  As a [role]
  I want [capability]
  So that [business value]

  Rule: [Measurable criterion 1]

    @success @ac1
    Scenario: [Happy path for AC1]
      Given [precondition]
      When [action]
      Then [observable outcome]
      And [verification]

    @error @ac1
    Scenario: [Error case for AC1]
      Given [precondition]
      When [invalid action]
      Then [error behavior]
      And [error verification]

  Rule: [Measurable criterion 2]

    @success @ac2
    Scenario: [Happy path for AC2]
      Given [precondition]
      When [action]
      Then [observable outcome]
```

### Feature ID Linkage

**Traceability** is maintained through Feature ID across all files:

```text
Feature ID: cli_init_project

Used in:
- specs/cli/init_project/behavior.feature: # Feature ID: cli_init_project
- src/cli/tests/steps_test.go: // Feature: cli_init_project
- src/cli/*_test.go: // Feature: cli_init_project (unit tests)
```

**Example**:

```go
// Feature: cli_init_project
// Godog step implementations
package tests

func iRunCommand(command string) error {
    // implementation
    return nil
}
```

### Decision Tree for AI Assistants

| User Request              | Action                                                                                                                                                                                                                    |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "Create a feature spec"   | 1. Create `specs/<module>/<feature>/` directory<br>2. Create `behavior.feature` with Feature ID, user story, and acceptance criteria                                                                                    |
| "Implement feature X"     | 1. Read `specs/<module>/<feature>/behavior.feature`<br>2. Create `src/<module>/tests/steps_test.go` (Godog steps)<br>3. Write unit tests in `src/<module>/*_test.go`<br>4. Implement feature code<br>5. Run: `godog` |
| "Add acceptance criteria" | Update acceptance criteria section in `specs/<module>/<feature>/behavior.feature`                                                                                                                                        |
| "Add a scenario"          | Add scenario to `specs/<module>/<feature>/behavior.feature` with appropriate @ac tag                                                                                                                                     |
| "Validate feature files"  | Check: Feature ID present, acceptance criteria documented, @ac tags link scenarios to criteria                                                                                                                           |
| "Run tests"               | `godog specs/**/behavior.feature` (BDD)<br>`go test ./src/...` (TDD)                                                                                                                                                    |

### Feature File Size Best Practices

**Scenario Count Guidelines per `.feature` file**:

| Scenario Count | Status        | Action                              |
| -------------- | ------------- | ----------------------------------- |
| **10-15**      | ✅ Ideal      | Optimal for maintainability         |
| **15-20**      | ✅ Acceptable | Still manageable                    |
| **20-30**      | ⚠️ Large      | Should refactor into multiple files |
| **30+**        | ❌ Too Large  | Must refactor                       |

**When to split a feature file**:

- Scenario count exceeds 20
- Multiple distinct sub-features exist
- File has both success and error scenarios that could be separated

**How to split**:

1. Create focused `.feature` files (e.g., `format_validation.feature`, `completeness_validation.feature`)
2. Move related scenarios to appropriate files
3. Ensure each file has its own Feature ID and acceptance criteria
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
godog specs/<module>/<feature>/*.feature

# Run specific sub-feature
godog specs/<module>/<feature>/sub_feature.feature
```

See [BDD Guide - Best Practices](docs/reference/testing/specifications/bdd.md#best-practices-feature-file-size-and-organization) for detailed splitting strategies.

### Common Tags

**Feature-level**: `@cli`, `@vscode`, `@io`, `@integration`, `@critical`

**Scenario-level**: `@success`, `@error`, `@flag`

**Acceptance Criteria links**: `@ac1`, `@ac2`, `@ac3`, etc.

**Verification tags (for implementation reports)**: `@IV` (Installation Verification), `@PV` (Performance Verification), OV is default when neither tag is present (Operational Verification)

**Risk control tags**: `@risk<ID>` (e.g., `@risk1`, `@risk2`) - Link user scenarios to risk control requirements

#### Risk Control Tags

Link user scenarios to risk control requirements defined in `specs/risk-controls/`.

**Format**: `@risk<ID>` (e.g., `@risk1`, `@risk2`, `@risk10`)

**How it works**:

1. Risk controls are defined as Gherkin scenarios in `specs/risk-controls/` that specify what the control requires
2. Each risk control scenario is tagged with `@risk<ID>`
3. User scenarios that implement the control are tagged with the same `@risk<ID>`

**Example**:

Risk control definition:

```gherkin
# specs/risk-controls/authentication-controls.feature
@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated
  And authentication MUST occur before granting access
```

User scenario implementation:

```gherkin
# specs/cli/user-authentication/behavior.feature
@success @ac1 @risk1
Scenario: Valid credentials grant access
  Given I have valid credentials
  When I run "simply login --user admin"
  Then I should be authenticated
```

**See**: [How to Link Risk Controls](docs/how-to-guides/testing/link-risk-controls.md) for detailed guide.

### Running Tests

**IMPORTANT**: Use `go test` (not the deprecated `godog` CLI) for full Go tooling support.

```bash
# Run BDD behavior tests (Godog) - recommended approach
cd src/<module>/tests
go test -v

# Run all tests from project root
go test -v ./src/...

# Run with test coverage
go test -cover ./src/commands/tests

# Run with race detector
go test -race ./src/commands/tests

# Run specific test function
go test -v -run TestFeatures ./src/commands/tests

# Run TDD unit tests
go test ./src/<module>

# Common combined command
cd src/commands/tests && go test -v
```

**Advanced options** (passed via test flags):
```bash
# Run tests matching tag (requires custom filtering in code)
# Tags are not directly supported via go test flags
# Use scenario filtering in your godog.Options instead
```

### Detailed Guides

- [BDD Guide](docs/reference/testing/specifications/bdd.md) - Godog scenarios, Gherkin syntax, step definitions, acceptance criteria
- [TDD Guide](docs/reference/testing/specifications/tdd.md) - Unit tests with Go
