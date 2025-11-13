# Compliance as Code Principles

## Introduction

"Compliance as Code" applies software engineering best practices to regulatory compliance. Instead of manual documentation and periodic reviews, organizations encode requirements as executable specifications, automate validation, and generate evidence as a delivery pipeline byproduct.

**Core insight**: Compliance requirements are specifications that can be tested, just like functional requirements.

---

## The Five Principles

Compliance as Code rests on five interconnected principles:

1. **Everything as Code** - All compliance artifacts in version control
2. **Continuous Validation** - Compliance checked on every commit
3. **Shift-Left Compliance** - Issues caught early when fixes are cheap
4. **Automated Evidence** - Evidence generated automatically from pipelines
5. **Executable Specifications** - Requirements expressed as automated tests

These principles work as a system - implementing only one or two provides limited value; implementing all five creates transformative change.

---

## Principle 1: Everything as Code

### What It Means

All compliance artifacts stored as version-controlled text files in Git. Instead of Word documents in SharePoint, everything lives in version control.

**In practice**:

- Policies in Markdown (or kept in existing document management system)
- **Requirements as Gherkin specifications** (MUST be in `specs/risk-controls/*.feature`)
- Procedures as Markdown with executable examples
- Evidence automatically referenced from version control

**Why It Matters**: Version control provides traceability, collaboration via pull requests, searchability, immutability, and automation capabilities impossible with traditional document management.

**See**: [Everything as Code](../everything-as-code/index.md) for detailed explanation of the paradigm and benefits.

---

## Principle 2: Continuous Validation

### What It Means

Compliance checked on every commit, not periodically. Compliance tests run in CI/CD pipeline alongside functional tests.

**In practice**:

- Risk control scenarios execute in CI (see specifications below)
- Security scans on every commit
- Policy violations fail builds
- Real-time compliance status visible in dashboards

**Why It Matters**: Continuous validation provides immediate feedback when violations occur, prevents compliance drift, scales without manual review overhead, and provides continuous audit readiness.

**See**: [CD Model Stages 1-6](../continuous-delivery/cd-model/cd-model-stages-1-6.md) for how continuous validation integrates into development stages.

---

## Principle 3: Shift-Left Compliance

### What It Means

Validate compliance as early as possible in the delivery lifecycle - ideally before code commits.

**Cost Differential**:

- **Pre-commit** (Stage 2): 5 minutes to fix
- **CI** (Stage 4): 15 minutes to fix
- **PLTE** (Stage 5): 1 hour to fix
- **Production** (Stage 11): Days to fix + incident response

**In practice**:

- Pre-commit hooks check for secrets, forbidden patterns
- CI validates all risk control scenarios
- PLTE runs acceptance tests
- Production monitoring detects drift

**See**: [Testing Strategy Overview](../continuous-delivery/testing/testing-strategy-overview.md) for comprehensive shift-left testing approach.

---

## Principle 4: Automated Evidence Collection

### What It Means

Evidence generated automatically as byproduct of pipeline execution. No manual evidence collection during audits.

**In practice**:

- Test results stored in Git LFS or artifact repository
- Deployment logs committed automatically
- Scan results referenced by commit SHA
- Traceability matrices generated from version control metadata

**Why It Matters**: Automated evidence eliminates manual collection overhead, ensures completeness (can't forget to collect), provides tamper-evident audit trail, and enables instant audit responses.

**Architecture**: Evidence collection integrates into CD Model at multiple stages (commit, PLTE, production).

---

## Principle 5: Executable Specifications

### What It Means

Compliance requirements expressed as Gherkin scenarios that can be executed as automated tests.

**Project Risk Controls** (`specs/risk-controls/auth-mfa.feature`):

```gherkin
# @industry:HEALTH @industry:PHARMA
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @hipaa @fda-21cfr11
# @implementation:required @automation:full

@risk-control:auth-mfa
Feature: Multi-Factor Authentication

  # Source: Risk Assessment RA-2025-001

  Rule: Authentication requires multiple factors

    @risk-control:auth-mfa-01
    Scenario: MFA required for all access
      Given a system with protected resources
      Then all user access MUST require at least two factors
      And authentication MUST occur before granting access
      And failed authentication attempts MUST be logged
```

**User Scenarios** link to risk controls via `@risk-control:<name>-<id>` tags:

```gherkin
Feature: cli_user-login

  Rule: Users must authenticate before accessing protected resources

    @success @ac1 @risk-control:auth-mfa-01
    Scenario: Valid credentials with MFA grant access
      Given I have valid credentials and MFA token
      When I run "r2r login --user admin --mfa"
      Then I should be authenticated
```

**Traceability Chain**: Regulatory requirement → Risk control specification (`@risk-control:auth-mfa-01`) → User scenarios (`@risk-control:auth-mfa-01` tag) → Step implementations → Production code

**See**:

- [Three-Layer Testing Approach](../specifications/three-layer-approach.md) - How ATDD/BDD/TDD work together
- [Risk Controls](../specifications/risk-controls.md) - Risk control specification pattern
- [Gherkin File Organization](../specifications/gherkin-concepts.md) - How to structure specifications

---

## The System Effect

These five principles create a virtuous cycle:

1. **Everything as Code** enables automation
2. **Continuous Validation** catches issues immediately
3. **Shift-Left** reduces fix costs
4. **Automated Evidence** eliminates audit overhead
5. **Executable Specifications** provide clear requirements

Together: **70-80% reduction in compliance overhead** while **improving compliance quality** and achieving **continuous audit readiness**.

---

## Implementation Sequence

Organizations typically implement in this order:

**Phase 1** (Pilot):

1. Start with Everything as Code (specs in Git)
2. Add Executable Specifications (risk controls in Gherkin)
3. Enable Continuous Validation (CI runs specs)

**Phase 2** (Scale):
4. Implement Shift-Left (pre-commit hooks)
5. Automate Evidence Collection (pipeline integration)

**Timeline**: 12-16 weeks for pilot, 8-12 weeks for automation tooling

**See**: [Transformation Framework](./transformation-framework.md) for detailed phased approach.

---

## Related Documentation

**Understanding the paradigm**:

- [Everything as Code](../everything-as-code/index.md) - Core principles and benefits

**Technical implementation**:

- [CD Model Overview](../continuous-delivery/cd-model/cd-model-overview.md) - Where compliance integrates
- [Testing Strategy](../continuous-delivery/testing/testing-strategy-overview.md) - Shift-left approach
- [Specifications](../specifications/index.md) - How to write executable specifications

**Transformation**:

- [Why Transform](./why-transformation.md) - Business case
- [Transformation Framework](./transformation-framework.md) - How to implement
