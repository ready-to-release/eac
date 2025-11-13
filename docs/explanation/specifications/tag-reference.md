# Tag Reference

Complete reference for the **testing taxonomy tags** used across the test suite.

---

## Overview

This reference documents the **testing taxonomy tags** that define test levels, verification types, dependencies, and compliance linkage.

**Testing Taxonomy Tags**:

- **Test Level Tags** - Define execution environment and scope (`@L0`-`@L4`)
- **Verification Tags** - Categorize validation type (REQUIRED: `@ov`, `@iv`, `@pv`, `@piv`, `@ppv`)
- **System Dependencies** - Declare required tooling (`@dep:*`)
- **Risk Controls** - Link to compliance requirements (`@risk-control:<name>-<id>`)

**Note**: For organizational tags (module, priority, acceptance criteria), see [Gherkin File Organization](gherkin-concepts.md#tag-strategy)

---

## Test Level Tags

Test level tags define the execution environment and scope based on the [Testing Taxonomy](../continuous-delivery/testing/index.md).

### `@L0` - Fast Unit Tests

**Execution**: Devbox or agent
**Scope**: Source and binary
**Dependencies**: All replaced with test doubles
**Speed**: Milliseconds
**Usage**: Go tests with `//go:build L0` build tag
**Trade-off**: Highest determinism, lowest domain coherency

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

**Execution**: Devbox or agent
**Scope**: Source and binary
**Dependencies**: All replaced with test doubles
**Speed**: Seconds
**Usage**: Go tests (default, no build tag needed)
**Trade-off**: Highest determinism, lowest domain coherency

**Example**:

```go
package mypackage_test

func TestUserService_CreateUser(t *testing.T) {
    // Unit test with mocked dependencies
}
```

### `@L2` - Emulated System Tests

**Execution**: Devbox or agent
**Scope**: Deployable artifacts
**Dependencies**: All replaced with test doubles
**Speed**: Seconds
**Usage**: Godog features (default if no level tag specified)
**Trade-off**: High determinism, high domain coherency

**Example**:

```gherkin
@L2 @dep:docker @ov
Feature: Container Integration Tests
  Tests requiring Docker for artifact validation
```

### `@L3` - In-Situ Vertical Tests

**Execution**: PLTE (Production-Like Test Environment)
**Scope**: Deployed system (single deployable unit boundaries)
**Dependencies**: All replaced with test doubles
**Speed**: Minutes
**Usage**: Godog features (automatically inferred from `@iv` or `@pv`)
**Trade-off**: Moderate determinism, high domain coherency

**Example**:

```gherkin
@L3 @iv
Feature: API Service Deployment Verification
  Validates deployment in PLTE with test doubles
```

### `@L4` - Testing in Production

**Execution**: Production
**Scope**: Deployed system (cross-service interactions)
**Dependencies**: All production, may use live test doubles
**Speed**: Continuous
**Usage**: Godog features (automatically inferred from `@piv` or `@ppv`)
**Trade-off**: High determinism, highest domain coherency

**Example**:

```gherkin
@L4 @piv
Feature: Production Smoke Tests
  Validates production deployment post-release
```

**Inference Rules**:

- Go tests without build tag → `@L1`
- Go tests with `//go:build L0` → `@L0`
- Godog features without level tag → `@L2`
- Features with `@iv` or `@pv` → `@L3`
- Features with `@piv` or `@ppv` → `@L4`

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

## System Dependency Tags

System dependency tags declare required tooling for test execution.

### Available Dependencies

**`@dep:docker`** - Docker engine required
**`@dep:git`** - Git CLI required
**`@dep:go`** - Go toolchain required
**`@dep:claude`** - Claude API access required
**`@dep:az-cli`** - Azure CLI required

### Dependency Checking

**Local Development**:

- Warning + skip tests with missing dependencies
- Allows development without all tools installed

**CI Environment**:

- Fail fast on missing dependencies
- Ensures CI has required tooling

**Override**: `--dep-check=warn|fail` flag

### Example

```gherkin
@L2 @dep:docker @dep:git @ov
Feature: Container Build Pipeline
  Tests requiring Docker and Git for artifact builds

  @ov
  Scenario: Build container from Git repository
    Given I have a Git repository with Dockerfile
    When I run the container build
    Then the container image should be created
    And the image should pass security scan
```

---

## Risk Control Tags

Risk control tags link scenarios to compliance and security requirements.

### Format

`@risk-control:<name>-<id>`

**Parts**:

- `<name>` - Control name (e.g., `auth-mfa`, `encrypt-rest`, `audit-trail`)
- `<id>` - Scenario number (e.g., `01`, `02`, `03`)

### Purpose

- Links implementation scenarios to risk control specifications
- Enables compliance traceability
- Supports audit evidence collection

### Example

**Control Specification** (`specs/risk-controls/auth-mfa.feature`):

```gherkin
@risk-control:auth-mfa
Feature: Multi-Factor Authentication

  # Source: Risk Assessment RA-2025-001
  # Addresses Risk: Unauthorized access to patient data

  Rule: Authentication requires multiple factors

    @risk-control:auth-mfa-01
    Scenario: MFA required for access
      Then authentication MUST require at least two factors
      And authentication MUST occur before granting access
```

**Implementation** (`specs/cli/login/specification.feature`):

```gherkin
@ov @risk-control:auth-mfa-01
Scenario: User logs in with MFA
  Given I have valid credentials and MFA token
  When I run "simply login --mfa"
  Then I should be authenticated
  And my session should be established
```

### Traceability

```bash
# Find all implementations of a control
grep -r "@risk-control:auth-mfa" specs/

# Run tests for specific control
godog run --tags="@risk-control:auth-mfa-01"

# Run all authentication controls
godog run --tags="@risk-control:auth.*"
```

---

## Tag Inheritance

Tags accumulate from Feature → Rule → Scenario levels.

### Accumulation Rules

```gherkin
@L2 @dep:docker
Feature: Container Tests

  @ov
  Rule: Container operations

    Scenario: Start container
      # Effective tags: @L2, @dep:docker, @ov
```

### Override Rules

**Test Level Tags** (`@L0`-`@L4`):

- Scenario level overrides feature level
- Allows mixing test levels within a feature

**Dependencies** (`@dep:*`):

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

### pre-commit

**Selects**: `@L0`, `@L1`, `@L2`
**Time**: 5-10 minutes
**Purpose**: Fast pre-commit validation
**Environment**: DevBox or Build Agent
**Run**: `eac test pre-commit`

### acceptance

**Selects**: `@iv`, `@ov`, `@pv`
**Infers**: `@L3` from `@iv` and `@pv`
**Time**: 1-2 hours
**Purpose**: PLTE deployment validation
**Environment**: PLTE (Production-Like Test Environment)
**Run**: `eac test acceptance`

### production-verification

**Selects**: `@L4` AND `@piv`
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
- Declare system dependencies with `@dep:*` when needed
- Link to risk controls with `@risk-control:<name>-<id>` when applicable

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

### Tag Usage Example

```gherkin
@L2 @dep:docker
Feature: cli_container-management
  Manage Docker containers through CLI

  @ov
  Rule: Containers can be started and stopped

    @ov @risk-control:container-isolation-01
    Scenario: Start container with resource limits
      Given I have a container configuration
      When I run "simply container start --memory 512m"
      Then the container should start
      And memory limit should be enforced

    @pv
    Scenario: Container starts within 5 seconds
      Given I have a container configuration
      When I run "simply container start"
      Then the container should start within 5 seconds
```

---

## Related Documentation

- [Testing Strategy Overview](../continuous-delivery/testing/testing-strategy-overview.md) - Test taxonomy and levels
- [Testing Strategy Integration](../continuous-delivery/testing/testing-strategy-integration.md) - Test levels by CD Model stage
- [Gherkin Concepts](gherkin-concepts.md) - Organizing specifications with tags
- [Risk Controls](risk-controls.md) - Risk control tagging for compliance
- [Three-Layer Approach](three-layer-approach.md) - ATDD/BDD/TDD integration
