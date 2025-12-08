# GxP Tagging

Understanding tagging for regulated software development in pharmaceutical and medical device contexts.

---

## Overview

When developing software for GxP-regulated environments (Good Manufacturing Practice, Good Clinical Practice, Good Laboratory Practice), additional tagging is required to ensure traceability, compliance, and proper risk management.

This document describes the **regulatory tagging taxonomy** used alongside the standard testing taxonomy to support:

- Pharmaceutical manufacturing (GMP)
- Clinical trials (GCP)
- Laboratory operations (GLP)
- Medical device software (ISO 13485, FDA 21 CFR Part 11)

**Regulatory Tags**:

- **Requirement Classification** - GxP classification (`@gxp`, `@gmp-critical-aspect`)
- **Risk Tracking** - Link to GxP risks (`@risk:<risk-id>`)
- **Compliance Controls** - Link to OSCAL controls (`@control:<id>`)

---

## Specification Hierarchy: URS → FS → DS

In regulated software development, specifications follow a three-level hierarchy:

### User Requirement Specification (URS)

**Purpose**: Highest-level specifications defining **what** the system must do from the user's perspective

**Focus**: Intended use, business requirements, regulatory requirements

**Format**: Feature files (Gherkin or Markdown) where the **feature name itself** serves as the URS identifier

**Naming Standard**: `<module>_<feature-name>` (e.g., `auth_user-authentication`)

**Example**:

```gherkin
@gxp
Feature: auth_user-authentication
  As a system administrator
  I want to control user access to GxP-critical functions
  So that only authorized personnel can perform regulated activities
```

**Approval**: Requires QA and System Owner approval

### Functional Specification (FS)

**Purpose**: Defines **how** the system behaves functionally in response to inputs

**Focus**: Observable behavior, scenarios, acceptance criteria

**Format**: Scenarios within feature files using Given-When-Then format

**Example**:

```gherkin
  Rule: Only authorized users can access critical functions

    @ov @gxp @risk-control:gxp-access-control
    Scenario: Unauthorized user cannot access critical function
      Given I am logged in as a standard user
      When I attempt to access the "Batch Release" function
      Then access is denied
      And I see "Insufficient privileges"
```

**Approval**: Derived from URS, approved via pull request review

### Design Specification (DS)

**Purpose**: Technical architecture and implementation details

**Focus**: System architecture, technology choices, data structures, APIs

**Format**: Standalone documentation (architecture diagrams, technical specifications)

**Example**: Solution Design Document describing database schema, API endpoints, deployment architecture

**Approval**: Reviewed by technical peers, approved by System Owner

### Relationship and Traceability

```text
URS (User Requirements)
  ↓ "satisfies"
 FS (Functional Specification - Scenarios)
  ↓ "implements"
DS (Design Specification - Architecture)
  ↓ "builds"
Code Implementation
  ↓ "verifies"
Test Results
```

**Traceability Chain**: Every requirement in URS must be traceable through FS scenarios to test results, ensuring complete verification.

---

## Regulatory Classification Tags

### Feature Naming as URS Identifier

**Purpose**: The feature name itself serves as the User Requirement Specification (URS) identifier

**Naming Standard**: `<module>_<feature-name>` format

**Examples**:

```gherkin
@gxp
Feature: auth_user-authentication
  As a system administrator...

@gxp
Feature: batch_release-workflow
  As a quality manager...
```

**Requirements**:

- Feature name must follow `<module>_<feature-name>` naming convention
- Name should be descriptive and unique within the system
- Used for traceability in regulatory documentation
- No separate `@URS:NAME` tag needed

---

### `@gxp` - GxP-Related Requirement

**Purpose**: Identify requirements related to GxP-controlled aspects of the software

**Usage**: Feature or scenario level

**Scope**: Any requirement that:

- Affects product quality
- Impacts patient safety
- Controls data integrity
- Supports GxP-regulated processes

**Example**:

```gherkin
@gxp
Feature: audit_trail-gxp-operations

  Rule: All GxP-critical actions must be logged

    @ov @gxp @risk:audit-trail-failure @control:au-2 @control:au-3
    Scenario: System records user action with timestamp
      Given I am logged in as a production operator
      When I approve a batch for release
      Then the action is logged with username, timestamp, and batch ID
      And the audit log cannot be modified
```

**Requirements**:

- All `@gxp` requirements **should** link to risk tracking (`@risk:<risk-id>`) if project-specific risks are identified
- All `@gxp` requirements **should** link to compliance controls (`@control:<id>`) for standardized controls
- Must include both positive and negative test scenarios (challenge tests)
- Requires approval from QA and System Owner

**Risk and Control Linkage**: When tagging with `@gxp`, you should:

1. Identify project risks and tag with `@risk:<risk-id>` (e.g., `@risk:audit-trail-failure`)
2. Link to applicable OSCAL controls with `@control:<id>` (e.g., `@control:au-2` for audit events)
3. Document risks in risk assessment documentation
4. Classify risk as High/Medium/Low

---

### `@gmp-critical-aspect` - GmP Critical Aspect

**Purpose**: Mark requirements as Critical Aspects (CA) for GmP (Good Manufacturing Practice) products only

**Usage**: Scenario level (only with `@gxp`)

**Scope**: Functions, features, or characteristics that ensure:

- Consistent product quality
- Patient safety
- Data integrity in manufacturing

**Important**: Only use for **GmP classified products** (not for GCP or GLP)

**Example**:

```gherkin
@gxp
Feature: batch_release-quality-control

  Rule: Batch release requires quality approval

    @ov @gxp @gmp-critical-aspect @risk:unapproved-release @control:ac-2
    Scenario: Quality manager approves batch meeting specifications
      Given a production batch has completed all quality tests
      And all test results meet specifications
      When the Quality Manager reviews the batch
      And approves the batch for release
      Then the batch status changes to "Approved for Release"
      And the approval is recorded in the audit trail
      And the batch cannot be modified after approval
```

**Validation Deviation**: If a test tagged with `@gmp-critical-aspect` fails after production deployment, it must be managed as a validation deviation per regulatory requirements.

**Requirements**:

- Always used together with `@gxp`
- Should link to risk tracking with `@risk:<risk-id>` if project-specific risks identified
- Should link to compliance controls with `@control:<id>` for applicable standards
- Requires enhanced testing (negative tests, challenge tests, boundary conditions)
- Failure triggers validation deviation process

---

## Risk and Control Tags for GxP

### `@risk:<risk-id>` - GxP Risk Tracking

**Purpose**: Link GxP scenarios to project-specific risks from risk assessments

**Usage**: Scenario level (recommended for all `@gxp` requirements with identified risks)

**Format**: `@risk:<risk-id>` in kebab-case (e.g., `@risk:brute-force-attack`)

**Example**:

```gherkin
@gxp
Feature: auth_user-authentication

  Rule: Failed login attempts must be monitored

    @ov @gxp @risk:brute-force-attack @control:ac-7
    Scenario: System locks account after 5 failed attempts
      Given I have a valid user account
      When I enter an incorrect password 5 times
      Then my account is locked
      And an administrator must unlock it
      And all failed attempts are logged
```

**Risk Documentation** (separate risk assessment document):

Risk assessments should be documented in dedicated risk management artifacts:

```markdown
# Risk Assessment: Authentication Security

**Risk ID**: brute-force-attack
**Source**: RA-2025-AUTH-001
**Description**: Unauthorized access due to brute force password attacks
**Likelihood**: Possible (30-70%)
**Impact**: Critical (Patient Safety / Data Integrity)
**Gross Risk**: High
**Net Risk**: Medium (with controls)

**Controls**:
- Account lockout after failed attempts (@control:ac-7)
- Audit logging of authentication events (@control:au-2, @control:au-3)
- Password complexity requirements (@control:ia-5(1))
```

### `@control:<id>` - OSCAL Compliance Controls

**Purpose**: Link GxP scenarios to standardized compliance controls

**Usage**: Scenario level (required for regulatory compliance)

**Format**: `@control:<family>-<number>` or `@control:<family>-<number>(<enhancement>)`

**Examples**: `@control:ac-2`, `@control:au-3`, `@control:ia-5(1)`

**Common GxP Controls**:

| Control | Description | GxP Relevance |
|---------|-------------|---------------|
| `@control:ac-2` | Account Management | User access control |
| `@control:au-2` | Audit Events | GxP audit trail requirements |
| `@control:au-3` | Audit Record Content | FDA 21 CFR Part 11 compliance |
| `@control:ia-5` | Authenticator Management | Identity verification |
| `@control:si-7` | Software Integrity | Software validation |

**Requirements**:

- Use `@risk:<risk-id>` for project-specific risks from risk assessments
- Use `@control:<id>` for standardized compliance controls
- Risk assessments should be documented separately (not as feature files)
- Link test scenarios to both risks and controls as appropriate

**See Also**: [Risk Controls](risk-controls.md) for complete documentation on risk and control tagging

---

## Integration with Testing Taxonomy

Regulatory tags work alongside standard testing taxonomy tags:

```gherkin
@gxp @L2 @deps:ldap
Feature: auth_user-authentication-ldap

  Rule: Authentication validates against corporate LDAP

    @ov @gxp @risk:ldap-auth-failure @control:ia-2 @control:au-2
    Scenario: Valid LDAP credentials grant access
      Given the LDAP server is available
      When I login with valid corporate credentials
      Then authentication succeeds
      And my session is established
      And login is recorded in audit trail
```

**Tag Types Present**:

- **Regulatory**: `@gxp`
- **Risk Tracking**: `@risk:ldap-auth-failure`
- **Compliance Controls**: `@control:ia-2`, `@control:au-2`
- **Testing Taxonomy**: `@L2`, `@ov`, `@deps:ldap`
- **Feature Name**: `auth_user-authentication-ldap` (serves as URS identifier)

---

## Best Practices

### Regulatory Tagging

✅ **DO**:

- Use feature naming standard `<module>_<feature-name>` for URS identification
- Add `@gxp` for any requirement affecting regulated processes
- Link to project risks with `@risk:<risk-id>` when specific risks are identified
- Link to compliance controls with `@control:<id>` for standardized requirements
- Use `@gmp-critical-aspect` only for GmP products
- Document risk assessments separately in dedicated risk management artifacts
- Maintain traceability from URS → FS → DS → Code → Tests
- Use lowercase for all tags

❌ **DON'T**:

- Use `@gmp-critical-aspect` for non-GmP products (GCP, GLP)
- Tag with `@gxp` without linking to appropriate risks and controls
- Omit negative/challenge tests for `@gxp` requirements
- Use separate `@URS:NAME` tag (feature name serves as identifier)
- Use deprecated `@risk-control:gxp-*` format (use `@risk:` and `@control:` instead)

---

## Traceability and Reporting

### Implementation Report Contents

At release approval, regulatory tags enable automatic generation of:

**Requirements Specifications (URS/FS)**:

- All features following `<module>_<feature-name>` naming convention
- Features tagged with `@gxp`
- Scenarios with acceptance criteria
- Linked risk control specifications

**Test Summary**:

- Installation Verification (IV) - scenarios tagged with `@iv`
- Operational Verification (OV) - scenarios tagged with `@ov`
- Performance Verification (PV) - scenarios tagged with `@pv`
- Manual test results for `@Manual` scenarios (with git-stored evidence)

**Risk Traceability Matrix**:

- Feature (URS) → `@gxp` scenarios → `@risk:<risk-id>` tags → Risk assessments
- Feature (URS) → `@gxp` scenarios → `@control:<id>` tags → OSCAL controls → Test results

---

## Related Documentation

**Testing Taxonomy**:

- [Tag Reference](tag-reference.md) - Core testing taxonomy tags
- [Three-Layer Approach](three-layer-approach.md) - Rule/Scenario/Unit Test integration

**Specifications**:

- [Gherkin File Organization](gherkin-concepts.md) - Organizing specifications
- [Risk Controls](risk-controls.md) - Risk control specifications
