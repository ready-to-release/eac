# Tag Reference

Complete reference for the **testing taxonomy tags** used across the test suite.

---

## Overview

This reference documents the **testing taxonomy tags** that define test levels, verification types, dependencies, and compliance linkage.

**Testing Taxonomy Tags**:

- **Test Level Tags** - Define execution environment and scope (`@L0`-`@L4`)
- **Verification Tags** - Categorize validation type (REQUIRED: `@ov`, `@iv`, `@pv`, `@piv`, `@ppv`)
- **Test Execution Control** - Control test execution behavior (`@ignore`, `@Manual`)
- **System Dependencies** - Declare required tooling (`@deps:*`)
- **Risk Control Tags** - Link to compliance requirements (`@control:<id>`, `@controls:<id1>,<id2>`)

**Note:** Depending on the context you work in, you might need additional tags to support specific regulatory requirements. See [GxP Tagging](gxp-tagging.md) as an example.

---

## Test Level Tags

Test level tags define the execution environment and scope based on the [Testing Taxonomy](../continuous-delivery/testing/index.md).

### `@L0` - Fast Unit Tests

- **Execution**: Devbox or agent
- **Scope**: Source and binary
- **Dependencies**: All replaced with test doubles
- **Speed**: Milliseconds
- **Usage**: Go tests with `//go:build L0` build tag, Godog features with `@L0` tag
- **Trade-off**: Highest determinism, lowest domain coherency

**Example**:

```go
//go:build L0
// +build L0

package mypackage_test

func TestValidateEmail(t *testing.T) {
    // Very fast unit test
}
```

### `@L1` - Unit Tests

- **Execution**: Devbox or agent
- **Scope**: Source and binary
- **Dependencies**: All replaced with test doubles
- **Speed**: Seconds
- **Usage**: Go tests (default, no build tag needed), Godog features with `@L1` tag
- **Trade-off**: Highest determinism, lowest domain coherency

**Example**:

```go
package mypackage_test

func TestUserService_CreateUser(t *testing.T) {
    // Unit test with mocked dependencies
}
```

### `@L2` - Emulated System Tests

- **Execution**: Devbox or agent
- **Scope**: Deployable artifacts
- **Dependencies**: All replaced with test doubles
- **Speed**: Seconds
- **Usage**: Go tests with `//go:build L2` build tag, Godog features (default if no level tag specified)
- **Trade-off**: High determinism, high domain coherency

**Example**:

```gherkin
@L2 @deps:docker @ov
Feature: Container Integration Tests
  Tests requiring Docker for artifact validation
```

### `@L3` - In-Situ Vertical Tests

- **Execution**: PLTE (Production-Like Test Environment)
- **Scope**: Deployed system (single deployable module boundaries)
- **Dependencies**: All replaced with test doubles
- **Speed**: Minutes
- **Usage**: Go tests with `//go:build L3` build tag, Godog features with `@L3` tag (automatically inferred from `@iv` or `@pv`)
- **Trade-off**: Moderate determinism, high domain coherency

**Example**:

```gherkin
@L3 @iv
Feature: API Service Deployment Verification
  Validates deployment in PLTE with test doubles
```

### `@L4` - Testing in Production

- **Execution**: Production
- **Scope**: Deployed system (cross-service interactions)
- **Dependencies**: All production, may use live test doubles
- **Speed**: Continuous
- **Usage**: Go tests with `//go:build L4` build tag, Godog features with `@L4` tag (automatically inferred from `@piv` or `@ppv`)
- **Trade-off**: High determinism, highest domain coherency

**Example**:

```gherkin
@L4 @piv
Feature: Production Smoke Tests
  Validates production deployment post-release
```

**Inference Rules**:

**Go Tests:**

- No build tag → `@L1`
- `//go:build L0` → `@L0`
- `//go:build L2` → `@L2`
- `//go:build L3` → `@L3`
- `//go:build L4` → `@L4`

**Godog Features:**

- No level tag → `@L2`
- Explicit `@L0`, `@L1`, `@L2`, `@L3`, or `@L4` → corresponding level
- Features with `@iv` or `@pv` → `@L3` (if no explicit level tag)
- Features with `@piv` or `@ppv` → `@L4` (if no explicit level tag)

---

## Verification Tags

**REQUIRED for all Gherkin scenarios**. Verification tags categorize the type of validation being performed.

### `@ov` - Operational Verification

**Purpose**: Functional tests validating business logic
**Requirement**: REQUIRED for all functional tests
**Usage**: L2, L3, and L4
**Description**: Tests that validate the operational behavior and business logic of the system

**Example**:

```gherkin
@ov
Scenario: User creates account with valid credentials
  Given I have valid registration information
  When I register a new account
  Then my account should be created
  And I should receive a confirmation email
```

### `@iv` - Installation Verification

**Purpose**: Smoke tests validating deployment success
**Requirement**: Use for post-deployment validation
**Usage**: L3 (PLTE) - automatically infers `@L3`
**Description**: Tests that verify the system deployed correctly and can start up

**Example**:

```gherkin
@iv
Scenario: API service deploys successfully to PLTE
  Given the API service is deployed to PLTE
  When I check the health endpoint
  Then the service should respond with status 200
  And all dependencies should report healthy
```

### `@pv` - Performance Verification

**Purpose**: Load tests and performance checks
**Requirement**: Use for performance validation
**Usage**: L3 (PLTE) - automatically infers `@L3`
**Description**: Tests that validate performance requirements are met

**Example**:

```gherkin
@pv
Scenario: API responds within SLA under load
  Given the API service is deployed to PLTE
  When I send 100 requests per second
  Then 95% of requests should complete within 200ms
  And no requests should timeout
```

### `@piv` - Production Installation Verification

**Purpose**: Smoke tests in production post-deployment
**Requirement**: Use for production deployment validation
**Usage**: L4 (Production) - automatically infers `@L4`
**Description**: Tests that verify production deployment succeeded with controlled side effects

**Example**:

```gherkin
@piv
Scenario: Production service is accessible post-deployment
  Given the service is deployed to production
  When I check the production health endpoint
  Then the service should respond successfully
  And monitoring should show healthy status
```

### `@ppv` - Production Performance Verification

**Purpose**: Production monitoring and alerting
**Requirement**: Use for continuous production validation
**Usage**: L4 (Production) - automatically infers `@L4`
**Description**: Continuous validation of production performance and availability

**Example**:

```gherkin
@ppv
Scenario: Production API maintains SLA
  Given the production service is running
  When synthetic monitoring runs every 5 minutes
  Then response times should be within SLA
  And error rates should be below threshold
```

**Requirements**:

- All Gherkin scenarios MUST have at least one verification tag
- Verification tags are NOT derived - must be explicitly specified
- Multiple verification tags can be combined (e.g., `@ov @iv`)
- Go unit tests (L0-L1) do not use Gherkin verification tags

---

## Test Execution Control Tags

### `@skip:<reason>` - Exclude from Test Execution with Reason

**Purpose**: Exclude features or scenarios from test suite runs with documented reason

**Format**: `@skip:<reason>`

**Pattern**: `^@skip:(?P<reason>[a-z]+)$`

**Usage**: Feature level (excludes all scenarios) OR Scenario level (excludes only that scenario)

**Valid Reason Codes**:

| Code | Name | Description |
|------|------|-------------|
| `wip` | Work In Progress | Test implementation not yet complete |
| `broken` | Broken Test | Test is broken and needs fixing |
| `flaky` | Flaky Test | Test intermittently fails, needs stabilization |
| `deprecated` | Deprecated Feature | Feature deprecated, test kept for reference |
| `blocked` | Blocked | Blocked by external dependency or decision |

**Examples**:

```gherkin
# Feature-level: Entire feature excluded (work in progress)
@skip:wip @ov
Feature: new-feature_experimental-api
  Scenario: Create resource    # ❌ EXCLUDED (wip)
  Scenario: Delete resource    # ❌ EXCLUDED (wip)

# Scenario-level: Only OAuth scenario excluded
@ov
Feature: stable-feature_authentication
  Scenario: Valid login        # ✅ RUNS
  @skip:wip
  Scenario: OAuth (WIP)        # ❌ EXCLUDED (wip)
  Scenario: Session expiry     # ✅ RUNS

# Flaky test - needs stabilization
@ov
Feature: performance-tests_api-load
  Scenario: Basic load         # ✅ RUNS
  @skip:flaky
  Scenario: Peak load          # ❌ EXCLUDED (flaky)
```

**Behavior**: `@skip:` is evaluated before other selectors. Skipped tests are excluded from all test suites regardless of other tags.

**Migration from `@ignore`**: The `@skip:<reason>` tag replaces the deprecated `@ignore` tag. Use `@skip:wip` as the direct replacement for most `@ignore` usage.

### `@Manual` - Manual Test Scenario

**Purpose**: Mark scenarios that must be executed manually (cannot be automated)

**Usage**: Scenario level (general use across all contexts)

**Constraint**: Cannot be combined with taxonomy level tags (`@L0`-`@L4`) - manual tests are by definition not automated at any level

**When to Use**:

- Requires human judgment or subjective evaluation
- Involves physical hardware interaction
- Depends on external systems not available in test environments
- Cost/complexity of automation exceeds benefit

**Documentation Requirements**:

- Include detailed test instructions as comments
- Document expected outcomes clearly
- Specify verification criteria
- Record results systematically

**Example - Usability Testing**:

```gherkin
@Manual @ov
Scenario: User finds the interface intuitive (usability test)
  Given I am a new user
  When I attempt to complete my first task
  Then I can complete it without consulting documentation
  # Manual observation by UX researcher
```

**Example - Performance Testing**:

```gherkin
@Manual @pv @deps:load-generator
Scenario: System handles 10,000 concurrent users (load test)
  Given the production environment is scaled to maximum capacity
  When 10,000 simulated users access the system simultaneously
  Then average response time is under 200ms
  And no errors occur
  # Manual execution using third-party load testing service
```

**Example - Hardware Integration**:

```gherkin
@Manual @ov @deps:hardware
Scenario: Printer outputs receipt correctly
  Given a transaction is completed
  When the receipt is printed
  Then the receipt contains all transaction details
  And the receipt is legible
  # Manual verification of physical printout
  # Test with: Epson TM-T88V printer
```

**Pipeline Execution and Evidence Collection**:

When a test suite encounters scenarios tagged with `@Manual`, the pipeline must **pause execution** to allow manual test execution:

1. **Pipeline Pause**: Test runner stops at `@Manual` scenarios and waits for manual completion
2. **Manual Execution**: Tester executes the scenario following documented instructions
3. **Evidence Collection**: Tester collects evidence (screenshots, photos, logs, signatures, timestamps)
4. **Evidence Storage**: Evidence MUST be committed to git repository
5. **Pipeline Resume**: After evidence is committed, pipeline continues

**Critical Requirement**: Evidence **cannot** be stored in separate systems (Excel spreadsheets, SharePoint, paper logbooks only). All evidence must live in the git repository to maintain:

- Full traceability from requirement → test → evidence
- Version control of test results
- Auditability and compliance
- Single source of truth

---

## Dependency Tags

Dependency tags declare required system tools, modules, and environments for test execution.

### System Dependencies (`@deps:`)

**Format**: `@deps:<system-dependency>`

**Purpose**: Declare external system dependencies required for test execution

**Available Dependencies**:

**`@deps:docker`** - Docker engine required
**`@deps:git`** - Git CLI required
**`@deps:go`** - Go toolchain required
**`@deps:az-cli`** - Azure CLI required

**Example**:

```gherkin
@L2 @deps:docker @ov
Feature: Container build tests
  Tests requiring Docker for artifact builds
```

### Module Dependencies (`@depm:`)

**Format**: `@depm:<module-name>`

**Purpose**: Declare internal module dependencies required for test execution

**Pattern**: `^@depm:(?P<module_name>[a-z0-9-]+)$`

**Example**: `@depm:r2r-cli`, `@depm:eac-commands`

**Valid Module Names**: Loaded from module contracts at runtime

**Usage**:

```gherkin
@L2 @depm:r2r-cli @ov
Feature: CLI integration tests
  Tests requiring the r2r-cli module

  @ov
  Scenario: Execute CLI command
    Given the r2r-cli module is available
    When I run "r2r version"
    Then the version should be displayed
```

**Purpose**: Enables dependency-aware test execution and module isolation

### Environment Dependencies (`@env:`)

**Format**: `@env:<env-moniker>`

**Purpose**: Declare specific test environment requirements

**Pattern**: `^@env:(?P<env_moniker>[a-z0-9-]+)$`

**Example**: `@env:isolated-test-project`

**Valid Environment Monikers**: Defined in environment contracts

**Usage**:

```gherkin
@L2 @env:isolated-test-project @ov
Feature: Isolated project tests
  Tests requiring a clean, isolated test project environment

  @ov
  Scenario: Initialize new project
    Given I am in an isolated test environment
    When I run "r2r init"
    Then a new project should be created
    And the environment should remain isolated
```

**Purpose**: Ensures tests run in appropriate environments with correct setup

### Dependency Checking

**Local Development**:

- Warning + skip tests with missing dependencies
- Allows development without all tools installed

**CI Environment**:

- Fail fast on missing dependencies
- Ensures CI has required tooling

**Override**: `--dep-check=warn|fail` flag

### Combined Dependencies Example

```gherkin
@L2 @deps:docker @deps:git @depm:eac-core @env:ci-build @ov
Feature: Container Build Pipeline
  Tests requiring Docker, Git, eac-core module, and CI build environment

  @ov
  Scenario: Build container from Git repository
    Given I have a Git repository with Dockerfile
    When I run the container build
    Then the container image should be created
    And the image should pass security scan
```

---

## Risk and Control Tags

Risk and control tags link scenarios to risk tracking and compliance requirements.

### `@risk:<risk-id>` - Risk Tracking

**Format**: `@risk:<risk-id>`

**Purpose**: Tag specific risks for tracking and mitigation

**Pattern**: `^@risk:(?P<risk_id>[a-z0-9-]+)$`

**Example**: `@risk:data-loss`, `@risk:unauthorized-access`

**Usage**:

```gherkin
@ov @risk:data-loss
Scenario: System prevents data loss during network interruption
  Given I am uploading a large file
  When the network connection is interrupted
  Then the upload should be paused
  And I can resume the upload when connection is restored
  And no data should be lost
```

**Purpose**: Enables risk-based test organization and traceability from risk assessments to verification scenarios

**Note**: `@risk:` tags are broader than `@control:` tags - use `@risk:` for project-specific or domain risks, and `@control:` for standardized compliance controls.

### OSCAL Control Tags

OSCAL control tags link scenarios to standardized security and compliance requirements using [NIST OSCAL](https://pages.nist.gov/OSCAL/) format.

### Formats

**Single Control**: `@control:<control-id>`
**Multiple Controls**: `@controls:<id1>,<id2>`

### Control ID Format

**Pattern**: `<family>-<number>` or `<family>-<number>(<enhancement>)`

**Parts**:
- `<family>` - Control family (2-4 lowercase letters: `ac`, `au`, `ia`, `sc`, etc.)
- `<number>` - Control number (1+ digits: `2`, `12`, etc.)
- `<enhancement>` - Optional enhancement number in parentheses: `(1)`, `(10)`

**Examples**:
- `ac-2` - Account Management (NIST 800-53)
- `au-3` - Audit Record Content
- `ia-5(1)` - Password-Based Authentication (enhancement)

### Purpose

- Links test scenarios to OSCAL catalog controls
- Enables automated compliance evidence collection
- Provides standardized control traceability
- Supports audit and assessment reporting

### Examples

**Single Control**:

```gherkin
@ov @control:ac-2
Scenario: Account creation requires approval
  Given a user registration request
  When an administrator reviews the request
  Then the account should require approval
  And the approval should be logged
```

**Control with Enhancement**:

```gherkin
@ov @control:ia-5(1)
Scenario: Password authentication enforces complexity
  Given a user creating a password
  When the password is validated
  Then it must meet complexity requirements
  And weak passwords must be rejected
```

**Multiple Controls**:

```gherkin
@ov @controls:ac-2,au-3
Scenario: Account creation is audited
  Given an account creation request
  When the account is created
  Then an audit record must be created
  And the record must include timestamp, user ID, and admin approver
```

### Traceability and Evidence Collection

**Find scenarios for a control**:

```bash
# Single control
grep -r "@control:ac-2" specs/

# All account management controls
grep -r "@control:ac-" specs/
```

**Validate control tags**:

```bash
# Check all @control: tags reference valid catalog controls
validate control-tags

# Output: Reports invalid control IDs with file locations
```

**Collect evidence**:

```bash
# Run tests
test <module> --suite acceptance

# Collect compliance evidence
create risk-assess <module> --profile specs/.risk-controls/<module>.profile.json

# Output: out/risk/<module>/assessment-results.json
# Contains: Controls + Test Evidence + Satisfied/Not-Satisfied Status
```

### OSCAL Profile Integration

Control tags work with OSCAL profiles to provide complete traceability:

```
OSCAL Catalog          → Standard control definitions (NIST 800-53)
     ↓
OSCAL Profile          → Selected controls for your system
     ↓
@control: tags         → BDD scenarios verifying controls
     ↓
Test Results           → Cucumber JSON with pass/fail status
     ↓
Assessment Results     → OSCAL evidence linking controls to tests
```

**Profile Example** (`specs/.risk-controls/auth-service.profile.json`):

```json
{
  "profile": {
    "metadata": { "title": "Authentication Service Controls" },
    "imports": [{
      "href": "../../../templates/specs/risk-catalog/controls.catalog.json",
      "include-controls": [{
        "with-ids": ["ac-2", "ac-3", "au-2", "ia-5", "ia-5(1)"]
      }]
    }]
  }
}
```

When you run `create-spec` with this profile, the AI automatically suggests applicable `@control:` tags.

### Migration from Old Format

**Deprecated** (old format):
```gherkin
@risk-control:auth-mfa-01
Scenario: MFA required
```

**Current** (OSCAL format):
```gherkin
@control:ia-2(1)
Scenario: Multi-factor authentication required
```

**Migration steps**:
1. Map old controls to OSCAL equivalents (e.g., `auth-mfa` → `ia-2(1)`)
2. Create OSCAL profile with selected controls
3. Replace `@risk-control:` tags with `@control:` tags
4. Run `validate control-tags` to verify
5. Collect evidence with `create risk-assess`

### Common NIST 800-53 Control Families

| Family | Description | Example Tags |
|--------|-------------|--------------|
| **AC** | Access Control | `@control:ac-2`, `@control:ac-3` |
| **AU** | Audit and Accountability | `@control:au-2`, `@control:au-3` |
| **IA** | Identification and Authentication | `@control:ia-2`, `@control:ia-5(1)` |
| **SC** | System and Communications Protection | `@control:sc-7`, `@control:sc-8(1)` |
| **SI** | System and Information Integrity | `@control:si-2`, `@control:si-10` |
| **CM** | Configuration Management | `@control:cm-2`, `@control:cm-6` |
| **IR** | Incident Response | `@control:ir-4`, `@control:ir-6` |

See [Risk Controls](risk-controls.md) for complete documentation.

---

## Tag Inheritance

Tags accumulate from Feature → Rule → Scenario levels.

### Accumulation Rules

```gherkin
@L2 @deps:docker
Feature: Container Tests

  @ov
  Rule: Container operations

    Scenario: Start container
      # Effective tags: @L2, @deps:docker, @ov
```

### Override Rules

**Test Level Tags** (`@L0`-`@L4`):

- Scenario level overrides feature level
- Allows mixing test levels within a feature

**Dependencies** (`@deps:*`):

- Accumulate (additive)
- Scenario inherits all feature dependencies

**Verification Tags** (`@ov`, `@iv`, etc.):

- Accumulate (additive)
- Scenario can add additional verification types

### Example of Override

```gherkin
@L2
Feature: Mixed-Level Tests
  # Feature says L2 (emulated system)

  @ov
  Scenario: Fast emulated test
    # Uses L2 from feature
    # Effective tags: @L2, @ov

  @L3 @iv
  Scenario: Deployment test in PLTE
    # Overrides to L3 (PLTE environment)
    # Effective tags: @L3, @iv
```

---

## Test Suites

Test suites select tests by tags for execution at specific CD Model stages.

**Note**: All test suites automatically exclude tests tagged with `@skip:<reason>`.

### pre-commit

**Selects**: `@L0`, `@L1`, `@L2`
**Excludes**: `@skip:<reason>`, `@Manual`
**Time**: 5-10 minutes
**Purpose**: Fast pre-commit validation
**Environment**: DevBox or Build Agent
**Run**: `eac test pre-commit`

### acceptance

**Selects**: `@iv`, `@ov`, `@pv`
**Excludes**: `@skip:<reason>`, `@Manual`
**Infers**: `@L3` from `@iv` and `@pv`
**Time**: 1-2 hours
**Purpose**: PLTE deployment validation
**Environment**: PLTE (Production-Like Test Environment)
**Run**: `eac test acceptance`

### production-verification

**Selects**: `@L4` AND `@piv`
**Excludes**: `@skip:<reason>`
**Time**: Continuous
**Purpose**: Production smoke tests
**Environment**: Production
**Run**: `eac test production-verification`

---

## Best Practices

### Required Tags

✅ **DO**:

- Add verification tag to ALL Gherkin scenarios (`@ov`, `@iv`, `@pv`, `@piv`, `@ppv`)
- Use `@ov` for all functional tests
- Declare system dependencies with `@deps:*` when needed
- Link to risk controls with `@control:<id>` when applicable

❌ **DON'T**:

- Omit verification tags (they are REQUIRED)
- Use legacy tags (`@success`, `@failure`, `@error` - deprecated)
- Use uppercase verification tags (`@IV`, `@PV` - use lowercase)

### Tag Organization

✅ **DO**:

- Group related tags together by category
- Apply common tags at feature level to reduce duplication
- Use consistent ordering for readability

❌ **DON'T**:

- Over-tag (too many tags reduces clarity)
- Use custom tag schemes without documentation
- Mix tag naming conventions

### Test Execution Control

✅ **DO**:

- Use `@skip:<reason>` with appropriate reason code for excluded tests
- Use `@skip:wip` temporarily for work-in-progress features
- Use `@skip:flaky` for intermittently failing tests while investigating
- Use `@skip:broken` for tests that need fixing
- Document why tests are skipped (comments or issue links)
- Review skipped tests regularly (weekly in active development)
- Remove `@skip:` as soon as tests are stable

❌ **DON'T**:

- Use `@skip:broken` as permanent solution (fix or remove them)
- Leave skipped tests without tracking (use issue numbers)
- Skip tests for extended periods (>1 sprint)
- Use deprecated `@ignore` tag (use `@skip:wip` instead)

### Manual Tests

✅ **DO**:

- Use `@Manual` for tests that cannot be reasonably automated
- Include detailed test instructions as comments
- Document expected outcomes and verification criteria
- Store manual tests alongside automated tests in feature files
- **Commit evidence to git** (execution records, screenshots, signatures)
- Pause pipeline execution when manual tests are encountered
- Record manual test results systematically in markdown format
- Periodically review if manual tests can be automated

❌ **DON'T**:

- Use `@Manual` to avoid writing automation (automation-first approach)
- Leave manual tests without clear step-by-step instructions
- **Store evidence in separate systems** (Excel, SharePoint, paper only) - breaks traceability
- Continue pipeline without collecting and committing manual test evidence
- Forget to document why automation isn't feasible
- Let manual tests grow without review (aim to automate when possible)

### Tag Usage Example

```gherkin
@L2 @deps:docker
Feature: cli_container-management
  Manage Docker containers through CLI

  @ov
  Rule: Containers can be started and stopped

    @ov @control:sc-39
    Scenario: Start container with resource limits
      Given I have a container configuration
      When I run "r2r container start --memory 512m"
      Then the container should start
      And memory limit should be enforced

    @pv
    Scenario: Container starts within 5 seconds
      Given I have a container configuration
      When I run "r2r container start"
      Then the container should start within 5 seconds
```

---

## Related Documentation

- [Testing Strategy Overview](../continuous-delivery/testing/testing-strategy-overview.md) - Test taxonomy and levels
- [Testing Strategy Integration](../continuous-delivery/testing/testing-strategy-integration.md) - Test levels by CD Model stage
- [Gherkin Concepts](gherkin-concepts.md) - Organizing specifications with tags
- [Risk Controls](risk-controls.md) - Risk control tagging for compliance
- [Three-Layer Approach](three-layer-approach.md) - Rule/Scenario/Unit Test integration

