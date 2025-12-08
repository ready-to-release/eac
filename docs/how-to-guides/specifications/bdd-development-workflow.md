# BDD Development Workflow

How to develop features using the three-layer testing approach.

## Workflow Overview

```text
Discovery → Specification → Implementation → Validation → Review
```

| Phase          | Activities                      | Outputs                  |
| -------------- | ------------------------------- | ------------------------ |
| Discovery      | Event Storming, Example Mapping | Domain vocabulary, cards |
| Specification  | Write Rules and Scenarios       | specification.feature    |
| Implementation | TDD with step definitions       | Production code          |
| Validation     | Run tests, stakeholder review   | Passing tests, approval  |
| Review         | Refine specs, add edge cases    | Updated specs            |

## Discovery Phase

### Step 1: Event Storming

Establish domain understanding and ubiquitous language.

**Output:** Domain vocabulary documented.

### Step 2: Example Mapping

Run 25-minute workshop with Three Amigos.

**Output:**

- Yellow card (user story)
- Blue cards (acceptance criteria)
- Green cards (examples)
- Pink cards (questions)

### Step 3: Create Specification

```gherkin
# specs/<module>/<feature>/specification.feature

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
```

### Step 4: Document Questions

Create `specs/<module>/<feature>/issues.md` for pink cards.

## Implementation Phase

### Step 1: Write Step Definitions

Create `src/<module>/tests/steps_test.go`:

```go
func (s *Suite) iAmInAnEmptyFolder(ctx context.Context) error {
    s.workDir = s.T.TempDir()
    return nil
}

func (s *Suite) iRun(ctx context.Context, cmd string) error {
    s.output, s.err = exec.Command("r2r", "init").CombinedOutput()
    return nil
}
```

### Step 2: Apply Canon TDD

For each scenario:

1. **List** behavioral variants
2. **Test** - write failing test
3. **Pass** - implement minimum code
4. **Refactor** - improve design
5. **Repeat** until done

### Step 3: Create Unit Tests

```go
// src/<module>/*_test.go

func Test_CreateConfig_InEmptyDirectory_ShouldSucceed(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "r2r.yaml")

    err := CreateConfig(configPath)

    if err != nil {
        t.Fatalf("CreateConfig failed: %v", err)
    }
}
```

### Step 4: Implement Production Code

Write code to make tests pass.

## Validation Phase

### Step 1: Run All Tests

```bash
# Run Gherkin scenarios
godog run specs/<module>/<feature>/

# Run unit tests
go test ./src/<module>/...
```

### Step 2: Verify Coverage

- All scenarios passing
- Edge cases covered
- Error handling tested

### Step 3: Stakeholder Validation

- Demo to Product Owner
- Validate behavior matches expectations
- Collect feedback

## Review Phase

### During Implementation

As you discover edge cases:

1. Add scenario to specification
2. Update step definitions
3. Add unit tests
4. Implement behavior
5. Commit spec with code

### After Implementation

- [ ] Retrospective: Did spec match build?
- [ ] Refactor for clarity
- [ ] Remove redundant scenarios
- [ ] Document lessons learned

## File Structure

```text
specs/
└── <module>/
    └── <feature>/
        ├── specification.feature  # Rules + Scenarios
        └── issues.md              # Questions to resolve

src/
└── <module>/
    ├── tests/
    │   └── steps_test.go          # Step definitions
    ├── *_test.go                  # Unit tests
    └── *.go                       # Production code
```

## Definition of Done

- [ ] All scenarios passing
- [ ] Code reviewed and refactored
- [ ] Specs synchronized with implementation
- [ ] Stakeholders validated behavior
- [ ] Edge cases discovered and tested
- [ ] Spec and code committed together

## Related

- [Canon TDD Workflow](../../reference/specifications/canon-tdd-workflow.md)
- [Three-Layer Approach](../../explanation/specifications/three-layer-approach.md)
- [Running Example Mapping](running-example-mapping.md)
- [Creating Feature Files](creating-feature-files.md)
