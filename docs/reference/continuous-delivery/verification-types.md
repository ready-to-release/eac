# Verification Types

Reference for verification types used in acceptance testing (Stage 5).

## Summary

| Type                      | Abbreviation | Purpose                         | Focus                         |
| ------------------------- | ------------ | ------------------------------- | ----------------------------- |
| Installation Verification | IV           | Confirm deployment correctness  | Infrastructure, configuration |
| Operational Verification  | OV           | Validate functional behavior    | User workflows, requirements  |
| Performance Verification  | PV           | Validate performance benchmarks | Response times, throughput    |

## Installation Verification (IV)

**Purpose:** Confirms the solution can be installed and configured correctly.

**Validates:**

- Deployment scripts and procedures
- Dependency satisfaction
- Configuration management
- Infrastructure provisioning

**Example Checks:**

- Application starts successfully
- Database connections established
- Configuration files loaded
- Health endpoints responding

## Operational Verification (OV)

**Purpose:** Ensures the solution operates as intended.

**Validates:**

- Functional requirements
- User workflows end-to-end
- System behavior under normal conditions
- Business rules and logic

**Example Checks:**

- User can log in and perform actions
- Data flows correctly through system
- Error handling works as specified
- Integrations function correctly

## Performance Verification (PV)

**Purpose:** Validates performance benchmarks are met.

**Validates:**

- Response times under expected load
- Resource utilization (CPU, memory, disk)
- Scalability characteristics
- Throughput capacity

**Example Metrics:**

| Metric            | Typical Threshold |
| ----------------- | ----------------- |
| P95 response time | < 200ms           |
| Throughput        | > 1000 req/s      |
| CPU utilization   | < 70%             |
| Memory usage      | < 80%             |

## Gherkin Tags

Use tags to identify verification type in specifications:

```gherkin
@iv
Scenario: Application starts successfully
  Given the deployment package
  When I deploy to the environment
  Then the health endpoint should respond with 200

@ov
Scenario: User can complete checkout
  Given I have items in my cart
  When I complete the checkout process
  Then my order should be placed

@pv
Scenario: API handles expected load
  Given the system is under normal conditions
  When 1000 concurrent users make requests
  Then P95 response time should be under 200ms
```

## Stage Mapping

| Stage                | Verification Types    | Environment |
| -------------------- | --------------------- | ----------- |
| Stage 5 (Acceptance) | IV, OV, PV            | PLTE        |
| Stage 6 (Extended)   | Extended PV, Security | PLTE        |

## Related

- [Acceptance Testing](../../explanation/continuous-delivery/cd-model/cd-model-stages-1-6.md#stage-5-acceptance-testing)
- [Tag Reference](../specifications/tag-reference.md)
- [Testing Strategy](../../explanation/continuous-delivery/testing/testing-strategy-overview.md)
