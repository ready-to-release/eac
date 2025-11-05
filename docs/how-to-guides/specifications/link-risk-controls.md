# Link Risk Controls to Scenarios

> **When to use**: Implementing or validating risk control requirements from assessments

## Overview

Risk controls are **Gherkin scenarios** in `specs/risk-controls/`. User scenarios implementing these controls are tagged with `@risk<ID>` for traceability.

**Benefits**: Clear requirements, audit trail, traceability, reusability, version control

---

## Quick Guide

### Step 1: Define Risk Control

**File**: `specs/risk-controls/authentication-controls.feature`

**Naming convention**: `[Control Category]-[Name].feature` (use dashes, no spaces, lowercase)

```gherkin
Feature: authentication-controls

  Risk controls for user authentication
  Source: Assessment-2025-001

  @risk1
  Scenario: RC-001 - User authentication required
    Given a system with protected resources
    Then all user access MUST be authenticated
    And authentication MUST occur before granting access
    And failed authentication attempts MUST be logged
```

**Format**:

- Tag: `@risk<ID>` (e.g., `@risk1`, `@risk2`)
- Name: `RC-<ID> - <Description>`
- Use MUST for mandatory requirements

### Step 2: Tag Implementation Scenarios

**File**: `specs/cli/user-authentication/specification.feature`

```gherkin
@cli @critical @security
Feature: cli_user-authentication

  As a system administrator
  I want secure user authentication
  So that only authorized users can access the system

  Rule: All access requires valid authentication

    @success @ac1 @risk1
    Scenario: Valid credentials grant access
      Given I have valid credentials
      When I run "simply login --user admin --password ***"
      Then I should be authenticated
      And my session should be active

    @error @ac1 @risk1
    Scenario: Invalid credentials deny access
      Given I have invalid credentials
      When I run "simply login --user wrong --password wrong"
      Then I should not be authenticated
      And the failed attempt should be logged
```

### Step 3: Verify Traceability

```bash
# Find risk control definition
grep -A 5 "@risk1" specs/risk-controls/

# Find all implementations
grep -r "@risk1" specs/ --exclude-dir=risk-controls
```

---

## Common Patterns

### One Risk Control → Multiple Scenarios

```gherkin
# Risk control
@risk1
Scenario: RC-001 - User authentication required
  Then all user access MUST be authenticated

# Implementation scenarios
@success @ac1 @risk1
Scenario: Login with valid credentials
  ...

@error @ac1 @risk1
Scenario: Login with invalid credentials
  ...

@error @ac1 @risk1
Scenario: Access without authentication blocked
  ...
```

### One Scenario → Multiple Risk Controls

```gherkin
@success @ac1 @risk1 @risk5
Scenario: Authenticated action creates audit trail
  Given I am authenticated as "admin"
  When I run "simply config set key=value"
  Then I should be authenticated          # @risk1
  And an audit entry should be created    # @risk5
```

### One Risk Control → Multiple Features

```gherkin
# Risk control
@risk5
Scenario: RC-005 - Audit trail completeness
  Then all changes MUST create audit trail entries

# Feature 1: CLI
@success @ac2 @risk5
Scenario: CLI command creates audit entry
  ...

# Feature 2: VSCode
@success @ac3 @risk5
Scenario: Commit creates audit entry
  ...
```

---

## Organization

### Directory Structure

**Naming**: `[Control Category]-[Name].feature` (use dashes, no spaces, lowercase)

```text
specs/risk-controls/
├── authentication-controls.feature      # @risk1, @risk2, @risk3
├── data-protection-controls.feature     # @risk10, @risk11, @risk12
├── audit-trail-controls.feature         # @risk5, @risk6, @risk7
├── privacy-controls.feature             # @risk20, @risk21
└── ai-model-controls.feature            # @risk30, @risk31
```

### Feature File Format

```gherkin
Feature: [Control Category]-[Name]

  [Brief description]
  Source: <Assessment-ID>
  Assessment Date: <YYYY-MM-DD>
```

**Examples**:

- `Feature: authentication-controls`
- `Feature: data-protection-encryption`
- `Feature: audit-trail-controls`

---

## Reporting

### Find All Implementations

```bash
# Find scenarios implementing risk1
grep -r "@risk1" specs/ --exclude-dir=risk-controls

# With scenario names
grep -B 2 "@risk1" specs/ --exclude-dir=risk-controls
```

### Coverage Report

```bash
# Count implementations per control
grep -r "@risk" specs/risk-controls/ | \
  grep -oP '@risk\K[0-9]+' | \
  sort -n | uniq | \
  while read id; do
    count=$(grep -r "@risk$id" specs/ --exclude-dir=risk-controls | wc -l)
    echo "Risk $id: $count scenarios"
  done
```

### Find Unimplemented Controls

```bash
# List controls without implementations
grep -r "@risk" specs/risk-controls/ | \
  grep -oP '@risk\K[0-9]+' | \
  sort -n | uniq | \
  while read id; do
    count=$(grep -r "@risk$id" specs/ --exclude-dir=risk-controls | wc -l)
    [ $count -eq 0 ] && echo "Risk $id: Not implemented"
  done
```

---

## Best Practices

1. **Define first**: Create risk control before tagging user scenarios
2. **Clear IDs**: Sequential (1, 2, 3) or categorical (10-19 auth, 20-29 data)
3. **Group related**: Keep related controls in same feature file
4. **Document source**: Reference assessment ID in feature description
5. **Use MUST**: Indicate mandatory requirements in control scenarios
6. **Atomic controls**: One requirement per control
7. **Scenario-level tags**: Prefer scenario tags over feature tags for precision

---

## Related Documentation

- [Risk Controls Explanation](../../explanation/specifications/risk-controls.md) - Understanding risk controls and identifying relevant controls for your domain
- [Gherkin Format Reference](../../reference/specifications/gherkin-format.md) - Specification syntax and tagging
- [Create Specifications](./create-specifications.md) - Write specification.feature files
