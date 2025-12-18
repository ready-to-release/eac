# Implementation Guides
>
> **Language-specific guides for implementing BDD specifications**

Learn how to implement executable specifications using your programming language and BDD framework.

---

## Overview

This section provides implementation-specific guidance for writing step definitions, organizing test code, and integrating BDD specifications with your test runner.

Each implementation guide covers:

- Framework installation and setup
- File organization and directory structure
- Writing step definitions
- Test level implementation (L0-L4)
- Build tags and test execution
- Best practices and patterns

---

## Available Guides

| Language | Framework | Status | Link |
|----------|-----------|--------|------|
| **Go** | Godog | ✅ Complete | [Go Implementation Guide](./go/) |

---

## Go Implementation

Complete guide for implementing BDD specifications with Go and Godog.

### Quick Start

**Installation**:

```bash
go get github.com/cucumber/godog/cmd/godog@latest
```

**Essential Commands**:

```bash
# Run all scenarios
godog run

# Run specific verification type
godog run --tags=@ov

# Run unit tests by level
go test -tags=L0 ./...
go test -tags=L2 ./...
```

**File Organization**:

```text
specs/<module>/<feature>/
└── specification.feature    # Gherkin specifications (WHAT)

src/<module>/tests/
├── <feature>_test.go        # Test runner and step definitions (HOW)
└── testdata/                # Test fixtures
```

See: [Go Implementation Guide](./go/) for complete details.

---

## Key Concepts

### 1. Separation of WHAT and HOW

**Specifications** (WHAT) are business-readable:

```gherkin
Feature: cli_init-project
  Rule: Creates project directory structure
    Scenario: Initialize in empty directory
      Given I am in an empty folder
      When I run "r2r init"
      Then a file named "r2r.yaml" should be created
```

**Step definitions** (HOW) are language-specific:

```go
func (s *StepContext) iAmInAnEmptyFolder() error {
    s.tempDir, _ = os.MkdirTemp("", "test-*")
    return os.Chdir(s.tempDir)
}
```

### 2. Test Levels with Build Tags

Different test levels run in different environments:

```go
//go:build L0
// Fast unit tests, no I/O, milliseconds

//go:build L2
// Integration tests with emulated dependencies
```

Execute by level:

```bash
go test -tags=L0 ./...  # Fast unit tests only
go test -tags=L2 ./...  # Integration tests
```

### 3. Step Definition Organization

**Recommended structure**:

```go
// TestFeatures is the test runner entry point
func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format: "pretty",
            Paths: []string{"../../specs/cli/init-project/"},
        },
    }
    suite.Run()
}

// InitializeScenario registers step definitions
func InitializeScenario(sc *godog.ScenarioContext) {
    ctx := &StepContext{}
    sc.Step(`^I am in an empty folder$`, ctx.iAmInAnEmptyFolder)
    sc.Step(`^I run "([^"]*)"$`, ctx.iRunCommand)
}
```

---

## Choosing an Implementation

### Go (Godog)

**Best for**:

- CLI tools and system utilities
- High-performance applications
- Concurrent/parallel testing
- Docker and Kubernetes tooling

**Strengths**:

- Fast execution
- Built-in concurrency primitives
- Strong typing catches errors early
- Single binary distribution

**Considerations**:

- Less mature BDD ecosystem than Ruby/Java
- Fewer step definition libraries
- Build tag system requires discipline

### Python (behave) - Planned

**Best for**:

- Data science and ML pipelines
- API testing and automation
- Web applications (Django, Flask)
- Quick prototyping

**Strengths**:

- Mature Cucumber ecosystem
- Rich library ecosystem
- Easy to read and write
- Excellent for scripting

### Java (Cucumber-JVM) - Planned

**Best for**:

- Enterprise applications
- Android applications
- Legacy system integration
- Large team environments

**Strengths**:

- Most mature Cucumber implementation
- Enterprise tool integration (Maven, Gradle)
- Strong typing and IDE support
- Extensive documentation

### TypeScript (Cucumber.js) - Planned

**Best for**:

- Web frontend applications
- Node.js backend services
- Full-stack JavaScript applications
- React/Vue/Angular testing

**Strengths**:

- JavaScript ecosystem compatibility
- TypeScript type safety
- Same language for frontend and backend
- Large BDD community

---

## Implementation Patterns

### Pattern 1: Shared Step Definitions

Reuse common steps across features:

```go
// steps/common.go
func RegisterCommonSteps(sc *godog.ScenarioContext) {
    sc.Step(`^I run "([^"]*)"$`, iRunCommand)
    sc.Step(`^the command should (succeed|fail)$`, commandStatus)
}

// tests/feature_test.go
func InitializeScenario(sc *godog.ScenarioContext) {
    RegisterCommonSteps(sc)  // Shared steps
    ctx := &FeatureContext{}
    sc.Step(`^I have a project config$`, ctx.iHaveConfig)  // Feature-specific
}
```

### Pattern 2: Context Management

Share state between steps:

```go
type StepContext struct {
    tempDir    string
    output     string
    exitCode   int
    config     *Config
}

func (s *StepContext) iRunCommand(cmd string) error {
    s.output, s.exitCode = executeCommand(cmd)
    return nil
}

func (s *StepContext) theCommandShouldSucceed() error {
    if s.exitCode != 0 {
        return fmt.Errorf("expected success, got exit code %d", s.exitCode)
    }
    return nil
}
```

### Pattern 3: Test Level Isolation

Different files for different test levels:

```
src/mymodule/tests/
├── unit_test.go           # //go:build L0
├── integration_test.go    # //go:build L2
└── system_test.go         # //go:build L3
```

---

## Getting Started

### If you're new to BDD implementation

1. **Understand the concepts first**:
   - [BDD Fundamentals](../concepts/bdd-fundamentals.md) - What is BDD?
   - [Three-Layer Approach](../concepts/three-layer-approach.md) - How layers work together

2. **Discover requirements**:
   - [Example Mapping](../discovery/example-mapping.md) - Workshop to find acceptance criteria
   - [Organizing Rules](../organization/organizing-rules.md) - Structure your specifications

3. **Choose your language guide**:
   - [Go Implementation Guide](./go/) - Currently the only complete guide

4. **Write your first feature**:
   - Start with [Go Overview](./go/overview.md) for installation
   - Follow [Step Definitions](./go/step-definitions.md) to implement steps
   - Check [Best Practices](./go/best-practices.md) for patterns

---

## Related Documentation

### Core Concepts

- [BDD Fundamentals](../concepts/bdd-fundamentals.md) - BDD principles and Gherkin
- [Three-Layer Approach](../concepts/three-layer-approach.md) - Rules/Scenarios/Unit Tests
- [Executable Specifications](../concepts/executable-specifications.md) - Living documentation

### Organization

- [File Structure](../organization/file-structure.md) - Separation of specs and implementation
- [Organizing Specifications](../organization/) - File and folder structure

### Testing Taxonomy

- [Test Levels](../taxonomy/test-levels.md) - L0-L4 execution environments
- [Verification Tags](../taxonomy/verification-tags.md) - @ov, @iv, @pv tags
- [Test Suites](../taxonomy/test-suites.md) - Pre-commit, acceptance, production

### Quality

- [Review and Iterate](../quality/review-and-iterate.md) - Maintaining specifications
