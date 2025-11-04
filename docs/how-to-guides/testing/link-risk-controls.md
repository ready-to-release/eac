# How to Link Risk Controls to BDD Scenarios

> **When to use**: When implementing or validating risk control requirements from risk assessments

## Overview

Risk controls are defined as **Gherkin scenarios** in `specs/risk-controls/`. User scenarios that implement these controls are tagged with `@risk<ID>` to create traceability.

**Benefits**:

- **Clear requirements**: Risk controls are explicit, testable requirements in Gherkin format
- **Traceability**: Direct link from control definition to implementation scenarios
- **Audit-ready**: Risk controls documented in version control with complete history
- **Reusability**: Multiple features can reference the same risk control
- **Change management**: All control changes tracked automatically in Git

## Step-by-Step Guide

### Step 1: Create Risk Control Scenario

Create or update a feature file in `specs/risk-controls/`:

**File**: `specs/risk-controls/authentication-controls.feature`

```gherkin
Feature: Authentication Risk Controls

  Risk controls related to user authentication
  Source: Assessment-2025-001

  @risk1
  Scenario: RC-001 - User authentication required
    Given a system with protected resources
    Then all user access MUST be authenticated
    And authentication MUST occur before granting access
    And failed authentication attempts MUST be logged
```

**Key points**:

- Tag with `@risk<ID>` (e.g., `@risk1`, `@risk2`)
- Use scenario name format: `RC-<ID> - <Description>`
- Use MUST for mandatory requirements
- Reference source assessment in feature description

### Step 2: Tag User Scenarios

In your feature implementation, tag scenarios with `@risk<ID>`:

**File**: `specs/cli/user-authentication/behavior.feature`

```gherkin
@cli @critical @security
Feature: User Authentication

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

Check that your scenario correctly links to the risk control:

```bash
# Find risk control definition
grep -A 5 "@risk1" specs/risk-controls/

# Find all implementations
grep -r "@risk1" specs/ --exclude-dir=risk-controls
```

---

## Common Patterns

### Pattern 1: One Risk Control, Multiple Scenarios

```gherkin
# Risk control
@risk1
Scenario: RC-001 - User authentication required
  Then all user access MUST be authenticated

# Implementation scenarios
@success @risk1
Scenario: Login with valid credentials
  ...

@error @risk1
Scenario: Login with invalid credentials
  ...

@error @risk1
Scenario: Access without authentication blocked
  ...
```

### Pattern 2: One Scenario, Multiple Risk Controls

```gherkin
# Implementation validates multiple controls
@success @ac1 @risk1 @risk5
Scenario: Authenticated user action creates audit trail
  Given I am authenticated as "admin"
  When I run "simply config set key=value"
  Then I should be authenticated          # @risk1
  And an audit entry should be created    # @risk5
```

### Pattern 3: Risk Control Spanning Multiple Features

```gherkin
# Risk control
@risk5
Scenario: RC-005 - Audit trail completeness
  Then all changes MUST create audit trail entries

# Feature 1: CLI audit logging
@success @risk5
Scenario: CLI command creates audit entry
  ...

# Feature 2: VSCode commit creates audit entry
@success @risk5
Scenario: Commit creates audit entry
  ...
```

---

## Risk Control Organization

Group related controls in feature files:

```text
specs/risk-controls/
├── authentication-controls.feature    # @risk1, @risk2, @risk3
├── data-protection-controls.feature   # @risk10, @risk11, @risk12
├── audit-controls.feature             # @risk5, @risk6, @risk7
├── privacy-controls.feature           # @risk20, @risk21
└── ai-controls.feature                # @risk30, @risk31
```

## Documenting Source Assessment

Each risk control feature file should reference the source assessment document:

**Pattern**:

```gherkin
Feature: [Control Category] Risk Controls

  [Brief description of the controls in this file]
  Source: <Assessment-ID>
  Assessment Date: <YYYY-MM-DD>
```

**Example**:

```gherkin
Feature: Authentication Risk Controls

  Risk controls related to user authentication and access control
  Source: Assessment-2025-001
  Assessment Date: 2025-01-15
```

**Best practices**:

- Include assessment ID for traceability
- Add assessment date for version tracking
- Keep feature description brief and focused
- Group related controls in the same feature file

---

## Querying and Reporting

### Find All Implementations of a Risk Control

```bash
# Find all scenarios implementing risk1
grep -r "@risk1" specs/ --exclude-dir=risk-controls

# With context (show scenario name)
grep -B 2 "@risk1" specs/ --exclude-dir=risk-controls
```

### Generate Coverage Report

```bash
# Count implementations per risk control
echo "Risk Control | Implementation Scenarios"
echo "-------------|-------------------------"

grep -r "@risk" specs/risk-controls/ | \
  grep -oP '@risk\K[0-9]+' | \
  sort -n | uniq | \
  while read id; do
    count=$(grep -r "@risk$id" specs/ --exclude-dir=risk-controls | wc -l)
    echo "Risk $id | $count scenarios"
  done
```

### Find Risk Controls Without Implementation

```bash
# List risk controls that have no tagged scenarios
grep -r "@risk" specs/risk-controls/ | \
  grep -oP '@risk\K[0-9]+' | \
  sort -n | uniq | \
  while read id; do
    count=$(grep -r "@risk$id" specs/ --exclude-dir=risk-controls | wc -l)
    if [ $count -eq 0 ]; then
      echo "Risk $id: No implementations found"
    fi
  done
```

### Generate Traceability Matrix

```bash
# Create CSV: Risk ID, Control Description, Implementation Count, Feature Files
echo "Risk ID,Control,Implementations,Features" > risk-traceability.csv

grep -r "@risk" specs/risk-controls/ | \
  while read line; do
    file=$(echo "$line" | cut -d: -f1)
    id=$(echo "$line" | grep -oP '@risk\K[0-9]+')
    desc=$(grep -A 1 "@risk$id" "$file" | grep "Scenario:" | cut -d: -f2- | xargs)
    count=$(grep -r "@risk$id" specs/ --exclude-dir=risk-controls | wc -l)
    features=$(grep -r "@risk$id" specs/ --exclude-dir=risk-controls | cut -d: -f1 | sort -u | xargs)
    echo "$id,\"$desc\",$count,\"$features\"" >> risk-traceability.csv
  done
```

## Best Practices

1. **Define before implement**: Create risk control scenario before tagging user scenarios

2. **Use clear IDs**: Number risk controls sequentially (1, 2, 3...) or by category (10-19 auth, 20-29 data, etc.)

3. **Group related controls**: Keep related controls in same feature file

4. **Document source**: Reference assessment document in feature description

5. **Use MUST for requirements**: Risk control scenarios should use "MUST" to indicate mandatory requirements

6. **Keep controls atomic**: Each risk control should address one requirement

7. **Tag at scenario level**: Prefer scenario-level tags over feature-level for precise traceability

---

## Example: Complete Workflow

### 1. Review Assessment Document

```text
Assessment-2025-001
Risk ID: R-023
Control: All user access must be authenticated before granting system access
Control Type: Preventive
```

### 2. Create Risk Control Scenario

```gherkin
# specs/risk-controls/authentication-controls.feature

@risk1
Scenario: RC-001 - User authentication required (R-023)
  Given a system with protected resources
  Then all user access MUST be authenticated
  And authentication MUST occur before granting access
  And failed authentication attempts MUST be logged
```

### 3. Implement Feature with Tagged Scenarios

```gherkin
# specs/cli/user-authentication/behavior.feature

@cli @critical @security
Feature: User Authentication

  @success @ac1 @risk1
  Scenario: Valid credentials grant access
    Given I have valid credentials
    When I run "simply login --user admin --password ***"
    Then I should be authenticated

  @error @ac1 @risk1
  Scenario: Invalid credentials deny access
    Given I have invalid credentials
    When I run "simply login --user admin --password wrong"
    Then I should not be authenticated
    And the failed attempt should be logged
```

---

## Related Documentation

- [BDD Format Reference](../../reference/testing/bdd-format.md)
- [Acceptance Criteria Tagging](../../reference/testing/atdd-format.md)
