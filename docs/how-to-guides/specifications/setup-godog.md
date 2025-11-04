# Setup Godog

Install and configure Godog for BDD behavior testing.

---

## Prerequisites

- Go 1.21 or later installed
- Command-line access (terminal/PowerShell)

---

## Step 1: Install Godog

Install Godog using go install:

```bash
go install github.com/cucumber/godog/cmd/godog@latest
```

---

## Step 2: Verify Godog Installation

Check that Godog is installed correctly:

```bash
godog version
```

**Expected output**:

```text
Godog version is: <version>
```

---

## Step 3: Add Godog Dependency to Project

In your project root, add Godog as a dependency:

```bash
go get github.com/cucumber/godog
```

---

## Step 4: Create Godog Configuration

### Create godog.yaml

Create `godog.yaml` in your project root with default settings:

```bash
cat > godog.yaml << 'EOF'
default:
  paths:
    - specs/**/behavior.feature
  format: pretty,junit:test-results/godog.xml
  tags: ~@wip
  strict: true
  stop-on-failure: false
EOF
```

**Configuration explained**:

- `paths`: Where to find .feature files
- `format`: Output formats (pretty console + JUnit XML)
- `tags`: Exclude work-in-progress scenarios (`~@wip` means "not @wip")
- `strict`: Fail on undefined steps
- `stop-on-failure`: Continue running all scenarios even after failure

---

## Step 5: Create Test Results Directory

```bash
mkdir -p test-results
```

---

## Step 6: Verify Setup

### Create Test Feature

Create a simple feature file to verify everything works:

```bash
mkdir -p specs/test
cat > specs/test/behavior.feature << 'EOF'
# Feature ID: test_verification
# Module: Test

@test
Feature: Godog verification

  @success
  Scenario: Verify Godog works
    Given this is a test step
    Then godog should be working
EOF
```

### Create Step Definitions

Create the step definitions:

```bash
cat > specs/test/step_definitions_test.go << 'EOF'
package test

import (
    "testing"
    "github.com/cucumber/godog"
)

func thisIsATestStep() error {
    return nil
}

func godogShouldBeWorking() error {
    return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^this is a test step$`, thisIsATestStep)
    ctx.Step(`^godog should be working$`, godogShouldBeWorking)
}

func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"behavior.feature"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}
EOF
```

### Initialize Go Module (if needed)

If you don't have a go.mod file in specs/test/:

```bash
cd specs/test
go mod init test
go get github.com/cucumber/godog
cd ../..
```

### Run Verification Test

#### Using godog command

```bash
godog specs/test/behavior.feature
```

**Expected output**:

```text
Feature: Godog verification

  Scenario: Verify Godog works          # specs/test/behavior.feature:6
    Given this is a test step           # step_definitions_test.go:7
    Then godog should be working        # step_definitions_test.go:11

1 scenarios (1 passed)
2 steps (2 passed)
```

#### Using go test

```bash
cd specs/test
go test -v
cd ../..
```

**Expected output**:

```text
=== RUN   TestFeatures
Feature: Godog verification

  Scenario: Verify Godog works
    Given this is a test step
    Then godog should be working

1 scenarios (1 passed)
2 steps (2 passed)
--- PASS: TestFeatures (0.00s)
PASS
```

---

## Step 7: Clean Up Test Files

Remove the test files after verification:

```bash
rm -rf specs/test
```

---

## Step 8: Configure for CI/CD (Optional)

### GitHub Actions Example

Create `.github/workflows/godog.yml`:

```yaml
name: BDD Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Godog
        run: go install github.com/cucumber/godog/cmd/godog@latest

      - name: Run BDD tests
        run: godog specs/**/behavior.feature

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: godog-results
          path: test-results/godog.xml
```

---

## Next Steps

- ✅ Godog is now installed and configured
- **Previous**: [Setup Gauge](./setup-gauge.md) for ATDD testing
- **Next**: [Create Feature Spec](./create-specifications.md) to start testing

---

## Related Documentation

- [Godog Commands](../../reference/specifications/godog-commands.md) - Command reference
- [Gherkin Format](../../reference/specifications/gherkin-format.md) - Specification syntax
- [ATDD and BDD with Gherkin](../../explanation/specifications/atdd-bdd-with-gherkin.md) - Understanding ATDD and BDD
- [Official Godog Docs](https://github.com/cucumber/godog) - Godog documentation
