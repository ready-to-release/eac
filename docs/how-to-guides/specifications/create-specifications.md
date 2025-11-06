# Create Specifications

Create specification.feature file with Rules and Scenarios for a new feature.

---

## Overview

This guide covers creating **specifications** (Gherkin files) in the `specs/` directory. For implementing step definitions (test code), see [Setup Godog](./setup-godog.md).

---

## Prerequisites

- [Godog installed and configured](./setup-godog.md)
- Domain vocabulary established through [Event Storming](./run-event-storming.md)
- [Example Mapping workshop](./run-example-mapping.md) completed for the feature
- Workshop cards (Yellow, Blue, Green, Red) available
- Glossary from Event Storming with domain terms

**See**: [Ubiquitous Language](../../explanation/specifications/ubiquitous-language.md) for understanding domain vocabulary foundation.

---

## Workflow: From Domain Discovery to Specifications

This guide implements the complete flow from domain discovery to executable specifications:

```text
Event Storming → Example Mapping → Gherkin Specification
     (vocabulary)      (features)       (executable specs)
```

**Steps**:

1. **Establish Ubiquitous Language** - Event Storming discovers domain vocabulary
2. **Discover Requirements** - Example Mapping applies vocabulary to specific features
3. **Write Specifications** - Create `specification.feature` using discovered terms

This guide focuses on **Step 3**: Creating a single `specification.feature` file that contains:

- **ATDD Layer**: `Rule:` blocks (acceptance criteria from Blue Cards)
- **BDD Layer**: `Scenario:` blocks nested under Rules (executable examples from Green Cards)
- **Ubiquitous Language**: All scenarios use domain terms from Event Storming

**See**: [Ubiquitous Language - Complete Flow](../../explanation/specifications/ubiquitous-language.md#the-complete-flow) for the full workflow.

---

## Step 1: Create Feature Directory in specs/

**Location**: `specs/<module>/<feature>/`

Determine the feature module and name, then create the directory:

```bash
mkdir -p specs/<module>/<feature>/
```

**Example**:

```bash
mkdir -p specs/cli/init-project
cd specs/cli/init-project
```

**Naming Conventions**:

- **Module names**: `cli`, `vscode`, `docs`, `mcp`, `src-commands`
- **Directory names**: Use dashes (e.g., `init-project`, `ai-commit-generation`)
- **Feature IDs**: Use underscores between module and feature, dashes within (e.g., `cli_init-project`)

---

## Step 2: Create specification.feature with Rules and Scenarios

### Determine Feature ID

**Format**: `<module>_<feature-name>`

**Examples**:

- `cli_init-project`
- `src-commands_ai-commit-generation`

### Create File

Create `specs/<module>/<feature>/specification.feature`:

```bash
touch specification.feature
```

### Write File Structure

Use your Example Mapping cards:

- **Yellow Card** → Feature description (user story)
- **Blue Cards** → Rule blocks (ATDD - acceptance criteria)
- **Green Cards** → Scenario blocks under Rules (BDD - executable examples)

**Template**:

```gherkin
@<module> @critical @<feature-name>
Feature: <module>_<feature-name>

  As a [role from Yellow Card]
  I want [capability from Yellow Card]
  So that [business value from Yellow Card]

  Background:
    Given [common precondition if needed]

  Rule: [Blue Card 1 - Acceptance Criterion]

    @success @ac1
    Scenario: [Green Card 1a - Happy path example]
      Given [precondition]
      When [action]
      Then [observable outcome]
      And [verification]

    @error @ac1
    Scenario: [Green Card 1b - Error case example]
      Given [error precondition]
      When [invalid action]
      Then [error behavior]
      And [error message]

  Rule: [Blue Card 2 - Acceptance Criterion]

    @success @ac2
    Scenario: [Green Card 2a - Example]
      Given [precondition]
      When [action]
      Then [outcome]
```

---

## Complete Example

**File**: `specs/cli/init-project/specification.feature`

```gherkin
@cli @critical @init
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project with a single command
  So that I can quickly start development with proper structure

  Rule: Creates project directory structure

    @success @ac1
    Scenario: Initialize in empty directory creates structure
      Given I am in an empty folder
      When I run "r2r init my-project"
      Then a directory named "my-project/src/" should exist
      And a directory named "my-project/tests/" should exist
      And a directory named "my-project/docs/" should exist

    @error @ac1
    Scenario: Initialize in existing project shows error
      Given I am in a directory with "r2r.yaml"
      When I run "r2r init"
      Then the command should fail
      And stderr should contain "already initialized"

  Rule: Generates valid configuration file

    @success @ac2
    Scenario: Generated YAML has default values
      Given I am in an empty folder
      When I run "r2r init my-project"
      Then a file named "my-project/r2r.yaml" should be created
      And the file should contain valid YAML
      And the YAML should have key "project.name"
```

---

## Step 3: Implement Step Definitions

After creating `specification.feature` in `specs/<module>/<feature>/`, implement the step definitions.

Step definitions live in `src/<module>/tests/` (separate from specifications in `specs/`).

**File**: `src/<module>/tests/steps_test.go`

```go
// Feature: <module>_<feature-name>
package tests

import (
    "context"
    "testing"
    "github.com/cucumber/godog"
)

func iAmInAnEmptyFolder(ctx context.Context) (context.Context, error) {
    // Implementation
    return ctx, nil
}

func iRun(ctx context.Context, command string) (context.Context, error) {
    // Implementation
    return ctx, nil
}

func aFileNamedShouldBeCreated(ctx context.Context, filename string) error {
    // Implementation
    return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^a file named "([^"]*)" should be created$`, aFileNamedShouldBeCreated)
}

func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../../specs/cli/init-project/specification.feature"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}
```

**See**: [Setup Godog](./setup-godog.md) for detailed step definition patterns.

---

## Step 4: Run Tests

### Run from src/ tests directory

```bash
cd src/cli/tests
go test -v
```

### Run specific feature

```bash
godog ../../../specs/cli/init-project/specification.feature
```

### Run with tags

```bash
# Run only success scenarios
godog --tags="@success" ../../../specs/cli/init-project/specification.feature

# Run specific acceptance criterion
godog --tags="@ac1" ../../../specs/cli/init-project/specification.feature
```

**See**: [Run Tests](./run-tests.md) for comprehensive testing guide.

---

## Step 5: Track Questions (Optional)

If you had Red Cards from Example Mapping, create an issues tracker:

**File**: `specs/cli/init_project/issues.md`

```markdown
# Open Questions

## RED-1: What if r2r.yaml already exists?

**Status**: Open
**Raised**: 2025-11-03
**Decision needed by**: Product Owner
**Resolution**: TBD

## RED-2: Should we support a --force flag?

**Status**: Open
**Raised**: 2025-11-03
**Decision needed by**: Development Team
**Resolution**: TBD
```

---

## Checklist

Before considering the feature spec complete:

- [ ] Feature directory created in `specs/<module>/<feature>/`
- [ ] `specification.feature` created with Feature, Rules, and Scenarios
- [ ] Feature uses domain terms from Event Storming glossary
- [ ] Yellow Card → Feature description (user story)
- [ ] Blue Cards → Rule blocks (acceptance criteria)
- [ ] Green Cards → Scenario blocks under Rules
- [ ] @ac tags link Scenarios to Rules
- [ ] Step definitions implemented in `src/<module>/tests/steps_test.go`
- [ ] Tests run successfully from `src/<module>/tests/`
- [ ] Red Cards tracked in `issues.md` (if any)

---

## Next Steps

- ✅ Feature specification created using Ubiquitous Language
- **Next**: Implement the feature using [TDD](../../reference/specifications/tdd-format.md)
- **Then**: [Run Tests](./run-tests.md) to validate implementation

---

## Related Documentation

### Understanding the Approach

- [Ubiquitous Language](../../explanation/specifications/ubiquitous-language.md) - Domain vocabulary foundation and complete workflow
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - How ATDD/BDD/TDD work together
- [ATDD and BDD with Gherkin](../../explanation/specifications/atdd-bdd-with-gherkin.md) - Detailed explanation

### Workshop Guides

- [Run Event Storming Workshop](./run-event-storming.md) - Discover domain vocabulary
- [Run Example Mapping Workshop](./run-example-mapping.md) - Apply vocabulary to features

### Reference

- [Gherkin Format](../../reference/specifications/gherkin-format.md) - Specification syntax reference
- [TDD Format](../../reference/specifications/tdd-format.md) - Unit testing guide
