# Tag Reference

> **Complete reference for the testing taxonomy tags used across the test suite.**

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

### Build Tags and Test Isolation

Different languages use different mechanisms to isolate test levels:

| Language | Mechanism | Example |
|----------|-----------|---------|
| Go | Build tags | `//go:build L0` |
| Python | pytest markers | `@pytest.mark.L0` |
| Java | JUnit tags | `@Tag("L0")` |
| TypeScript | Jest tags | `// @jest L0` |
| C# | NUnit categories | `[Category("L0")]` |

Consult your language implementation guide for:
- Exact syntax
- Test runner configuration
- Execution commands

### `@L0` - Fast Unit Tests

**Characteristics**:

- **Execution**: Devbox or agent
- **Scope**: Source and binary
- **Dependencies**: All replaced with test doubles (pure functions, no I/O)
- **Speed**: Milliseconds (microseconds per test)
- **Isolation**: Maximum - no external dependencies, no filesystem, no network
- **Trade-off**: Highest determinism, lowest domain coherency

**When to Use**:

- Testing business logic without side effects
- Data parsing and validation
- Calculations and transformations
- Algorithm correctness

**Example** (Gherkin):

```gherkin
@L0 @ov
Scenario: Validate email format
  Given the input "user@example.com"
  When I validate the email format
  Then the result should be valid
```

> **Implementation**: Use language-specific build tags or test markers to isolate L0 tests.
> See implementation guides for syntax:
> - [Go Implementation Guide](../../reference/specifications/go-implementation-guide.md)

### `@L1` - Unit Tests

**Characteristics**:

- **Execution**: Devbox or agent
- **Scope**: Source and binary
- **Dependencies**: All replaced with test doubles (mocked dependencies)
- **Speed**: Seconds (milliseconds per test)
- **Isolation**: High - can use temp files, simple mocks
- **Trade-off**: Highest determinism, lowest domain coherency

**When to Use**:

- Unit testing with mocked dependencies
- Testing components with filesystem access (temp directories)
- Testing with in-memory databases
- Service layer testing with mocked repositories

**Example** (Gherkin):

```gherkin
@L1 @ov
Scenario: Create user with valid data
  Given a user repository
  When I create a user with email "user@example.com"
  Then the user should be persisted
```

> **Implementation**: Use language-specific build tags or test markers to isolate L1 tests.
> See implementation guides for syntax:
> - [Go Implementation Guide](../../reference/specifications/go-implementation-guide.md)

### `@L2` - Emulated System Tests

**Characteristics**:

- **Execution**: Devbox or agent
- **Scope**: Deployable artifacts (binaries, containers)
- **Dependencies**: All replaced with test doubles (emulated services)
- **Speed**: Seconds to minutes
- **Isolation**: Moderate - uses test containers, emulated APIs
- **Trade-off**: High determinism, high domain coherency

**When to Use**:

- Integration testing with emulated dependencies
- Container and artifact validation
- Testing with test containers (databases, message queues)
- End-to-end testing with mocked external services

**Example** (Gherkin):

```gherkin
@L2 @deps:docker @ov
Feature: Container Integration Tests
  Tests requiring Docker for artifact validation

  Scenario: Container starts successfully
    Given a Docker environment
    When I start the application container
    Then the health check should pass
```

> **Implementation**: Use language-specific build tags or test markers to isolate L2 tests.
> See implementation guides for syntax:
> - [Go Implementation Guide](../../reference/specifications/go-implementation-guide.md)

### `@L3` - In-Situ Vertical Tests

**Characteristics**:

- **Execution**: PLTE (Production-Like Test Environment)
- **Scope**: Deployed system (single deployable module boundaries)
- **Dependencies**: All replaced with test doubles (isolated module testing)
- **Speed**: Minutes
- **Isolation**: Low - real infrastructure, mocked external dependencies
- **Trade-off**: Moderate determinism, high domain coherency

**When to Use**:

- Installation verification in PLTE
- Deployment validation
- Pre-production verification
- Testing deployed modules in isolation

**Example** (Gherkin):

```gherkin
@L3 @iv
Feature: API Service Deployment Verification
  Validates deployment in PLTE with test doubles

  Scenario: API responds to health check
    Given the service is deployed to PLTE
    When I call the health endpoint
    Then the response should be 200 OK
```

> **Implementation**: Use language-specific build tags or test markers to isolate L3 tests.
> See implementation guides for syntax:
> - [Go Implementation Guide](../../reference/specifications/go-implementation-guide.md)

### `@L4` - Testing in Production

**Characteristics**:

- **Execution**: Production environment
- **Scope**: Deployed system (cross-service interactions)
- **Dependencies**: All production, may use live test doubles
- **Speed**: Continuous (minutes to hours)
- **Isolation**: None - tests run against live production
- **Trade-off**: High determinism, highest domain coherency

**When to Use**:

- Production smoke tests after deployment
- Continuous production monitoring
- Post-installation verification in production
- Synthetic monitoring and health checks

**Example** (Gherkin):

```gherkin
@L4 @piv
Feature: Production Smoke Tests
  Validates production deployment post-release

  Scenario: Production health check passes
    Given the production environment
    When I check the system health
    Then all critical services should be running
    And response time should be under 200ms
```

> **Implementation**: Use language-specific build tags or test markers to isolate L4 tests.
> See implementation guides for syntax:
> - [Go Implementation Guide](../../reference/specifications/go-implementation-guide.md)

### Test Level Implementation

Test levels can be implemented using various mechanisms depending on the language and test framework:

- **Build system tags/attributes** - Language-specific build tags (e.g., Go build tags)
- **Test file naming conventions** - Naming patterns to indicate level (e.g., `*_integration_test.js`)
- **Separate test directories** - Organizing tests by level in different folders
- **Test framework categories** - Framework-specific test categories or markers

**Gherkin Scenario Inference**:

- No level tag specified → defaults to `@L2`
- Explicit `@L0`, `@L1`, `@L2`, `@L3`, or `@L4` → corresponding level
- Scenarios with `@iv` or `@pv` → inferred as `@L3` (if no explicit level)
- Scenarios with `@piv` or `@ppv` → inferred as `@L4` (if no explicit level)

> **Implementation**: Test frameworks provide various mechanisms for test isolation. See your language-specific implementation guide for details on build tags, test markers, or other isolation mechanisms.

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

**Behavior**: `@skip:` is evaluated before other selectors. Skipped tests are excluded from all test suites regardless of other tags.

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

---

## Control Tags

Control tags link scenarios to standardized security and compliance requirements using [OSCAL Risk Controls](risk-controls.md) format.

**Purpose**: Link test scenarios to OSCAL control requirements for automated compliance evidence collection.

### Control Tag Formats

**Single Control**: `@control:<control-id>`
**Multiple Controls**: `@controls:<id1>,<id2>`

**Examples**: `@control:ac-2`, `@control:ia-5(1)`, `@control:cis-5.1`

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
- Documents risk mitigation through control verification

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

## Related Documentation

- [Testing Strategy Overview](../continuous-delivery/testing/testing-strategy-overview.md) - Test taxonomy and levels
- [Testing Strategy Integration](../continuous-delivery/testing/testing-strategy-integration.md) - Test levels by CD Model stage
- [Gherkin Concepts](gherkin-concepts.md) - Organizing specifications with tags
- [Risk Controls](risk-controls.md) - Risk control tagging for compliance
- [Three-Layer Approach](three-layer-approach.md) - Rule/Scenario/Unit Test integration
