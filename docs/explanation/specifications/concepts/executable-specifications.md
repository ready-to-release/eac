# Executable Specifications
>
> **From discovery to implementation**

Learn how to move from collaborative discovery to running, automated specifications.

---

## From Discovery to Implementation

The complete workflow proceeds through these phases:

1. **Discovery**: Example Mapping workshop → produces colored cards
2. **Formulation**: Formulate Gherkin specifications from example map in `specs/`
3. **Implementation**: Write step definitions in `src/` → implement features
4. **Validation**: All scenarios pass → feature complete

### Key Points

**Before development**: Run Example Mapping, write specifications in `specs/`

**During development**: Implement steps in `src/`, write unit tests, implement features

**After development**: All scenarios pass = acceptance criteria met

---

## Key Principles

### Measurable Acceptance Criteria

**Bad** (subjective):

```gherkin
Rule: The interface is user-friendly
Rule: Performance is good
Rule: Error messages are helpful
```

**Good** (measurable):

```gherkin
Rule: Creates 3 directories (src/, tests/, docs/)
Rule: Command completes in under 2 seconds
Rule: Error message contains "already initialized" text
```

### Collaboration Before Code

BDD is **collaborative** - it requires:

- Product Owner (business perspective)
- Engineer (technical perspective)
- Tester (quality perspective)

**Don't**: Have developers write specifications alone

**Do**: Run Example Mapping workshop with all roles present

### Acceptance Criteria Drive Development

Acceptance criteria define "done":

- Development starts when criteria are clear
- Development ends when all criteria pass
- No "scope creep" mid-implementation

### Living Documentation

Gherkin specifications serve as:

- **Requirements documentation** (what the feature does)
- **Automated tests** (validation that it works)
- **Audit trail** (proof of testing for compliance)

All in one place, always up to date.

### Behavior Over Implementation

Focus on **what** the system does, not **how** it does it:

**Bad**:

```gherkin
When the ConfigManager loads the file
And the YAML parser deserializes the content
Then the Config struct is populated
```

**Good**:

```gherkin
When I run "r2r init"
Then a file named "r2r.yaml" should be created
And the configuration should contain default values
```

---

## Complete Example

### Discovery Phase: Example Mapping

**Yellow Card** (User Story):

> As a developer, I want to initialize a CLI project so that I can quickly start development

**Blue Card** (Acceptance Criteria):

> Rule: Creates project directory structure

**Green Cards** (Examples):

> 1. Initialize in empty directory → creates r2r.yaml
> 2. Initialize in existing project → shows error

**Pink Card** (Questions):

> What if directory has other files?

### Formulation Phase: Gherkin Specification

**File**: `specs/cli/init-project/specification.feature`

```gherkin
@cli @critical
Feature: cli_init-project

  As a developer
  I want to initialize a CLI project
  So that I can quickly start development

  Rule: Creates project directory structure

    @ov
    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "r2r init"
      Then a file named "r2r.yaml" should be created
      And a directory named "src" should be created
      And I should see "Project initialized successfully"

    @ov
    Scenario: Initialize in existing project
      Given I am in a directory with "r2r.yaml"
      When I run "r2r init"
      Then the command should fail
      And I should see error "Project already initialized"
```

### Implementation Phase: Step Definitions

**File**: `src/cli/tests/steps_test.go`

```go
package tests

import (
    "github.com/cucumber/godog"
)

// Feature: cli_init-project
func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^I am in an empty folder$`, iAmInAnEmptyFolder)
    ctx.Step(`^I run "([^"]*)"$`, iRun)
    ctx.Step(`^a file named "([^"]*)" should be created$`, aFileNamedShouldBeCreated)
}

func iAmInAnEmptyFolder() error {
    // Setup empty directory
    tmpDir := os.TempDir()
    os.Chdir(tmpDir)
    return nil
}

func iRun(command string) error {
    // Execute command
    cmd := exec.Command("sh", "-c", command)
    output, err := cmd.CombinedOutput()
    testContext.lastOutput = string(output)
    testContext.lastError = err
    return nil
}

func aFileNamedShouldBeCreated(filename string) error {
    // Assert file exists
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return fmt.Errorf("file %s does not exist", filename)
    }
    return nil
}
```

### Validation Phase: Running Tests

```bash
# Run BDD scenarios
godog run specs/cli/init-project/

# Output:
# Feature: cli_init-project
#   Scenario: Initialize in empty directory ✓
#   Scenario: Initialize in existing project ✓
# 2 scenarios (2 passed)
# 7 steps (7 passed)
```

---

## Specification Quality Checklist

Use this checklist to evaluate specification quality:

**Clarity**:

- [ ] Uses ubiquitous language (domain terms)
- [ ] Readable by stakeholders
- [ ] Clear Given/When/Then structure
- [ ] No technical implementation details

**Measurability**:

- [ ] Acceptance criteria are objective
- [ ] Pass/fail is unambiguous
- [ ] Observable behavior (not internal state)

**Completeness**:

- [ ] Happy path covered
- [ ] Error cases covered
- [ ] Edge cases covered
- [ ] All acceptance criteria have scenarios

**Maintainability**:

- [ ] Scenarios are independent
- [ ] Steps are reusable
- [ ] File size is manageable (< 30 scenarios)
- [ ] Synchronized with implementation

---

## Anti-Patterns to Avoid

### Testing Implementation Details

**Bad**:

```gherkin
Then the ConfigManager.Parse() method should return Config struct
And the Config.Validate() method should return no errors
```

**Good**:

```gherkin
Then the configuration file should be valid
And I should be able to run commands
```

### Vague Acceptance Criteria

**Bad**:

```gherkin
Rule: The system works well
```

**Good**:

```gherkin
Rule: Command completes within 2 seconds for projects with 100 files
```

### Skipping Collaboration

**Bad**:

- Developer writes specifications alone
- No Example Mapping workshop
- Stakeholders see specs only after implementation

**Good**:

- Example Mapping with all roles
- Specifications reviewed before coding
- Continuous stakeholder feedback

### Specifications Drift from Implementation

**Bad**:

- Code changes, specifications don't update
- Failing scenarios ignored
- Specifications become outdated documentation

**Good**:

- Specifications updated with code changes
- All scenarios pass before merging
- Specifications are living documentation

---

## Best Practices Summary

### DO

- ✅ Run Example Mapping before coding
- ✅ Write specifications in ubiquitous language
- ✅ Focus on observable behavior
- ✅ Make acceptance criteria measurable
- ✅ Keep specifications synchronized with code
- ✅ Review specifications with stakeholders
- ✅ Ensure all scenarios pass

### DON'T

- ❌ Write specifications alone
- ❌ Use technical jargon in scenarios
- ❌ Test implementation details
- ❌ Write subjective acceptance criteria
- ❌ Let specifications drift from code
- ❌ Skip stakeholder collaboration
- ❌ Ignore failing scenarios

---

## Related Documentation

- [BDD Fundamentals](./bdd-fundamentals.md) - Understanding BDD
- [Specifications Evolution](./specifications-evolution.md) - How specifications evolve
- [Three-Layer Approach](./three-layer-approach.md) - Rules/Scenarios/Unit Tests
- [Example Mapping](../discovery/example-mapping.md) - Requirements discovery
