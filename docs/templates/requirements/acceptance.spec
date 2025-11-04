# [Feature Name]

> **Feature ID**: `<component>_<feature_name>`
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Component**: `<Component Name>` (e.g., API, UI, Backend, Service, CLI)
> **Tags**: `<tags>`
> **Priority**: `<Critical | High | Medium | Low>`
> **Risk Control**: `@risk<ID>` (optional - if this feature implements risk control requirements)

## Risk Control (Optional)

If this feature implements a risk control from a risk assessment:

* **Risk ID**: `<RISK-ID>` (e.g., R-023, R-015)
* **Assessment**: `<ASSESSMENT-ID>` (e.g., Assessment-2025-001)
* **Control Requirement**: [Brief description of what the control requires]
* **Control Type**: [Preventive | Detective | Corrective]
* **Risk Control Tag**: `@risk<ID>` (e.g., @risk1, @risk5)

**Traceability**: Risk control defined in [specs/risk-controls/](../../risk-controls/)

---

## User Story

**As a** [role/persona - e.g., end user, developer, administrator]
**I want** [capability/functionality - what the user needs to do]
**So that** [business value/benefit - why this matters]

### Business Context

[Brief explanation of why this feature is needed and its impact on the user, business, or system. Include any relevant background information.]

---

## Acceptance Criteria

### AC1: [First Acceptance Criterion - Brief Title]

**Type**: [Installation Verification (IV) | Operational Verification (OV) | Performance Verification (PV)]

[Detailed description of what must be true for this criterion to be met. Be specific and measurable.]

**Validated by**: behavior.feature -> `@ac1` scenarios

**Success Metrics**:
* [Measurable metric 1 - e.g., "Configuration file is created with valid schema"]
* [Measurable metric 2 - e.g., "User can complete action without errors"]
* [Measurable metric 3 - e.g., "System responds with expected output"]

---

### AC2: [Second Acceptance Criterion - Brief Title]

**Type**: [Installation Verification (IV) | Operational Verification (OV) | Performance Verification (PV)]

[Detailed description of what must be true for this criterion to be met. Be specific and measurable.]

**Validated by**: behavior.feature -> `@ac2` scenarios

**Success Metrics**:
* [Measurable metric 1]
* [Measurable metric 2]
* [Measurable metric 3]

---

### AC3: [Third Acceptance Criterion - Brief Title]

**Type**: [Installation Verification (IV) | Operational Verification (OV) | Performance Verification (PV)]

[Detailed description of what must be true for this criterion to be met. Be specific and measurable.]

**Validated by**: behavior.feature -> `@ac3` scenarios

**Success Metrics**:
* [Measurable metric 1]
* [Measurable metric 2]
* [Measurable metric 3]

---

## Acceptance Tests

These tests are written in Gauge markdown format and validate the acceptance criteria above.

### AC1: [First Acceptance Criterion]
**Validated by**: behavior.feature -> `@ac1` scenarios

#### Scenario: [Descriptive scenario name]

* [Step describing precondition or setup - e.g., "System is in initial state"]
* [Step describing action performed - e.g., "User performs action X"]
* [Step describing expected outcome - e.g., "Result Y is produced"]
* [Step describing verification - e.g., "Output contains expected data"]

**Expected Result**: [Clear statement of what success looks like]

---

### AC2: [Second Acceptance Criterion]
**Validated by**: behavior.feature -> `@ac2` scenarios

#### Scenario: [Descriptive scenario name]

* [Step describing precondition or setup]
* [Step describing action performed]
* [Step describing expected outcome]
* [Step describing verification]

**Expected Result**: [Clear statement of what success looks like]

---

### AC3: [Third Acceptance Criterion]
**Validated by**: behavior.feature -> `@ac3` scenarios

#### Scenario: [Descriptive scenario name]

* [Step describing precondition or setup]
* [Step describing action performed]
* [Step describing expected outcome]
* [Step describing verification]

**Expected Result**: [Clear statement of what success looks like]

---

## Example Mapping Summary

This section documents the Example Mapping workshop results using the card metaphor:
- **Yellow Card** = User Story (the feature we're building)
- **Blue Cards** = Rules/Acceptance Criteria (what must be true)
- **Green Cards** = Concrete Examples (specific test cases)

### Yellow Card (User Story)
**As a** [role] **I want** [capability] **so that** [benefit]

### Blue Card 1: [Rule/Acceptance Criterion 1]
* **Green Card 1a**: [Context] → [Action] → [Expected Result]
* **Green Card 1b**: [Context] → [Action] → [Expected Result]
* **Green Card 1c**: [Context] → [Action] → [Expected Result]

### Blue Card 2: [Rule/Acceptance Criterion 2]
* **Green Card 2a**: [Context] → [Action] → [Expected Result]
* **Green Card 2b**: [Context] → [Action] → [Expected Result]
* **Green Card 2c**: [Context] → [Action] → [Expected Result]

### Blue Card 3: [Rule/Acceptance Criterion 3]
* **Green Card 3a**: [Context] → [Action] → [Expected Result]
* **Green Card 3b**: [Context] → [Action] → [Expected Result]

**Red Cards (Questions)**:
* [Unresolved question 1 - track in issues.md if needed]
* [Unresolved question 2]

---

## Dependencies

### System Requirements
* [Requirement 1: e.g., Runtime version >= X.Y]
* [Requirement 2: e.g., Database available]
* [Requirement 3: e.g., External service configured]

### Feature Dependencies
* [Feature ID or name that must exist first]
* [Another dependency if applicable]

---

## Non-Functional Requirements

### Performance
* [Performance requirement: e.g., "Operation must complete within X seconds"]
* [Resource requirement: e.g., "Memory usage must not exceed Y MB"]
* [Throughput requirement: e.g., "Must handle Z requests per second"]

### Security
* [Security requirement 1: e.g., "Authentication required for all operations"]
* [Security requirement 2: e.g., "Data encrypted in transit and at rest"]

### Usability
* [Usability requirement 1: e.g., "Error messages must be clear and actionable"]
* [Usability requirement 2: e.g., "Interface follows accessibility guidelines"]

### Reliability
* [Reliability requirement 1: e.g., "Uptime of 99.9%"]
* [Reliability requirement 2: e.g., "Graceful degradation on service failure"]

---

## Risks and Assumptions

### Assumptions
* [Assumption 1: e.g., "Users have necessary permissions"]
* [Assumption 2: e.g., "Network connectivity is available"]
* [Assumption 3: e.g., "Input data follows expected schema"]

### Risks
* **[Risk 1 Title]**: [Description] - **Mitigation**: [How we'll address it]
* **[Risk 2 Title]**: [Description] - **Mitigation**: [How we'll address it]

---

## Questions and Open Issues

Track unresolved questions in [issues.md](./issues.md) if needed.

* [ ] [Question 1: e.g., "Should we support offline mode?"]
* [ ] [Question 2: e.g., "What's the maximum allowed input size?"]
* [ ] [Question 3: e.g., "How should we handle concurrent modifications?"]

---

## Related Documentation

* [Link to design documents]
* [Link to API specifications]
* [Link to architecture diagrams]
* [Link to related features]

---

## Verification Type Summary

This feature includes the following verification types for implementation reports:

| Verification Type | Acceptance Criteria | Scenario Count |
|-------------------|---------------------|----------------|
| **Installation Verification (IV)** | AC[X] | [N scenarios with @IV tag] |
| **Operational Verification (OV)** | AC[Y] | [M scenarios without @IV or @PV] |
| **Performance Verification (PV)** | AC[Z] | [P scenarios with @PV tag] |

**Total Scenarios**: [N + M + P]

**Verification Type Definitions**:
- **IV**: Verifies installation, setup, deployment, and configuration
- **OV**: Verifies functional behavior, business logic, and error handling (default)
- **PV**: Verifies performance requirements, response times, and resource usage

---

## Notes

[Any additional notes, context, or clarifications that don't fit elsewhere]

---

## Template Instructions

**To use this template**:

1. **Replace all placeholders** in angle brackets `<like this>` and square brackets `[like this]`
2. **Define Feature ID** following pattern: `<component>_<feature_name>` (e.g., `api_user_auth`, `ui_dashboard`, `service_notification`)
3. **Write user story** in "As a... I want... So that..." format
4. **Define 2-5 acceptance criteria** - each should be measurable and testable
5. **Assign verification types** to each AC:
   - Use **IV** for installation/setup/configuration scenarios
   - Use **OV** for functional behavior scenarios (most common)
   - Use **PV** for performance requirement scenarios
6. **Document Example Mapping** results from workshop (Yellow/Blue/Green cards)
7. **Link to BDD scenarios** using @ac1, @ac2, @ac3 tags in behavior.feature
8. **Add risk control information** (optional) if feature implements risk requirements:
   - Reference the risk control scenario in specs/risk-controls/
   - Tag BDD scenarios with @risk<ID> to create traceability
   - See: docs/how-to-guides/testing/link-risk-controls.md
9. **Remove unused sections** if not applicable to your feature

**Verification Type Guidelines**:
- **Installation Verification (IV)**: Tests that verify system is properly installed, configured, and ready to use
- **Operational Verification (OV)**: Tests that verify functional requirements and business logic work correctly
- **Performance Verification (PV)**: Tests that verify system meets performance, scalability, and resource requirements

These verification types are used to group test results in audit/regulatory documentation and implementation reports.
