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

- **Module names**: Use kebab-case (e.g., `cli`, `vscode`, `src-commands`, `vscode-extension`)
- **Directory names**: Use kebab-case matching feature names (e.g., `init-project`, `design-command`)
- **Feature names**: Use kebab-case format `[module-name_feature-name]` (e.g., `cli_init-project`, `src-commands_design-command`)

---

## Step 2: Create specification.feature with Rules and Scenarios

**Important**: The canonical template at `templates/specs/specification.feature` includes:

- **Architectural notes** (lines 1-9) - Explains specs/ vs src/ separation
- **Instructions** (lines 11-17) - Step-by-step guide inline with template
- **Complete examples** - All tag types and scenarios

**Copy the template file directly to get the full architectural context.** Documentation examples shown here omit the architectural notes and instructions to keep them concise.

---

### Determine Feature ID

**Format**: `[module-name_feature-name]` (kebab-case)

**Naming Convention**:

- **Module**: Use kebab-case (e.g., `src-commands`, `vscode-extension`, `cli`)
- **Feature**: Use kebab-case (e.g., `design-command`, `commit-workflow`, `init-project`)
- **Separator**: Single underscore `_` between module and feature

**Examples**:

- ✅ `src-commands_design-command` (kebab-case module and feature)
- ✅ `vscode-extension_commit-workflow` (kebab-case module and feature)
- ✅ `cli_init-project` (single-word module, kebab-case feature)
- ✅ `mcp-server_github-integration` (kebab-case both)

**Incorrect Examples** (avoid):

- ❌ `src_commands_design_command` (no kebab-case)
- ❌ `srcCommands_designCommand` (camelCase)
- ❌ `SrcCommands_DesignCommand` (PascalCase)

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
@module @critical
Feature: [module-name_feature-name]

  As a [role from Yellow Card]
  I want [capability from Yellow Card]
  So that [business value from Yellow Card]

  Background:
    Given [common precondition if needed]

  Rule: [Blue Card 1 - Acceptance Criterion]

    @success @ac1 @IV
    Scenario: [Green Card 1a - Installation verification]
      Given [precondition]
      When [setup action]
      Then [installation verified]
      And [configuration verified]

    @success @ac1
    Scenario: [Green Card 1b - Happy path example]
      Given [precondition]
      When [action]
      Then [observable outcome]
      And [verification]

    @error @ac1
    Scenario: [Green Card 1c - Error case example]
      Given [error precondition]
      When [invalid action]
      Then [error behavior]
      And [error message]

  Rule: [Blue Card 2 - Acceptance Criterion]

    @success @ac2 @PV
    Scenario: [Green Card 2a - Performance verification]
      Given [performance precondition]
      When [action under load]
      Then [outcome within SLA]
      And [resource usage within limits]

    @success @ac2 @risk1
    Scenario: [Green Card 2b - Risk control example]
      Given [security precondition]
      When [authenticated action]
      Then [access granted]
      And [audit logged]
```

**Naming Convention**:

The feature name MUST follow kebab-case format: `[module-name_feature-name]`

- **Module name**: kebab-case (e.g., `src-commands`, `vscode-extension`)
- **Feature name**: kebab-case (e.g., `design-command`, `init-project`)
- **Separator**: Single underscore `_`

**Examples**:

- `src-commands_design-command` ✅
- `vscode-extension_commit-workflow` ✅
- `cli_init-project` ✅

**Tags**:

- `@module` - Required, identifies owning module
- `@critical` - Required, indicates critical functionality
- `@<feature-name>` - Optional, add if feature has multiple specification files

**Verification Tags**:

- `@IV` - Installation Verification (deployment, setup, configuration tests)
- `@OV` - Operational Verification (default if no @IV/@PV, standard functional tests)
- `@PV` - Performance Verification (load tests, SLA compliance)

**Risk Control Tags**:

- `@risk<N>` - Links scenario to risk control requirement (see [Link Risk Controls](./link-risk-controls.md))

**Template Source**:

The canonical template with architectural notes and step-by-step instructions is at:

- **File**: `templates/specs/specification.feature`
- **Usage**: Copy this template when creating new specifications
- **Note**: The template includes architectural notes (explaining specs/ vs src/ separation) and inline instructions. These are helpful in the template file but are NOT shown in documentation examples to keep them concise.

**See also**:

- [Link Risk Controls](./link-risk-controls.md) - Risk control tagging guide

---

## Complete Example

**File**: `specs/cli/init-project/specification.feature`

```gherkin
@cli @critical
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

**Note**: This example uses `cli_init-project` - a single-word module (`cli`) with kebab-case feature (`init-project`). For multi-word modules, use kebab-case for both parts (e.g., `src-commands_design-command`).

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

## Related Documentation

### Understanding the Approach

- [Ubiquitous Language](../../explanation/specifications/ubiquitous-language.md) - Domain vocabulary foundation and complete workflow
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md) - How ATDD/BDD/TDD work together
- [ATDD and BDD with Gherkin](../../explanation/specifications/atdd-bdd-with-gherkin.md) - Detailed explanation

### Workshop Guides

- [Run Event Storming Workshop](./run-event-storming.md) - Discover domain vocabulary
- [Run Example Mapping Workshop](./run-example-mapping.md) - Apply vocabulary to features
