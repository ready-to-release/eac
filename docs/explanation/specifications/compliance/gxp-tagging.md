# GxP Tagging
>
> **Understanding tagging for regulated software development in pharmaceutical and medical device contexts.**

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
- **Compliance Controls** - Link to risk controls (`@control:<id>`)

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

    @ov @gxp @control:ac-3
    Scenario: Unauthorized user cannot access critical function
      # Mitigates: Unauthorized access to critical GxP functions
      # See: docs/risk-assessment/ra-2025-access-control.md
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

    @ov @gxp @control:au-2 @control:au-3
    Scenario: System records user action with timestamp
      # Mitigates: Audit trail integrity risk
      # See: docs/risk-assessment/ra-2025-audit-trail.md
      Given I am logged in as a production operator
      When I approve a batch for release
      Then the action is logged with username, timestamp, and batch ID
      And the audit log cannot be modified
```

**Requirements**:

- All `@gxp` requirements **must** link to applicable compliance controls (`@control:<id>`)
- Must include both positive and negative test scenarios (challenge tests)
- Requires approval from QA and System Owner
- Risk assessments documented separately and mapped to controls

**Control Linkage**: When tagging with `@gxp`, you should:

1. Identify applicable risk controls for the GxP requirement
2. Tag scenarios with `@control:<id>` (e.g., `@control:au-2` for audit events)
3. Document risks in separate risk assessment artifacts
4. Map risks to controls in risk documentation
5. Use scenario comments to reference risk assessments

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

    @ov @gxp @gmp-critical-aspect @control:ac-2
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
- Must link to applicable compliance controls with `@control:<id>`
- Requires enhanced testing (negative tests, challenge tests, boundary conditions)
- Failure triggers validation deviation process
- Risks documented in external risk assessment artifacts

---

## Risk Control Tags for GxP

### `@control:<id>` - GxP Compliance Controls

**Purpose**: Link GxP scenarios to standardized compliance controls

**Usage**: Scenario level (required for regulatory compliance)

**Format**: `@control:<family>-<number>` or `@control:<family>-<number>(<enhancement>)`

**Examples**: `@control:ac-2`, `@control:au-3`, `@control:ia-5(1)`

**Example**:

```gherkin
@gxp
Feature: auth_user-authentication

  Rule: Failed login attempts must be monitored

    @ov @gxp @control:ac-7
    Scenario: System locks account after 5 failed attempts
      # Mitigates: Brute force attack risk
      # See: docs/risk-assessment/ra-2025-auth-security.md
      Given I have a valid user account
      When I enter an incorrect password 5 times
      Then my account is locked
      And an administrator must unlock it
      And all failed attempts are logged
```

**Risk Documentation** (external artifacts):

Risk assessments should be documented separately and map risks to controls:

```markdown
# Risk Assessment: Authentication Security (RA-2025-AUTH-001)

**Risk**: Brute force password attacks
**Likelihood**: Possible (30-70%)
**Impact**: Critical (Patient Safety / Data Integrity)
**Gross Risk**: High | **Net Risk**: Medium (with controls)

**Mitigating Controls**:
- AC-7: Unsuccessful Logon Attempts (account lockout)
- AU-2: Audit Events (authentication logging)
- IA-5(1): Password-Based Authentication (complexity)

**Verification**: See specs/auth-service/*.feature scenarios
```

**Common GxP Controls**:

| Control         | Description              | GxP Relevance                 |
| --------------- | ------------------------ | ----------------------------- |
| `@control:ac-2` | Account Management       | User access control           |
| `@control:au-2` | Audit Events             | GxP audit trail requirements  |
| `@control:au-3` | Audit Record Content     | FDA 21 CFR Part 11 compliance |
| `@control:ia-5` | Authenticator Management | Identity verification         |
| `@control:si-7` | Software Integrity       | Software validation           |

**Requirements**:

- Use `@control:<id>` for linking scenarios to compliance controls
- Document risks in external risk assessment artifacts
- Map risks to controls in risk documentation
- Use scenario comments to reference risk assessments
- Controls inherently address risks per catalog definitions

**See Also**: [Risk Control Traceability](risk-controls.md) for complete documentation

---

## Integration with Testing Taxonomy

Regulatory tags work alongside standard testing taxonomy tags:

```gherkin
@gxp @L2 @deps:ldap
Feature: auth_user-authentication-ldap

  Rule: Authentication validates against corporate LDAP

    @ov @gxp @control:ia-2 @control:au-2
    Scenario: Valid LDAP credentials grant access
      # Mitigates: LDAP authentication failure risk
      # See: docs/risk-assessment/ra-2025-ldap-auth.md
      Given the LDAP server is available
      When I login with valid corporate credentials
      Then authentication succeeds
      And my session is established
      And login is recorded in audit trail
```

**Tag Types Present**:

- **Regulatory**: `@gxp`
- **Compliance Controls**: `@control:ia-2`, `@control:au-2`
- **Testing Taxonomy**: `@L2`, `@ov`, `@deps:ldap`
- **Feature Name**: `auth_user-authentication-ldap` (serves as URS identifier)

---

## Best Practices

### Regulatory Tagging

✅ **DO**:

- Use feature naming standard `<module>_<feature-name>` for URS identification
- Add `@gxp` for any requirement affecting regulated processes
- Link to compliance controls with `@control:<id>` for applicable requirements
- Use `@gmp-critical-aspect` only for GmP products
- Document risk assessments separately in dedicated risk management artifacts
- Use scenario comments to reference risk assessment documents
- Maintain traceability from URS → FS → DS → Code → Tests
- Use lowercase for all tags

❌ **DON'T**:

- Use `@gmp-critical-aspect` for non-GmP products (GCP, GLP)
- Tag with `@gxp` without linking to appropriate controls
- Omit negative/challenge tests for `@gxp` requirements
- Use separate `@URS:NAME` tag (feature name serves as identifier)

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

- Risk Assessment (external) → identifies risks → maps to risk controls
- Feature (URS) → `@gxp` scenarios → `@control:<id>` tags → risk controls → Test results
- Test Results → Evidence → Assessment Results → Audit compliance

---

## Related Documentation

**Testing Taxonomy**:

- [Testing Taxonomy](../taxonomy/) - Core testing taxonomy tags
- [Three-Layer Approach](../concepts/three-layer-approach.md) - Rule/Scenario/Unit Test integration

**Specifications**:

- [File Structure](../organization/file-structure.md) - Organizing specifications
- [Risk Controls](risk-controls.md) - Risk control specifications
