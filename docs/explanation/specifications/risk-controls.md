# Risk Controls in Executable Specifications

> **Understanding risk-based testing and compliance traceability**

## What Are Risk Controls?

Risk controls are **mitigation measures** that address identified risks in your system. They answer:

- **What could go wrong?** (Risk)
- **What must we do to prevent it?** (Control)
- **How do we prove it works?** (Evidence)

**Traditional approach**:

- Risk assessment → Spreadsheet of controls → Manual verification → Periodic audits
- Controls documented separately from implementation
- Evidence gathered retroactively
- Traceability reconstructed manually

**Executable risk controls approach**:

- Risk assessment → Gherkin control scenarios → Automated verification → Continuous evidence
- Controls are executable specifications
- Evidence generated automatically
- Traceability exists in real-time

---

## Why Executable Risk Controls?

### The Compliance Problem

Regulated industries (medical devices, financial services, aerospace, etc.) require:

1. **Risk assessment** - Identify what could go wrong
2. **Risk controls** - Define mitigation measures
3. **Verification** - Prove controls are implemented correctly
4. **Traceability** - Link risks → controls → implementation → evidence
5. **Audit trail** - Demonstrate compliance over time

**Traditional challenge**: These artifacts live in different systems:

- Risk assessment: Excel/SharePoint
- Control requirements: Word documents
- Implementation: Code repository
- Test evidence: Test management tools
- Audit trail: Manual compilation

**Result**: Weeks of effort to prepare for audits, high risk of missing traceability gaps.

### The Executable Solution

**Risk controls as Gherkin scenarios** in version control:

```gherkin
@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated
  And authentication MUST occur before granting access
  And failed authentication attempts MUST be logged
```

**Benefits**:

1. **Clear requirements**: Control requirements written in business language
2. **Executable verification**: Controls can be tested automatically
3. **Version controlled**: Changes tracked in Git with full history
4. **Reusable**: One control → many implementations
5. **Traceable**: @risk tags link controls to implementation
6. **Audit-ready**: Evidence exists continuously, not retroactively

---

## When You Need Risk Controls

### Regulated Domains

You likely need risk controls if you work in:

- **Medical devices** (FDA 21 CFR Part 820, ISO 13485, IEC 62304)
- **Pharmaceuticals** (FDA 21 CFR Part 11, GxP)
- **Financial services** (SOX, PCI-DSS, GDPR)
- **Aerospace** (DO-178C, DO-254)
- **Automotive** (ISO 26262, ASPICE)
- **Critical infrastructure** (IEC 61508, NERC CIP)
- **Data privacy** (GDPR, HIPAA, CCPA)

### Risk-Based Development

Even outside regulated domains, risk controls help when:

- **High consequence of failure** (safety-critical, financial loss, reputation damage)
- **Security requirements** (authentication, authorization, encryption, audit logging)
- **Compliance obligations** (SOC 2, ISO 27001, contractual requirements)
- **Audit requirements** (internal audits, external audits, certification)

### When You Don't Need Them

Skip risk controls for:

- **Low-risk internal tools** with no compliance requirements
- **Prototypes and experiments** not going to production
- **Simple utilities** with minimal consequences of failure
- **Open source side projects** without regulatory obligations

---

## Identifying Relevant Controls

### Step 1: Conduct Risk Assessment

**Identify risks** in your domain:

**Medical device example**:

- Risk: Incorrect dosage calculation could harm patient
- Control: System validates dosage against patient weight and drug database
- Severity: High | Probability: Medium → Risk Level: High

**Financial system example**:

- Risk: Unauthorized access could lead to fraudulent transactions
- Control: Multi-factor authentication required for all transactions above threshold
- Severity: High | Probability: Medium → Risk Level: High

**Methods**:

- Failure Mode and Effects Analysis (FMEA)
- Hazard Analysis
- Threat Modeling (STRIDE, PASTA)
- Compliance gap analysis

### Step 2: Define Control Requirements

For each high/medium risk, define **what must be true** to mitigate it:

**Authentication risk controls**:

- RC-001: All user access MUST be authenticated
- RC-002: Failed authentication attempts MUST be logged
- RC-003: Session tokens MUST expire after inactivity
- RC-004: Passwords MUST meet complexity requirements

**Data protection risk controls**:

- RC-010: Sensitive data MUST be encrypted at rest
- RC-011: Data transmission MUST use TLS 1.2+
- RC-012: Encryption keys MUST be rotated quarterly

### Step 3: Organize by Category

Group related controls:

```text
specs/risk-controls/
├── authentication-controls.feature      # RC-001 to RC-009
├── data-protection-controls.feature     # RC-010 to RC-019
├── audit-trail-controls.feature         # RC-020 to RC-029
├── input-validation-controls.feature    # RC-030 to RC-039
└── access-control.feature               # RC-040 to RC-049
```

### Step 4: Write as Gherkin Scenarios

**Control requirement** → **Gherkin scenario**

**From**: "All user access must be authenticated"

**To**:

```gherkin
@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated
  And authentication MUST occur before granting access
  And failed authentication attempts MUST be logged
```

**Key characteristics**:

- Use MUST/SHALL for mandatory requirements
- State what the system must do (not how)
- Be verifiable (can be tested)
- Reference source (assessment ID)

---

## The Traceability Chain

### From Risk to Evidence

```text
Risk Assessment
    ↓
Risk Control Definition (Gherkin scenario in specs/risk-controls/)
    ↓
Feature Implementation (User scenarios tagged with @risk<ID>)
    ↓
Test Execution (Automated via Godog)
    ↓
Test Evidence (Results linked to control via @risk tag)
    ↓
Audit Trail (Git history + test results)
```

### Example: Authentication Risk

**1. Risk identified** (from FMEA):

- ID: R-001
- Description: Unauthorized access to patient data
- Severity: High
- Mitigation: Require authentication for all access

**2. Control defined** (`specs/risk-controls/authentication-controls.feature`):

```gherkin
@risk1
Scenario: RC-001 - User authentication required
  Given a system with protected resources
  Then all user access MUST be authenticated
```

**3. Implementation** (`specs/cli/login/specification.feature`):

```gherkin
@success @ac1 @risk1
Scenario: Login with valid credentials
  Given I have valid credentials
  When I run "r2r login"
  Then I should be authenticated
```

**4. Execution**:

```bash
godog specs/**/specification.feature
# Scenario: Login with valid credentials ... passed
```

**5. Evidence**:

- Scenario tagged @risk1 passed
- Control RC-001 verified
- Risk R-001 mitigated
- Git commit SHA provides version traceability

**6. Audit query**:

```bash
# "Show me all scenarios that verify authentication control"
grep -r "@risk1" specs/ --exclude-dir=risk-controls
```

---

## Common Compliance Frameworks

### FDA 21 CFR Part 11 (Pharmaceuticals)

**Key requirements**:

- Electronic signature validation
- Audit trails
- System validation
- Access controls

**Risk control mapping**:

- RC-001: User authentication
- RC-005: Audit trail completeness
- RC-020: Electronic signature verification

### ISO 13485 / IEC 62304 (Medical Devices)

**Key requirements**:

- Software safety classification
- Risk management
- Verification and validation
- Traceability

**Risk control mapping**:

- RC-030: Input validation for safety-critical parameters
- RC-040: Output verification for medical calculations

### PCI-DSS (Payment Card Industry)

**Key requirements**:

- Protect cardholder data
- Maintain secure systems
- Access control
- Monitor and test networks

**Risk control mapping**:

- RC-010: Encrypt data at rest
- RC-011: Encrypt data in transit
- RC-001: Multi-factor authentication

### GDPR (Data Privacy)

**Key requirements**:

- Data protection by design
- Right to be forgotten
- Consent management
- Data breach notification

**Risk control mapping**:

- RC-050: User consent must be explicit and documented
- RC-051: Personal data must be deletable on request
- RC-052: Data breaches must be logged and reported

---

## Best Practices

### Do

✅ **Start with risk assessment** - Don't create controls without identified risks

✅ **Use clear IDs** - Sequential (RC-001, RC-002) or categorical (RC-AUTH-001)

✅ **Reference source** - Link to assessment document/ID

✅ **Use MUST/SHALL** - Make mandatory requirements explicit

✅ **Keep atomic** - One control requirement per scenario

✅ **Review regularly** - Update controls when risks change

### Don't

❌ **Don't create controls "just in case"** - Only for identified risks

❌ **Don't make controls too specific** - They should be implementation-agnostic

❌ **Don't skip traceability** - Always tag implementation scenarios

❌ **Don't forget to execute** - Controls are worthless if not verified

❌ **Don't duplicate** - Reuse controls across features

---

## Related Documentation

- [Link Risk Controls (How-To)](../../how-to-guides/specifications/link-risk-controls.md) - Step-by-step implementation
- [Gherkin Format Reference](../../reference/specifications/gherkin-format.md) - Tag syntax
- [Three-Layer Testing](./three-layer-approach.md) - How risk controls fit in ATDD/BDD/TDD
