# Verification Tags
>
> **Operational, installation, and performance verification types**

**REQUIRED for all Gherkin scenarios**. Verification tags categorize the type of validation being performed.

---

## `@ov` - Operational Verification

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

---

## `@iv` - Installation Verification

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

---

## `@pv` - Performance Verification

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

---

## `@piv` - Production Installation Verification

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

---

## `@ppv` - Production Performance Verification

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

---

## Verification Tag Requirements

- All Gherkin scenarios MUST have at least one verification tag
- Verification tags are NOT derived - must be explicitly specified
- Multiple verification tags can be combined (e.g., `@ov @iv`)
- Go unit tests (L0-L1) do not use Gherkin verification tags

---

## When to Use Each Tag

### Use `@ov` When

- Testing business logic and functional requirements
- Validating user workflows and use cases
- Verifying feature behavior
- Testing API endpoints and data operations
- Most common verification type

### Use `@iv` When

- Validating deployment to PLTE succeeded
- Checking service health after deployment
- Verifying configuration was applied correctly
- Smoke testing after infrastructure changes
- Post-deployment sanity checks

### Use `@pv` When

- Load testing and stress testing
- Validating response time requirements
- Testing concurrent user scenarios
- Measuring throughput and latency
- Capacity planning validation

### Use `@piv` When

- Smoke testing production post-deployment
- Verifying production service is accessible
- Read-only validation in production
- Health check monitoring
- Production deployment gates

### Use `@ppv` When

- Continuous production monitoring
- Synthetic transaction monitoring
- SLA compliance validation
- Production alerting thresholds
- Performance regression detection

---

## Combining Verification Tags

Some scenarios validate multiple aspects:

```gherkin
@ov @iv
Scenario: Service deployment includes functional validation
  Given the service is deployed
  When I check the health endpoint
  Then the service should respond with status 200
  And I can create a new resource
```

This scenario validates both:

- **Installation** (`@iv`): Service deployed and responds
- **Operations** (`@ov`): Core functionality works

---

## Best Practices

### DO

- ✅ Add verification tag to ALL Gherkin scenarios
- ✅ Use `@ov` for all functional tests (most common)
- ✅ Use `@iv` for deployment validation in PLTE
- ✅ Use `@piv` for production smoke tests
- ✅ Combine tags when scenario validates multiple aspects

### DON'T

- ❌ Omit verification tags (they are REQUIRED)
- ❌ Use legacy tags (`@success`, `@failure`, `@error` - deprecated)
- ❌ Use uppercase verification tags (`@IV`, `@PV` - use lowercase)
- ❌ Confuse verification tags with test levels (they serve different purposes)

---

## Verification Tags and Test Levels

Verification tags and test levels serve different purposes:

| Verification Tag | Test Level | Meaning |
|-----------------|------------|---------|
| `@ov` | Any (L0-L4) | Functional test at any level |
| `@iv` | L3 (inferred) | Deployment test in PLTE |
| `@pv` | L3 (inferred) | Performance test in PLTE |
| `@piv` | L4 (inferred) | Smoke test in production |
| `@ppv` | L4 (inferred) | Monitoring in production |

**Example**:

```gherkin
@L2 @ov
Scenario: Functional test with emulated dependencies
  # Level: L2 (emulated system)
  # Type: Operational verification

@iv
Scenario: Deployment validation in PLTE
  # Level: L3 (inferred from @iv)
  # Type: Installation verification
```

---

## Related Documentation

- [Test Levels](./test-levels.md) - L0-L4 execution environments
- [Test Suites](./test-suites.md) - How verification tags map to test suites
- [Tag Inheritance](./tag-inheritance.md) - How tags accumulate
