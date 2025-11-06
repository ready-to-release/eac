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

## Available MCP Tools

### Structurizr Architecture Documentation

**What**: MCP server for creating C4 architecture diagrams programmatically.

**When to use**: User asks to create/update architecture documentation for a module.

**Available tools**:
- `create_workspace` - Create architecture workspace for a module
- `add_container` - Add containers to the architecture
- `add_relationship` - Define relationships between containers
- `export_workspace` - Save DSL to `docs/reference/design/<module>/workspace.dsl`

**Example**: "Create architecture documentation for the docs module"

**Setup**: See `src/mcp/structurizr-lite/QUICKSTART.md` for Claude Desktop configuration.

---

## Testing Specifications

This project uses a **three-layer testing approach** unified in Gherkin:

- **ATDD** (Acceptance Criteria as Rule blocks) → Godog
- **BDD** (Behavior Scenarios under Rules) → Godog
- **TDD** (Unit tests) → Go test

**Key Principles**:
1. ATDD and BDD are conceptually distinct layers but technically unified in a single `.feature` file using Gherkin's `Rule:` syntax
2. **Specifications (WHAT) vs Implementation (HOW)**: Specifications live in `specs/`, test implementations live in `src/`

**Full documentation**: [docs/reference/specifications/](docs/reference/specifications/index.md)

### Quick Reference

| Layer    | Purpose                  | Representation         | Tool    | Specification File       | Specification Location      | Implementation File   | Implementation Location |
| -------- | ------------------------ | ---------------------- | ------- | ------------------------ | --------------------------- | --------------------- | ----------------------- |
| **ATDD** | Acceptance Criteria      | `Rule:` blocks         | Godog   | `specification.feature`  | `specs/<module>/<feature>/` | `steps_test.go`       | `src/<module>/tests/`   |
| **BDD**  | Executable Scenarios     | `Scenario:` under Rule | Godog   | `specification.feature`  | `specs/<module>/<feature>/` | `steps_test.go`       | `src/<module>/tests/`   |
| **TDD**  | Unit Tests               | Go test functions      | Go test | N/A                      | N/A                         | `*_test.go`           | `src/<module>/`         |

### Directory Structure

**IMPORTANT**: Specifications and test implementations are deliberately separated:

**Specifications (WHAT to test)** - Located in `specs/`:
```text
specs/<module>/<feature_name>/
├── specification.feature           # Gherkin specs - Rules (ATDD) + Scenarios (BDD)
└── issues.md                        # Optional: questions and blockers
```

**Test Implementations (HOW to test)** - Located in `src/`:
```text
src/<module>/
├── *.go                             # Production code
├── *_test.go                        # TDD unit tests (co-located with code)
└── tests/
    └── steps_test.go                # Godog step definitions for specification.feature
```

**Key Principles**:
- **Specifications live in `specs/`** - Gherkin feature files describe WHAT the system should do
- **Test implementations live in `src/`** - Go test code describes HOW to verify the specifications
- **Separation of concerns** - Business-readable specs separate from technical test code
- **Traceability** - Feature IDs link specifications to their implementations

**Modules**: `cli/`, `vscode/`, `docs/`, `mcp/`

### Creating Feature Specifications

**Step 1**: Create feature directory

```bash
mkdir -p specs/<module>/<feature_name>
```

**Step 2**: Create specification.feature with Rules and Scenarios

**File**: `specs/<module>/<feature_name>/specification.feature`

```gherkin
# Feature ID: <module>_<feature_name>
# Module: <Module>

@<module> @critical @<feature_name>
Feature: [Feature Name]

  As a [role]
  I want [capability]
  So that [business value]

  Rule: [Acceptance Criterion 1]

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

  Rule: [Acceptance Criterion 2]

    @success @ac2
    Scenario: [Happy path for AC2]
      Given [precondition]
      When [action]
      Then [observable outcome]
```

**Step 3**: Implement step definitions (in `src/` directory)

**IMPORTANT**: Step definitions are implemented in `src/`, NOT in `specs/`

**File**: `src/<module>/tests/steps_test.go`

```go
// Feature: <module>_<feature_name>
// Godog step implementations for both ATDD and BDD layers
//
// This file implements steps for the specification at:
// specs/<module>/<feature_name>/specification.feature
package tests

import (
    "context"
    "github.com/cucumber/godog"
)

func InitializeScenario(sc *godog.ScenarioContext) {
    // Register step definitions that implement the scenarios
    // from specs/<module>/<feature_name>/specification.feature
    sc.Step(`^I do something$`, iDoSomething)
}

func iDoSomething() error {
    // Implementation
    return nil
}
```

**Directory Structure After Step 3**:
```
specs/<module>/<feature_name>/
└── specification.feature        # Specifications (WHAT to test)

src/<module>/tests/
└── steps_test.go                # Implementations (HOW to test)
```

### Feature ID Linkage

**Traceability** is maintained through Feature ID across all files:

```text
Feature ID: cli_init-project

Used in:
- specs/cli/init-project/specification.feature: # Feature ID: cli_init-project
- src/cli/tests/steps_test.go: // Feature: cli_init-project
- src/cli/*_test.go: // Feature: cli_init-project (unit tests)
```

**Example**:

```go
// Feature: cli_init-project
// Godog step implementations for specs/cli/init-project/specification.feature
package tests

func iRunCommand(command string) error {
    // implementation
    return nil
}
```

### Decision Tree for AI Assistants

| User Request              | Action                                                                                                                                                                                    |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| "Create a feature spec"   | 1. Create `specs/<module>/<feature>/` directory<br>2. Create `specification.feature` with Feature ID, user story, Rules (ATDD), and Scenarios (BDD)                                     |
| "Implement feature X"     | 1. Read `specs/<module>/<feature>/specification.feature`<br>2. Create `src/<module>/tests/steps_test.go` (Godog steps)<br>3. Write unit tests in `src/<module>/*_test.go`<br>4. Implement feature code<br>5. Run: `go test` |
| "Add acceptance criteria" | Add new `Rule:` block in `specs/<module>/<feature>/specification.feature`                                                                                                                |
| "Add a scenario"          | Add scenario under appropriate `Rule:` block in `specification.feature` with appropriate @ac tag                                                                                         |
| "Validate feature files"  | Check: Feature ID present, Rules define acceptance criteria, @ac tags link scenarios to Rules                                                                                            |
| "Run tests"               | `go test ./src/<module>/tests` (BDD via Godog)<br>`go test ./src/<module>` (TDD unit tests)                                                                                             |

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

1. Create focused `.feature` files (e.g., `format-validation.feature`, `completeness-validation.feature`)
2. Move related scenarios to appropriate files
3. Ensure each file has its own Feature ID and acceptance criteria
4. Delete old monolithic `specification.feature`

**Example**: Module Detection (40 scenarios) split into:

- `automation-module-detection.feature` (8 scenarios)
- `source-module-detection.feature` (8 scenarios)
- `infrastructure-module-detection.feature` (8 scenarios)
- `documentation-module-detection.feature` (8 scenarios)
- `module-detection-edge-cases.feature` (8 scenarios)

**Run split features**:

```bash
# Run all scenarios for a feature from the test implementation directory
cd src/<module>/tests
go test -v

# The test runner will automatically discover all .feature files in the specs directory
```

See [Gherkin Format Guide](docs/reference/specifications/gherkin-format.md) for detailed organization strategies.

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
# specs/cli/user-authentication/specification.feature
@success @ac1 @risk1
Scenario: Valid credentials grant access
  Given I have valid credentials
  When I run "simply login --user admin"
  Then I should be authenticated
```

**See**: [How to Link Risk Controls](docs/how-to-guides/specifications/link-risk-controls.md) for detailed guide.

### Running Tests

**IMPORTANT**:
- Use `go test` (not the deprecated `godog` CLI) for full Go tooling support
- **Tests run from `src/` directory** and read specifications from `specs/` directory
- Test implementations are in `src/`, not in `specs/`

```bash
# Run BDD/ATDD tests (Godog) - recommended approach
# These run the step definitions in src/ against specifications in specs/
cd src/<module>/tests
go test -v

# Run all tests from project root
go test -v ./src/...

# Run specific test function
go test -v -run TestFeatures ./src/<module>/tests

# Run TDD unit tests (also in src/)
go test ./src/<module>

# Run with coverage
go test -cover ./src/<module>/tests

# From project root - run specific module's tests
go test -v ./src/commands/tests    # Runs against specs/src-commands/*/specification.feature
```

**Architecture Reminder**:
```
specs/src-commands/ai-commit-generation/specification.feature  (WHAT to test)
                    ↓
                    ↓ Referenced by
                    ↓
src/commands/tests/steps_test.go                              (HOW to test)
src/commands/tests/godog_test.go                              (Test runner config)
```

### Detailed Guides

- [Testing Approach](docs/explanation/specifications/three-layer-approach.md) - How ATDD/BDD/TDD work together with Rule blocks
- [Gherkin Format](docs/reference/specifications/gherkin-format.md) - Rule blocks, scenarios, tags, and syntax
- [TDD Guide](docs/reference/specifications/tdd-format.md) - Unit tests with Go
