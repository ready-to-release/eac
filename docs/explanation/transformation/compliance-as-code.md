# Compliance as Code Principles

## Introduction

"Compliance as Code" applies software engineering practices to regulatory compliance.

Organizations encode requirements as executable specifications, automate validation, and generate evidence as a delivery pipeline byproduct.

**Core insight**: Compliance requirements are specifications that can be tested, just like functional requirements.

---

## The Five Principles

Compliance as Code rests on five interconnected principles:

1. **Everything as Code** - All compliance artifacts in version control
2. **Continuous Validation** - Compliance checked on every commit
3. **Shift-Left Compliance** - Issues caught early when fixes are cheap
4. **Automated Evidence** - Evidence generated automatically from pipelines
5. **Executable Specifications** - Requirements expressed as automated tests

These principles work as a system - implementing only one or two provides limited value;
implementing all five creates transformative change.

---

## Principle 1: Everything as Code

### What It Means

All compliance artifacts stored as version-controlled text files in Git instead of Word documents in SharePoint.

**In practice**:

- Policies in Markdown (or kept in existing document management system)
- Procedures as Markdown with executable examples
- Evidence automatically referenced from version control

!!! note "Requirements Storage"
    Requirements can be stored as Gherkin specifications

**Why It Matters**: Version control provides traceability, collaboration via pull requests,
searchability, immutability, and automation capabilities impossible with traditional document management.

**See**: [Everything as Code](../everything-as-code/index.md) for detailed explanation of the paradigm and benefits.

---

## Principle 2: Continuous Validation

### What It Means

Compliance checked on every commit. Compliance tests run in CI/CD pipeline alongside functional tests.

**In practice**:

- Risk control scenarios execute in CI (see specifications below)
- Security scans on every commit
- Policy violations fail builds
- Real-time compliance status visible in dashboards

**Why It Matters**: Continuous validation provides immediate feedback when violations occur,
prevents compliance drift, scales without manual review overhead, and provides continuous audit readiness.

**See**: [CD Model Stages 1-7](../continuous-delivery/cd-model/stages.md#development-stages) for how continuous validation integrates into development stages.

---

## Principle 3: Shift-Left Compliance

### What It Means

Validate compliance as early as possible in the delivery lifecycle - ideally before code commits.

!!! tip "Shift-Left Economics"
    The earlier you catch issues, the cheaper they are to fix. Issues found at later stages require:

    - More coordination across teams
    - Incident response procedures
    - Potential regulatory reporting
    - Customer communication
    - Significantly longer resolution time

**In practice**:

- Pre-commit hooks check for secrets, forbidden patterns
- CI validates all risk control scenarios
- PLTE runs acceptance tests
- Production monitoring detects drift

**See**: [Testing Strategy](../continuous-delivery/testing/index.md) for comprehensive shift-left testing approach.

---

## Principle 4: Automated Evidence Collection

### What It Means

Evidence generated automatically as byproduct of pipeline execution - no manual collection during audits.

**In practice**:

- Test results stored in Git LFS or artifact repository
- Deployment logs committed automatically
- Scan results referenced by commit SHA
- Traceability matrices generated from version control metadata

**Why It Matters**: Automated evidence eliminates manual collection overhead, ensures completeness, provides tamper-evident audit trail, and enables instant audit responses.

**Architecture**: Evidence collection integrates into CD Model at multiple stages (commit, PLTE, production).

---

## Principle 5: Executable Specifications

### What It Means

User scenarios expressed as executable Gherkin specifications, linked to risk controls that define compliance requirements.

### The Risk Control Architecture

**Three Layers**:

1. **Risk Catalog**
   - Library of all available risk controls
   - You don't modify this - it's your control library
   - Example controls: AC-2 (Account Management), IA-2 (Authentication), AU-2 (Audit Events)

2. **Risk Profile**
   - Selects which controls from the catalog apply to YOUR solution
   - Single repository-wide profile
   - Based on your risk assessment

3. **Specifications**
   - User scenarios tagged with `@control:<id>` to link to profile controls
   - Tests that verify control satisfaction
   - Evidence that proves compliance

**Implementation Note**: The catalog and profile use OSCAL (Open Security Controls Assessment Language) for machine-readable compliance.

---

## Related Documentation

**Understanding the paradigm**:

- [Everything as Code](../everything-as-code/index.md) - Core principles and benefits

**Technical implementation**:

- [CD Model Overview](../continuous-delivery/cd-model/overview.md) - Where compliance integrates
- [Testing Strategy](../continuous-delivery/testing/index.md) - Shift-left approach
- [Specifications](../specifications/index.md) - How to write executable specifications

**Transformation**:

- [Why Transform](./why-transformation.md) - Business case
- [Transformation Framework](./transformation-framework.md) - How to implement
