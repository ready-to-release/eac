# ATDD: Acceptance Test-Driven Development

**[<- Back to Testing Overview](./index.md)**

## Table of Contents

- [What is ATDD?](#what-is-atdd)
- [ATDD with Gauge](#atdd-with-gauge)
- [acceptance.spec File Structure](#acceptancespec-file-structure)
- [Gauge Installation and Setup](#gauge-installation-and-setup)
- [Example Mapping Workshop (Collaborative Discovery)](#example-mapping-workshop-collaborative-discovery)
- [Workflow](#workflow)
- [Running Gauge Tests](#running-gauge-tests)
- [Complete Example](#complete-example)
- [Related Resources](#related-resources)

---

## What is ATDD?

**Acceptance Test-Driven Development (ATDD)** is a collaborative approach where business stakeholders, developers, and testers define acceptance criteria **before** development begins. It focuses on capturing business value and measurable success criteria from the customer's perspective.

This project uses **[Gauge](https://gauge.org/)** to write and execute ATDD specifications in `acceptance.spec` files.

### Key Characteristics

| Aspect | Description |
|--------|-------------|
| **Who** | Product owner, business stakeholders, developers, testers |
| **When** | Before feature work begins |
| **Format** | Gauge specifications (markdown with executable steps) |
| **Location** | `acceptance.spec` files in `requirements/<module>/<feature>/` |
| **Purpose** | Define business value and validate acceptance criteria |
| **Tool** | [Gauge](https://gauge.org/) - Executable specification framework |

## ATDD with Gauge

Gauge allows us to write acceptance criteria as executable specifications in markdown format.

### Why Gauge for ATDD?

- **Markdown format** - Natural language, easy for non-technical stakeholders
- **Executable** - Specifications become automated tests
- **Collaborative** - Business language maps directly to test steps
- **Clear reporting** - HTML/XML reports show which acceptance criteria pass/fail
- **Test data management** - Built-in support for tables and parameters

## acceptance.spec File Structure

ATDD content appears in `acceptance.spec` files within each feature directory.

### Template Structure

**File**: `requirements/<module>/<feature_name>/acceptance.spec`

```markdown
# [Feature Name]

> **Feature ID**: <module>_<feature_name>
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: <Module>
> **Tags**: <tags>

## User Story

* As a [user role]
* I want [capability]
* So that [business value]

## Acceptance Criteria

* [Measurable criterion 1]
* [Measurable criterion 2]
* [Measurable criterion 3]
* [Measurable criterion 4]

## Acceptance Tests

### AC1: [Criterion 1 description]
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, <module>

* [Gauge step 1]
* [Gauge step 2]
* [Gauge step 3]
* [Verification step]

### AC2: [Criterion 2 description]
**Validated by**: behavior.feature -> @ac2 scenarios

* [Gauge step 1]
* [Gauge step 2]
* [Verification step]
```

### Component Breakdown

#### 1. Metadata Header

**Purpose**: Links this spec to related files and provides context

**Example**:

```markdown
# Initialize Project

> **Feature ID**: cli_init_project
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: CLI
> **Tags**: cli, critical
```

#### 2. User Story (Bullet List)

**Format**: As a / I want / So that (each as bullet point)

**Purpose**: Captures WHO needs WHAT and WHY

**Example**:

```markdown
## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure
```

**Guidelines**:

- **As a [role]**: Specify the user persona or stakeholder
- **I want [capability]**: Describe the desired functionality
- **So that [value]**: Explain the business benefit or outcome

#### 3. Acceptance Criteria (Bullet List)

**Format**: Bullet list with measurable outcomes

**Purpose**: Defines what "done" means from a business perspective

**Example**:

```markdown
## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file
* Command completes in under 2 seconds
* Works on Linux, macOS, and Windows
* Exits with clear success/error messages
* Handles existing projects gracefully
```

**Guidelines**:

- Keep to 2-6 items (if more, feature is too large)
- Each must be **measurable** (pass/fail, not subjective)
- Include functional AND non-functional requirements
- Cross-platform, performance, security, usability

#### 4. Acceptance Tests (Gauge Scenarios)

**Purpose**: Executable tests that validate each acceptance criterion

**Example**:

```markdown
## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Verify "cc.yaml" file exists
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify "docs/" directory exists

### AC2: Command completes in under 2 seconds
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: performance

* Create empty test directory
* Start performance timer
* Run "cc init" command
* Stop performance timer
* Assert execution time is less than "2" seconds
```

---

## Gauge Installation and Setup

### Installation

```bash
# Install Gauge
go install github.com/getgauge/gauge@latest

# Install Gauge for Go language
gauge install go

# Verify installation
gauge version
```

### Project Setup

Create Gauge manifest and properties:

**File**: `manifest.json` (project root)

```json
{
  "Language": "go",
  "Plugins": ["html-report", "xml-report"],
  "SpecsDir": "requirements"
}
```

**File**: `gauge.properties` (project root)

```properties
gauge_specs_dir = requirements
gauge_reports_dir = test-results/gauge
```

### Step Implementations

Gauge steps are implemented in Go test files within each feature directory.

**File**: `requirements/<module>/<feature>/acceptance_test.go`

```go
// Feature: cli_init_project
// Type: ATDD (Gauge)
package init_project_test

import (
    "github.com/getgauge-contrib/gauge-go/gauge"
    "github.com/getgauge-contrib/gauge-go/testsuit"
)

func init() {
    // Register Gauge steps
    gauge.Step("Create empty test directory", createEmptyTestDirectory)
    gauge.Step("Run <command> command", runCommand)
    gauge.Step("Verify <file> file exists", verifyFileExists)
    gauge.Step("Verify <dir> directory exists", verifyDirectoryExists)
    gauge.Step("Assert execution time is less than <seconds> seconds", assertExecutionTime)
}

func createEmptyTestDirectory() {
    // Implementation
    testsuit.T.Log("Creating empty test directory")
    // ... create temp directory
}

func runCommand(command string) {
    // Implementation
    testsuit.T.Logf("Running command: %s", command)
    // ... execute command
}

func verifyFileExists(filename string) {
    // Implementation
    testsuit.T.Logf("Verifying file exists: %s", filename)
    // ... check file exists
}

func verifyDirectoryExists(dirname string) {
    // Implementation
    testsuit.T.Logf("Verifying directory exists: %s", dirname)
    // ... check directory exists
}

func assertExecutionTime(seconds string) {
    // Implementation
    testsuit.T.Logf("Asserting execution time < %s seconds", seconds)
    // ... verify timing
}
```

---

## Example Mapping Workshop (Collaborative Discovery)

**[Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/)** is a time-boxed workshop (15-25 minutes) that uses colored index cards to discover requirements collaboratively.

### Four Card Colors

| Color | Represents | Maps To | Count |
|-------|------------|---------|-------|
| 🟡 Yellow | User Story | User Story section in acceptance.spec | 1 |
| 🔵 Blue | Rules/Acceptance Criteria | Acceptance Criteria + Acceptance Tests in acceptance.spec | 2-6 |
| 🟢 Green | Concrete Examples | Scenarios in behavior.feature (see [BDD Guide](./bdd.md)) | 2-4 per Blue |
| 🔴 Red | Questions/Uncertainties | issues.md or follow-up stories | 0-N |

### Workshop Structure

**Time**: 15-25 minutes (strictly timeboxed)

**Participants**:

- Product Owner (defines business value)
- Developer (technical feasibility)
- Tester (edge cases, scenarios)

**Process**:

1. **Place Yellow Card** (2 min)
   - Write user story: "As a [X], I want [Y], so that [Z]"
   - Place at top of table/board

2. **Generate Blue Cards** (8-12 min)
   - Brainstorm rules and acceptance criteria
   - Write one rule per Blue Card
   - Place below Yellow Card
   - Aim for 2-6 Blue Cards

3. **Create Green Cards** (5-10 min)
   - For each Blue Card, create concrete examples
   - Each Green Card: [context] → [action] → [result]
   - Place below corresponding Blue Card
   - Aim for 2-4 Green Cards per Blue Card

4. **Capture Red Cards** (ongoing)
   - Write questions or uncertainties as they arise
   - Place to the side
   - Don't try to resolve during workshop

5. **Assess Readiness** (2 min)
   - **Ready**: 2-6 Blue Cards, each with 2-4 Green Cards, few Red Cards
   - **Too Large**: >6 Blue Cards → Split into multiple stories
   - **Too Uncertain**: Many Red Cards → Needs research/spike

### Visual Layout

```text
+---------------------------------------+
| [YELLOW CARD]                         |
| As a developer, I want to init        |
| project, so that I can start quickly  |
+---------------------------------------+
          |
          v
+-----------------+  +-----------------+  +-----------------+
| [BLUE CARD 1]   |  | [BLUE CARD 2]   |  | [BLUE CARD 3]   |
| Creates dirs    |  | Generates config|  | Handles errors  |
+-----------------+  +-----------------+  +-----------------+
  |                    |                    |
  v                    v                    v
[GREEN 1a]          [GREEN 2a]          [GREEN 3a]
[GREEN 1b]          [GREEN 2b]          [GREEN 3b]

[RED CARDS - TO THE SIDE]
[RED 1] What if config exists?
[RED 2] Support --force flag?
```

### Converting Cards to acceptance.spec

#### Yellow Card → User Story Section

**Yellow Card**:

```text
As a developer
I want to initialize a CLI project with one command
So that I can quickly start development
```

**Becomes** in acceptance.spec:

```markdown
## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure
```

#### Blue Cards → Acceptance Criteria + Acceptance Tests

**Blue Card 1**:

```text
Creates project directory structure
```

**Becomes** in acceptance.spec:

```markdown
## Acceptance Criteria

* Creates project directory structure

## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Verify "cc.yaml" file exists
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify "docs/" directory exists
```

#### Green Cards → BDD Scenarios

Green Cards are converted to scenarios in `behavior.feature` (see [BDD Guide](./bdd.md#from-example-mapping-to-scenarios))

#### Red Cards → Questions Tracker

**Red Card**:

```text
What if cc.yaml already exists?
```

**Becomes** in issues.md:

```markdown
## RED-1: What if cc.yaml already exists?
**Status**: Open
**Raised**: 2025-10-30
**Decision needed by**: Product Owner
```

### Complete Example: Workshop to acceptance.spec

**Workshop Output**:

```text
[YELLOW]
As a developer, I want to initialize a CLI project with one command,
so that I can quickly start development

[BLUE-1] Creates project directory structure
  [GREEN-1a] Empty folder → init → creates src/, tests/, docs/
  [GREEN-1b] Existing project → init → error "already initialized"

[BLUE-2] Generates valid configuration file
  [GREEN-2a] New project → init → creates cc.yaml with defaults
  [GREEN-2b] With --name flag → cc.yaml contains custom name

[BLUE-3] Command completes in under 2 seconds
  [GREEN-3a] Standard project → measure time → <2s

[RED-1] What if cc.yaml already exists?
```

**Converted to acceptance.spec**:

```markdown
# Initialize Project

> **Feature ID**: cli_init_project
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: CLI
> **Tags**: cli, critical

## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure

## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file
* Command completes in under 2 seconds

## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Verify "cc.yaml" file exists
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify "docs/" directory exists

### AC2: Generates valid configuration file
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Read "cc.yaml" file contents
* Verify YAML is valid
* Verify file contains "name" field
* Verify file contains "version" field
* Verify file permissions are "0644"

### AC3: Command completes in under 2 seconds
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: performance

* Create empty test directory
* Start performance timer
* Run "cc init" command
* Stop performance timer
* Assert execution time is less than "2" seconds
```

---

## Workflow

### ATDD Development Process with Gauge

```text
ATDD Workflow (8 steps)

1. Stakeholder Meeting
   +-- Gather business requirements
   +-- Identify user personas
   +-- Understand problem domain
   +-- Schedule Example Mapping workshop

2. Define Business Value (YELLOW CARD)
   +-- Write "As a/I want/So that" user story
   +-- Focus on WHY (business value)
   +-- Get stakeholder agreement

3. Example Mapping Workshop (15-25 min)
   +-- Gather: Product Owner, Developer, Tester
   +-- Place Yellow Card (user story) at top
   +-- Discover Blue Cards (rules/acceptance criteria)
   +-- Create Green Cards (concrete examples)
   +-- Capture Red Cards (questions/blockers)
   +-- Assess: Ready, Too Large, or Too Uncertain?

4. Document Rules as Acceptance Criteria (BLUE CARDS)
   +-- Blue Cards -> Acceptance Criteria in acceptance.spec
   +-- Make each criterion testable (pass/fail)
   +-- Include functional AND non-functional requirements
   +-- Keep list focused (2-6 items)

5. Resolve Questions (RED CARDS)
   +-- Address Red Cards from workshop
   +-- Research technical constraints
   +-- Get stakeholder decisions
   +-- Update issues.md with resolutions
   +-- Create follow-up stories if necessary

6. Create acceptance.spec (Gauge)
   +-- Create feature directory structure
   +-- Add metadata header (Feature ID, links)
   +-- Write User Story section (Yellow Card)
   +-- Write Acceptance Criteria (Blue Cards)
   +-- Write Acceptance Tests (Gauge steps for each Blue Card)
   +-- Link to behavior.feature with @ac tags

7. Implement Gauge Steps
   +-- Create acceptance_test.go in feature directory
   +-- Implement step functions
   +-- Wire up to actual implementation
   +-- Run gauge run to verify
   +-- Iterate until all tests pass

8. Proceed to BDD
   +-- Convert Green Cards to behavior.feature
   +-- See: BDD Workflow -> [BDD Guide](./bdd.md#workflow)
```

### Prerequisites

Before starting ATDD:

- Business stakeholders available for Example Mapping workshop
- Problem domain understood
- User personas identified
- Gauge installed and configured

### Outputs

After completing ATDD:

- acceptance.spec file created with:
  - User story (Yellow Card)
  - Acceptance criteria (Blue Cards)
  - Executable Gauge tests
  - Feature ID and metadata
  - Links to behavior.feature
- acceptance_test.go with Gauge step implementations
- issues.md with Red Cards (if any)
- Green Cards ready for BDD conversion

---

## Running Gauge Tests

### Execute Acceptance Tests

```bash
# Run all acceptance tests
gauge run requirements/

# Run tests for specific module
gauge run requirements/cli/

# Run tests for specific feature
gauge run requirements/cli/init_project/

# Run with tags
gauge run --tags "critical" requirements/
gauge run --tags "cli & critical" requirements/

# Generate reports
gauge run --html-report requirements/
```

### Gauge Output

```text
# Initialize Project

## Acceptance Tests

  ### AC1: Creates project directory structure     ✓

  ### AC2: Generates valid configuration file     ✓

  ### AC3: Command completes in under 2 seconds   ✓


Successfully generated html-report to => test-results/gauge/html-report/index.html
Specifications: 1 executed     1 passed     0 failed     0 skipped
Scenarios:      3 executed     3 passed     0 failed     0 skipped

Total time taken: 1.234s
```

---

## Complete Example

### acceptance.spec

**File**: `requirements/cli/init_project/acceptance.spec`

```markdown
# Initialize Project

> **Feature ID**: cli_init_project
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: CLI
> **Tags**: cli, critical

## User Story

* As a developer
* I want to initialize a CLI project with a single command
* So that I can quickly start development with proper structure

## Acceptance Criteria

* Creates project directory structure
* Generates valid configuration file
* Exits with clear success/error messages
* Command completes in under 2 seconds
* Works on Linux, macOS, and Windows

## Acceptance Tests

### AC1: Creates project directory structure
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Verify "cc.yaml" file exists
* Verify "src/" directory exists
* Verify "tests/" directory exists
* Verify "docs/" directory exists
* Verify command exit code is "0"

### AC2: Generates valid configuration file
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, cli

* Create empty test directory
* Run "cc init" command
* Read "cc.yaml" file contents
* Verify YAML is valid
* Verify YAML contains key "name"
* Verify YAML contains key "version"
* Verify file permissions are "0644"

### AC3: Exits with clear success/error messages
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: critical, cli

#### Success Case
* Create empty test directory
* Run "cc init" command
* Capture stdout
* Verify stdout contains "successfully initialized"
* Verify stdout contains project path

#### Error Case
* Create directory with existing "cc.yaml"
* Run "cc init" command
* Capture stderr
* Verify stderr contains "already exists"
* Verify stderr contains "already initialized"
* Verify command exit code is "1"

### AC4: Command completes in under 2 seconds
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: performance

* Create empty test directory
* Start performance timer
* Run "cc init" command
* Stop performance timer
* Assert execution time is less than "2" seconds
* Log actual execution time

### AC5: Works on Linux, macOS, and Windows
**Validated by**: behavior.feature -> @ac5 scenarios

Tags: cross-platform

* Detect current operating system
* Create empty test directory
* Run "cc init" command
* Verify paths use OS-specific separators
* Verify file permissions are OS-appropriate
* Verify command succeeds
```

### acceptance_test.go

**File**: `requirements/cli/init_project/acceptance_test.go`

```go
// Feature: cli_init_project
// Type: ATDD (Gauge)
package init_project_test

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/getgauge-contrib/gauge-go/gauge"
    "github.com/getgauge-contrib/gauge-go/testsuit"
)

var (
    testDir         string
    commandOutput   string
    commandError    string
    exitCode        int
    startTime       time.Time
    executionTime   time.Duration
)

func init() {
    // Register Gauge steps
    gauge.Step("Create empty test directory", createEmptyTestDirectory)
    gauge.Step("Run <command> command", runCommand)
    gauge.Step("Verify <file> file exists", verifyFileExists)
    gauge.Step("Verify <dir> directory exists", verifyDirectoryExists)
    gauge.Step("Verify command exit code is <code>", verifyExitCode)
    gauge.Step("Read <file> file contents", readFileContents)
    gauge.Step("Verify YAML is valid", verifyYAMLIsValid)
    gauge.Step("Verify YAML contains key <key>", verifyYAMLContainsKey)
    gauge.Step("Start performance timer", startPerformanceTimer)
    gauge.Step("Stop performance timer", stopPerformanceTimer)
    gauge.Step("Assert execution time is less than <seconds> seconds", assertExecutionTime)
}

func createEmptyTestDirectory() {
    var err error
    testDir, err = os.MkdirTemp("", "gauge-test-*")
    if err != nil {
        testsuit.T.Errorf("Failed to create temp directory: %v", err)
    }
    testsuit.T.Logf("Created test directory: %s", testDir)
}

func runCommand(command string) {
    testsuit.T.Logf("Running command: %s", command)
    startTime = time.Now()

    // Execute command implementation
    // ... (call actual CLI code)

    executionTime = time.Since(startTime)
}

func verifyFileExists(filename string) {
    filePath := filepath.Join(testDir, filename)
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        testsuit.T.Errorf("File does not exist: %s", filePath)
    } else {
        testsuit.T.Logf("✓ File exists: %s", filePath)
    }
}

func verifyDirectoryExists(dirname string) {
    dirPath := filepath.Join(testDir, dirname)
    if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
        testsuit.T.Errorf("Directory does not exist: %s", dirPath)
    } else {
        testsuit.T.Logf("✓ Directory exists: %s", dirPath)
    }
}

func verifyExitCode(code string) {
    expectedCode := 0
    fmt.Sscanf(code, "%d", &expectedCode)
    if exitCode != expectedCode {
        testsuit.T.Errorf("Exit code mismatch: expected %d, got %d", expectedCode, exitCode)
    } else {
        testsuit.T.Logf("✓ Exit code correct: %d", exitCode)
    }
}

func startPerformanceTimer() {
    startTime = time.Now()
    testsuit.T.Log("Started performance timer")
}

func stopPerformanceTimer() {
    executionTime = time.Since(startTime)
    testsuit.T.Logf("Execution time: %v", executionTime)
}

func assertExecutionTime(seconds string) {
    var maxSeconds float64
    fmt.Sscanf(seconds, "%f", &maxSeconds)
    maxDuration := time.Duration(maxSeconds * float64(time.Second))

    if executionTime > maxDuration {
        testsuit.T.Errorf("Execution too slow: %v > %v", executionTime, maxDuration)
    } else {
        testsuit.T.Logf("✓ Performance OK: %v < %v", executionTime, maxDuration)
    }
}

// ... additional step implementations
```

---

## Related Resources

- **[BDD Guide](./bdd.md)** - Convert Green Cards to Godog scenarios
- **[TDD Guide](./tdd.md)** - Implement features with unit tests
- **[Testing Overview](./index.md)** - Complete testing strategy
- **[Gauge Documentation](https://docs.gauge.org/)** - Official Gauge docs

---

**Next**: Convert Green Cards into [Godog BDD scenarios](./bdd.md).
