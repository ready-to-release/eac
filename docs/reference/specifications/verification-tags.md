# Verification Tags Reference

Quick reference for test classification tags used in implementation reports.

---

## Overview

Verification tags classify test scenarios for regulatory/audit documentation and implementation reports. These tags control how test results are grouped in compliance documentation.

**Usage**: Apply tags at the **Scenario level** in `behavior.feature` files

---

## Tag Definitions

| Tag | Name | Purpose | When to Use |
|-----|------|---------|-------------|
| **@IV** | Installation Verification | Verifies installation, setup, deployment, configuration, and version checks | Installation scripts, version validation, configuration setup, deployment verification |
| **@PV** | Performance Verification | Verifies performance requirements, response times, resource usage, and throughput | Response time thresholds, resource limits, concurrent operations, load testing |
| **(none)** | Operational Verification (OV) | Verifies functional behavior, business logic, and error handling | All functional scenarios not tagged with @IV or @PV (default) |

---

## Installation Verification (@IV)

### Purpose

Verifies that the system is correctly installed, configured, and ready for use.

### When to Use

Use **@IV** tag for scenarios that verify:

- Installation and deployment processes
- Version numbers and build information
- Installation paths and file locations
- Environment configuration and setup
- Baseline system configuration
- System requirements validation

### Examples

```gherkin
@success @ac1 @IV
Scenario: Install CLI on clean system
  Given a clean test environment
  When I run the installation script
  Then the CLI should be installed at "/usr/local/bin/r2r"
  And the version should be "1.0.0"
  And the configuration file should exist

@success @ac1 @IV
Scenario: Verify CLI version after installation
  Given the CLI is installed
  When I run "r2r --version"
  Then the output should show version "1.0.0"
  And the command should exit with code 0

@success @ac2 @IV
Scenario: Validate environment configuration
  Given the CLI is installed
  When I check the environment configuration
  Then all required environment variables should be set
  And the configuration file should be valid
```

### Anti-Examples (NOT @IV)

These should NOT be tagged with @IV:

```gherkin
# Functional behavior - use OV (no tag)
@success @ac1
Scenario: Run help command
  Given the CLI is installed
  When I run "r2r --help"
  Then help text should be displayed

# Error handling - use OV (no tag)
@error @ac2
Scenario: Handle invalid command
  Given the CLI is installed
  When I run "r2r invalid-command"
  Then an error message should be displayed
```

---

## Performance Verification (@PV)

### Purpose

Verifies that the system meets performance requirements and operates within specified thresholds.

### When to Use

Use **@PV** tag for scenarios that verify:

- Response time thresholds
- Resource usage limits (CPU, memory, disk)
- Throughput and concurrent operations
- Load handling capabilities
- Performance degradation boundaries
- Scalability requirements

### Examples

```gherkin
@success @ac3 @PV
Scenario: Status command responds within performance threshold
  Given the CLI is installed
  And the project is initialized
  When I run "r2r status"
  Then the command should complete within 2 seconds
  And the command should exit with code 0

@success @ac4 @PV
Scenario: Handle large file processing within time limit
  Given a file with 10000 lines exists
  When I run "r2r process large-file.txt"
  Then the processing should complete within 5 seconds
  And memory usage should stay below 100MB

@success @ac5 @PV
Scenario: Concurrent command execution
  Given the CLI is installed
  When I run 10 concurrent "r2r status" commands
  Then all commands should complete successfully
  And total execution time should be less than 5 seconds
```

### Anti-Examples (NOT @PV)

These should NOT be tagged with @PV:

```gherkin
# Functional behavior without performance requirement - use OV (no tag)
@success @ac1
Scenario: Process file successfully
  Given a file "input.txt" exists
  When I run "r2r process input.txt"
  Then the file should be processed successfully

# Even if scenario mentions timing, if it's not a performance REQUIREMENT, don't use @PV
@success @ac2
Scenario: Command completes quickly
  Given the CLI is installed
  When I run "r2r init"
  Then the command should complete
  # No specific time threshold = not @PV
```

---

## Operational Verification (OV - Default)

### Purpose

Verifies functional behavior, business logic, error handling, and user-facing features.

### When to Use

**OV is the default** - do NOT add any tag for operational verification scenarios.

Use OV (no tag) for scenarios that verify:

- Functional behavior and business logic
- Command execution and outputs
- Error handling and validation
- Data processing and transformations
- User interactions and workflows
- Integration between components
- Edge cases and boundary conditions

### Examples

```gherkin
# OV scenarios - NO @IV or @PV tag
@success @ac1
Scenario: Initialize project in empty directory
  Given I am in an empty folder
  When I run "r2r init"
  Then a file named "r2r.yaml" should be created
  And directories "src/", "tests/", "docs/" should exist
  And the command should exit with code 0

@error @ac2
Scenario: Initialize in existing project shows error
  Given I am in a directory with "r2r.yaml"
  When I run "r2r init"
  Then the command should fail
  And stderr should contain "already initialized"

@success @ac3
Scenario: Deploy to staging environment
  Given the application is built
  When I run "r2r deploy staging"
  Then the deployment should succeed
  And the deployment log should be created
```

---

## Tagging Decision Tree

Use this decision tree to determine which tag to apply:

```text
Does the scenario verify installation/setup/configuration/version?
├─ YES → Use @IV tag
└─ NO  → Does the scenario have a specific performance requirement?
          ├─ YES → Use @PV tag
          └─ NO  → Use OV (no tag) - functional/operational verification
```

---

## Running Tests by Verification Type

### Godog Commands

```bash
# Run Installation Verification scenarios only
godog --tags="@IV" specs/**/behavior.feature

# Run Performance Verification scenarios only
godog --tags="@PV" specs/**/behavior.feature

# Run Operational Verification scenarios only (exclude @IV and @PV)
godog --tags="~@IV && ~@PV" specs/**/behavior.feature
```

### Generate Separate Reports

```bash
# Generate IV report
godog --tags="@IV" --format=junit:test-results/iv-godog.xml specs/**/behavior.feature

# Generate PV report
godog --tags="@PV" --format=junit:test-results/pv-godog.xml specs/**/behavior.feature

# Generate OV report
godog --tags="~@IV && ~@PV" --format=junit:test-results/ov-godog.xml specs/**/behavior.feature
```

---

## Implementation Report Usage

Verification tags control how test results are grouped in implementation reports.

### Report Structure

```markdown
## Test Results

### Installation Verification (IV)
- Scenario: Install CLI on clean system ✓
- Scenario: Verify CLI version after installation ✓

### Operational Verification (OV)
- Scenario: Initialize project in empty directory ✓
- Scenario: Deploy to staging environment ✓
- Scenario: Handle invalid command ✓

### Performance Verification (PV)
- Scenario: Status command responds within threshold ✓
- Scenario: Handle large file processing ✓
```

### Report Template

See [Implementation Report Template](../../templates/implementation-report.md) for full structure.

---

## Best Practices

### ✅ Do

- **Tag installation scenarios** with @IV (setup, deployment, configuration)
- **Tag performance scenarios** with @PV (time limits, resource thresholds)
- **Leave functional scenarios untagged** (default to OV)
- **Use consistent tagging** across all features
- **Document performance thresholds** in scenario descriptions
- **Run verification types separately** for audit reports

### ❌ Don't

- **Don't tag functional scenarios** with @IV or @PV
- **Don't use @IV for runtime behavior** (only for installation/setup)
- **Don't use @PV without specific thresholds** (must have measurable performance requirement)
- **Don't mix verification types** in a single scenario
- **Don't forget to generate separate reports** for compliance

---

## Related Documentation

- [Gherkin Format](./gherkin-format.md) - Specification syntax and tags
- [Godog Commands](./godog-commands.md) - Running tests
- [Implementation Report Template](../../templates/implementation-report.md) - Report structure
- [Testing How-to Guides](../../how-to-guides/specifications/index.md) - Task-oriented guides
