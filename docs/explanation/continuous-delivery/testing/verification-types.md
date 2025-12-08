# Verification Types

## Introduction

Verification types are a **taxonomy for acceptance testing** that classifies tests by **what** they verify, not **how** they execute. This classification helps teams ensure comprehensive validation at **Stage 5 (Acceptance Testing)** in production-like environments (PLTE).

**Three verification types**:

- **IV (Installation Verification)**: Can the solution be deployed and configured correctly?
- **OV (Operational Verification)**: Does the solution operate as specified?
- **PV (Performance Verification)**: Does the solution meet performance requirements?

This taxonomy originates from aerospace and defense industries where comprehensive acceptance testing is critical for mission success and safety.

---

## Why Verification Types Matter

### The Problem: Incomplete Acceptance Testing

Teams often focus heavily on operational testing ("Does the feature work?") while neglecting installation and performance validation:

❌ **Common gap**: Feature works in development, fails in production due to misconfigured environment
❌ **Common gap**: Feature works functionally but performance degrades under load
❌ **Common gap**: Deployment procedure untested, fails during actual production deployment

### The Solution: Comprehensive Verification Taxonomy

Verification types ensure **three critical questions** are answered before production:

1. **IV**: Can we deploy it? (Installation procedures, configuration, infrastructure)
2. **OV**: Does it work? (Functional requirements, business logic, user workflows)
3. **PV**: Is it fast enough? (Response times, throughput, resource utilization)

All three must pass for production readiness. Missing any dimension creates risk.

### Stage 5 Integration

Verification types map directly to **Stage 5 (Acceptance Testing)**:

- **PLTE environment**: Production-like for realistic validation
- **IV, OV, PV**: All three types executed
- **Pass criteria**: 100% of IV, OV, PV tests must pass before Stage 6

Stage 6 (Extended Testing) builds on this foundation:

- Extended PV (longer-duration performance testing)
- Security testing (DAST)
- Compliance testing

---

## Installation Verification (IV)

### What It Is

IV confirms the solution can be **installed, deployed, and configured correctly** in the target environment.

**Focus**: Infrastructure, deployment procedures, configuration management

**Question answered**: "Can we deploy this successfully?"

### What IV Validates

**Deployment Procedures**:

- Deployment scripts execute successfully
- Infrastructure provisioning works (IaC)
- Application installation completes
- Database migrations apply correctly

**Dependency Satisfaction**:

- All dependencies available and correct versions
- External services reachable
- Database connections established
- Message queues accessible

**Configuration Management**:

- Configuration files loaded correctly
- Environment variables applied
- Secrets retrieved from vault
- Feature flags initialized

**Infrastructure Provisioning**:

- Compute resources created
- Networking configured
- Load balancers operational
- Storage volumes attached

### Example IV Tests

**Application starts successfully**:

```gherkin
@iv
Scenario: Application starts and responds to health checks
  Given the deployment package for version 1.2.0
  When I deploy to the PLTE environment
  Then the application should start within 30 seconds
  And the health endpoint should respond with 200 OK
```

**Database connection established**:

```gherkin
@iv
Scenario: Database connection is configured correctly
  Given the application is deployed
  When I check the database health endpoint
  Then the database should be reachable
  And the connection pool should be initialized
```

**Configuration loaded**:

```gherkin
@iv
Scenario: Application configuration is loaded from environment
  Given the application is deployed with environment variables
  When I query the application configuration
  Then the API base URL should match the PLTE environment
  And the feature flags should be loaded from configuration service
```

**Infrastructure provisioned**:

```gherkin
@iv
Scenario: Load balancer is configured correctly
  Given the infrastructure is provisioned via Terraform
  When I check the load balancer status
  Then the load balancer should be healthy
  And all application instances should be registered
```

### Why IV is Critical

**Production deployment confidence**:

- If IV passes in PLTE, deployment to production will likely succeed
- Deployment procedures are tested, not just assumed
- Configuration issues caught before production

**Infrastructure as Code validation**:

- IaC definitions validated in realistic environment
- Drift detection (does infrastructure match IaC?)
- Provisioning time measured

**Early feedback on operational concerns**:

- Startup time acceptable? (< 30 seconds for fast rollback)
- Dependencies available? (external service reachable)
- Configuration correct? (environment-specific settings)

### Common IV Test Patterns

| Pattern                    | Validates                              | Example                                |
| -------------------------- | -------------------------------------- | -------------------------------------- |
| Health check               | Application started and responsive     | `GET /health` returns 200              |
| Database connectivity      | Database connection established        | `GET /health/db` returns 200           |
| External service           | Integration with external services     | `GET /health/redis` returns 200        |
| Configuration              | Config loaded from environment         | Query config endpoint, verify values   |
| Feature flags              | Flags initialized                      | Query flags service, verify connection |
| Artifact integrity         | Correct version deployed               | Check version endpoint or manifest     |

---

## Operational Verification (OV)

### What It Is

OV ensures the solution **operates as intended**, validating functional requirements and business logic.

**Focus**: User workflows, business rules, functional correctness

**Question answered**: "Does this work correctly?"

### What OV Validates

**Functional Requirements**:

- Features work as specified
- Business logic correct
- User workflows complete end-to-end
- Edge cases handled appropriately

**User Workflows**:

- Authentication and authorization
- CRUD operations (Create, Read, Update, Delete)
- Complex workflows (checkout, payment, order processing)
- Multi-step processes

**System Behavior**:

- Error handling (meaningful error messages)
- Validation (input validation, business rule enforcement)
- State transitions (workflow states, lifecycle management)
- Integrations (service-to-service communication)

**Business Rules**:

- Calculations (pricing, tax, discounts)
- Permissions (role-based access control)
- Workflows (approval processes, state machines)
- Constraints (business rule enforcement)

### Example OV Tests

**User can complete checkout**:

```gherkin
@ov
Scenario: User completes checkout with valid payment
  Given I am logged in as a customer
  And I have items in my shopping cart
  When I proceed to checkout
  And I enter valid payment information
  Then my order should be placed successfully
  And I should receive an order confirmation email
  And my cart should be empty
```

**API validates input correctly**:

```gherkin
@ov
Scenario: API rejects invalid user registration
  Given the user registration API
  When I submit a registration with an invalid email format
  Then the API should return 400 Bad Request
  And the error message should indicate "Invalid email format"
```

**Workflow state transitions correctly**:

```gherkin
@ov
Scenario: Order progresses through fulfillment workflow
  Given I have placed an order
  When the warehouse confirms shipment
  Then the order status should change to "Shipped"
  And the customer should receive a shipping notification
```

**Integration works end-to-end**:

```gherkin
@ov
Scenario: Payment processing integrates with payment gateway
  Given I am completing checkout
  When I submit payment details
  Then the payment should be processed via the payment gateway
  And the payment gateway should return a transaction ID
  And the order should be marked as paid
```

### Why OV is Critical

**Functional correctness**:

- Features work as designed
- Business logic correct
- User workflows complete successfully

**Production confidence**:

- Realistic environment (PLTE) validates real-world behavior
- External integrations tested (not just mocked)
- End-to-end workflows validated

**Requirement traceability**:

- OV tests map directly to requirements
- Acceptance criteria validated
- Stakeholder expectations confirmed

### Common OV Test Patterns

| Pattern                | Validates                           | Example                                  |
| ---------------------- | ----------------------------------- | ---------------------------------------- |
| Happy path             | Standard workflow succeeds          | User registers, logs in, completes action|
| Error handling         | Failures handled gracefully         | Invalid input rejected with error msg    |
| Edge cases             | Boundary conditions handled         | Empty cart, max quantity, etc.           |
| Permissions            | Authorization rules enforced        | Admin can delete, user cannot            |
| Integrations           | Service-to-service communication    | API calls external service successfully  |
| State transitions      | Workflow progresses correctly       | Order: Created → Paid → Shipped          |

---

## Performance Verification (PV)

### What It Is

PV validates the solution **meets performance requirements** under expected load.

**Focus**: Response times, throughput, resource utilization, scalability

**Question answered**: "Is it fast enough?"

### What PV Validates

**Response Times**:

- P50 (median): Typical user experience
- P95: Slow but acceptable user experience
- P99: Slowest acceptable user experience
- Max: Worst-case scenario

**Throughput**:

- Requests per second (RPS)
- Transactions per second (TPS)
- Concurrent users supported

**Resource Utilization**:

- CPU usage (< 70% under expected load)
- Memory usage (< 80% under expected load)
- Disk I/O (not saturated)
- Network bandwidth (not saturated)

**Scalability**:

- Horizontal scaling (adding instances improves throughput)
- Response time stable as load increases
- No resource leaks (memory, connections)

### Example PV Tests

**API meets response time requirements**:

```gherkin
@pv
Scenario: User API responds within 200ms under normal load
  Given the system is under normal load conditions (500 req/s)
  When I measure response times for the user API
  Then the P95 response time should be under 200ms
  And the P99 response time should be under 500ms
```

**System handles expected concurrent users**:

```gherkin
@pv
Scenario: System supports 1000 concurrent users
  Given the system is deployed in the PLTE environment
  When 1000 concurrent users access the application
  Then all requests should complete successfully
  And the P95 response time should remain under 300ms
  And the error rate should be below 0.1%
```

**Resource utilization acceptable**:

```gherkin
@pv
Scenario: Resource utilization under expected load
  Given the system is under expected peak load (1000 req/s)
  When I measure resource utilization
  Then CPU usage should be below 70%
  And memory usage should be below 80%
  And no memory leaks should be detected over 1 hour
```

**Database queries performant**:

```gherkin
@pv
Scenario: Product search query performance
  Given the database contains 1 million products
  When I execute a product search query
  Then the query should complete in under 100ms
  And the database CPU should remain below 50%
```

### Why PV is Critical

**Performance regressions caught early**:

- Before production deployment
- In realistic environment (PLTE)
- Under expected load

**Performance requirements validated**:

- Objective thresholds defined
- Measured in production-like environment
- Compared to previous releases (regression detection)

**Scalability confidence**:

- Can handle expected load
- Resources scale appropriately
- No bottlenecks under pressure

### Performance Thresholds

**Typical thresholds** (vary by application):

| Metric            | Threshold Example | Rationale                               |
| ----------------- | ----------------- | --------------------------------------- |
| P95 response time | < 200ms           | 95% of users get fast response          |
| P99 response time | < 500ms           | Even slow requests are acceptable       |
| Max response time | < 2 seconds       | No request unacceptably slow            |
| Throughput        | > 1000 req/s      | Handle expected peak load               |
| CPU utilization   | < 70%             | Headroom for traffic spikes             |
| Memory usage      | < 80%             | Prevent OOM, allow for growth           |
| Error rate        | < 0.1%            | Errors rare even under load             |

### Common PV Test Patterns

| Pattern              | Validates                               | Example                              |
| -------------------- | --------------------------------------- | ------------------------------------ |
| Load test            | Performance under expected load         | 1000 concurrent users                |
| Stress test          | Behavior under extreme load             | 5x expected load                     |
| Spike test           | Handling sudden traffic increase        | 0 → 1000 users in 10 seconds         |
| Soak test            | Long-running stability (memory leaks)   | Expected load for 12 hours           |
| Scalability test     | Performance as resources scale          | 1 → 10 instances, measure throughput |

---

## Verification Types in Stage 5

### Comprehensive Acceptance Testing

**Stage 5 (Acceptance Testing)** executes all three verification types in a **production-like test environment (PLTE)**:

```text
Stage 5: Acceptance Testing (PLTE)
├── Installation Verification (IV)
│   ├── Deploy to PLTE
│   ├── Verify infrastructure provisioned
│   ├── Verify configuration loaded
│   └── Verify health checks passing
├── Operational Verification (OV)
│   ├── Verify functional requirements
│   ├── Verify user workflows
│   ├── Verify business rules
│   └── Verify integrations
└── Performance Verification (PV)
    ├── Verify response times
    ├── Verify throughput
    ├── Verify resource utilization
    └── Verify scalability
```

**Sequential execution**:

1. **IV first**: If can't deploy, no point testing functionality
2. **OV next**: If doesn't work, no point testing performance
3. **PV last**: Validate performance of working system

**Pass criteria**: All three must pass (100% test pass rate)

### Stage 5 vs Stage 6

**Stage 5 (Acceptance Testing)**:

- IV, OV, PV (baseline validation)
- Shorter duration (minutes to 1 hour)
- Expected load scenarios
- Pass/fail gate for Stage 6

**Stage 6 (Extended Testing)**:

- Extended PV (longer duration, extreme load)
- Security testing (DAST)
- Compliance testing
- Longer duration (hours)

---

## Gherkin Integration

### Tagging Verification Types

Use Gherkin tags to classify scenarios by verification type:

```gherkin
@iv
Scenario: Application starts successfully
  Given the deployment package
  When I deploy to the environment
  Then the health endpoint should respond with 200

@ov
Scenario: User can log in
  Given I am a registered user
  When I enter valid credentials
  Then I should be logged in successfully

@pv
Scenario: API handles expected load
  Given the system is under normal load
  When 1000 concurrent users make requests
  Then P95 response time should be under 200ms
```

### Benefits of Tagging

**Selective execution**:

```bash
# Run only IV tests
godog --tags @iv

# Run IV and OV, skip PV
godog --tags "@iv,@ov"

# Run PV tests only
godog --tags @pv
```

**Reporting**:

```text
Verification Type Coverage:
- IV: 15 scenarios, 15 passed (100%)
- OV: 42 scenarios, 42 passed (100%)
- PV: 8 scenarios, 8 passed (100%)
```

**Traceability**:

- Map tests to verification type
- Ensure comprehensive coverage across all types
- Identify gaps (e.g., no PV tests for critical API)

### Tag Contracts

Define expected tags in `.tag-contracts.yaml`:

```yaml
verification-types:
  description: All acceptance tests must be tagged with verification type
  required: true
  allowed_tags:
    - iv   # Installation Verification
    - ov   # Operational Verification
    - pv   # Performance Verification
  validation:
    - feature_files: "specs/acceptance/**/*.feature"
    - enforce: each scenario must have exactly one verification type tag
```

Validation ensures:

- Every acceptance test tagged with one verification type
- No tests missing verification type classification
- Coverage reports accurate

---

## Verification vs Test Levels

### Different Classification Dimensions

**Test Levels** (by scope):

- **L1 (Unit)**: Isolated components (Stage 2)
- **L2 (Integration)**: Service interactions (Stage 3)
- **L3 (System)**: End-to-end workflows (Stage 5)
- **L4 (Acceptance)**: User-facing validation (Stage 5)

**Verification Types** (by purpose):

- **IV**: Can we deploy it?
- **OV**: Does it work?
- **PV**: Is it fast enough?

**Relationship**: Verification types apply to **Stage 5 acceptance tests specifically**, while test levels apply throughout the CD Model.

### Mapping

| Verification Type | Test Level | Typical Environment | CD Model Stage |
| ----------------- | ---------- | ------------------- | -------------- |
| IV                | L3, L4     | PLTE                | Stage 5        |
| OV                | L4         | PLTE                | Stage 5        |
| PV                | L3, L4     | PLTE                | Stage 5, 6     |

**Note**: Unit tests (L1) and integration tests (L2) don't use IV/OV/PV classification - that's specific to acceptance testing in production-like environments.

---

## Anti-Patterns

### Anti-Pattern 1: Only OV Tests

**Problem**: Focusing solely on operational verification, ignoring IV and PV

**Impact**:

- Deployment procedures untested
- Performance issues discovered in production
- Incomplete acceptance validation

**Solution**: Ensure comprehensive coverage across IV, OV, PV

### Anti-Pattern 2: IV in Development Environment

**Problem**: Testing installation in development, not PLTE

**Impact**: Deployment procedures validated in unrealistic environment, fail in production

**Solution**: Execute IV in production-like environment (PLTE)

### Anti-Pattern 3: PV with Unrealistic Load

**Problem**: Testing performance with 10 users when production expects 1000

**Impact**: Performance issues not detected until production

**Solution**: Define realistic load scenarios based on production expectations

### Anti-Pattern 4: No PV Tests

**Problem**: Skipping performance verification entirely

**Impact**: Performance regressions reach production, user experience degrades

**Solution**: Define performance requirements, validate in Stage 5

### Anti-Pattern 5: Manual Verification

**Problem**: Performing IV/OV/PV manually instead of automated

**Impact**: Inconsistent, time-consuming, error-prone, not repeatable

**Solution**: Automate all verification types, integrate into CD pipeline

---

## Best Practices Summary

1. **Comprehensive coverage**: Ensure IV, OV, PV all have tests
2. **Tag consistently**: Use `@iv`, `@ov`, `@pv` tags in Gherkin
3. **Sequential execution**: IV → OV → PV (fail fast if deployment fails)
4. **Realistic environment**: Execute in PLTE, not development
5. **Objective thresholds**: Define pass/fail criteria (e.g., P95 < 200ms)
6. **Automation**: All verification types automated, integrated into pipeline
7. **Traceability**: Map tests to requirements via verification types
8. **Coverage tracking**: Report on IV/OV/PV coverage separately
9. **Stage 5 mandatory**: All three types must pass before Stage 6
10. **Extend in Stage 6**: Extended PV, security, compliance build on Stage 5

---

## Next Steps

- [Testing Strategy Overview](./testing-strategy-overview.md) - How verification types fit in overall strategy
- [CD Model Stages 1-6](../cd-model/cd-model-stages-1-6.md#stage-5-acceptance-testing) - Stage 5 in detail
- [Acceptance Testing](../cd-model/cd-model-stages-1-6.md#stage-5-acceptance-testing) - Full acceptance testing explanation
- [Environments](../architecture/environments.md#plte) - PLTE environment details

## Quick Reference

### Summary

| Type                      | Abbreviation | Purpose                         | Focus                         |
| ------------------------- | ------------ | ------------------------------- | ----------------------------- |
| Installation Verification | IV           | Confirm deployment correctness  | Infrastructure, configuration |
| Operational Verification  | OV           | Validate functional behavior    | User workflows, requirements  |
| Performance Verification  | PV           | Validate performance benchmarks | Response times, throughput    |

### Performance Metrics

| Metric            | Typical Threshold |
| ----------------- | ----------------- |
| P95 response time | < 200ms           |
| Throughput        | > 1000 req/s      |
| CPU utilization   | < 70%             |
| Memory usage      | < 80%             |

### Stage Mapping

| Stage                | Verification Types    | Environment |
| -------------------- | --------------------- | ----------- |
| Stage 5 (Acceptance) | IV, OV, PV            | PLTE        |
| Stage 6 (Extended)   | Extended PV, Security | PLTE        |

{{ diataxis_footer() }}