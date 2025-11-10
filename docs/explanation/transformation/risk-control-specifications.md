# Risk Control Specifications

## Introduction

Risk control specifications are the bridge between regulatory requirements and automated compliance validation. They express "what must be true" about a system in a format that's both human-readable for compliance officers and executable by automated tests.

This document explains the risk control specification pattern, why Gherkin format is used, and how to create effective specifications that enable automated compliance validation.

**What you'll learn**: You'll understand the pattern for expressing compliance requirements as executable specifications, see examples across different regulatory frameworks, and learn how to implement traceability from requirement to evidence.

---

## The Pattern

Risk control specifications follow a clear pattern with two layers:

### Risk Control (The Requirement)

Defines what must be true from a regulatory perspective:

```gherkin
@risk1
Scenario: RC-001 - Authentication required for all access
  Given a system with protected resources
  Then all user access MUST be authenticated
  And authentication MUST occur before granting access
  And failed authentication attempts MUST be logged
```

**Key characteristics**:

- Uses `@risk<N>` tag for traceability (e.g., @risk1, @risk2)
- Written in business language, not technical implementation
- Uses RFC 2119 language (MUST, SHOULD, MAY)
- Focuses on observable requirements, not how to implement them

### Implementation Specification (The Test)

Proves the requirement is met through executable scenarios:

```gherkin
@success @ac1 @risk1
Scenario: Valid credentials grant access
  Given I have valid credentials
  When I attempt to access protected resource
  Then I should be authenticated
  And access should be logged
```

**Key characteristics**:

- Links to risk control via `@risk1` tag
- Written from user perspective
- Provides concrete, testable example
- Can be executed as automated test

### Traceability

The `@risk<N>` tag creates automatic linkage:

- **One requirement → Multiple implementations**: A single risk control may require multiple test scenarios to prove full compliance
- **Multiple requirements → One implementation**: A single test may satisfy multiple risk controls
- **Automated matrix generation**: Tools parse tags to generate traceability matrices

---

## Why Gherkin Format?

### Human-Readable

Gherkin's Given-When-Then format is accessible to non-technical stakeholders:

- **Compliance officers** can read and approve specifications without technical knowledge
- **Auditors** can understand requirements and validation without reviewing code
- **Business stakeholders** can validate that requirements match business needs
- **Regulators** can map specifications to regulatory text

This accessibility enables collaboration between technical and non-technical teams.

### Executable

Gherkin specifications become automated tests:

- **Step definitions** implement the test logic
- **Test runners** (Godog, Cucumber) execute scenarios in CI/CD pipelines
- **Results** prove compliance automatically on every commit
- **Evidence** generated from test execution

This executability ensures specifications stay synchronized with actual system behavior—specifications cannot drift from reality because they're continuously validated.

### Version-Controlled

Specifications live in Git alongside application code:

- **Change tracking**: Every modification tracked with who, when, why
- **Pull request workflow**: Changes reviewed and approved like code
- **Audit trail**: Complete history from Git commits
- **Collaboration**: Teams work together through version control

This provides stronger audit trails than traditional document management systems.

### Language-Agnostic

Gherkin separates specification from implementation:

- **Same specification, different implementation**: Can use Godog (Go), Cucumber (Ruby/JS/Java), SpecFlow (.NET)
- **Stable specifications**: Implementation can change without rewriting specifications
- **Platform independent**: Works across languages and frameworks

---

## Structure of Risk Control Specifications

### Directory Organization

Risk controls are stored separately from implementation specifications:

```text
specs/
└── risk-controls/
    ├── authentication-controls.feature
    ├── data-protection-controls.feature
    ├── audit-trail-controls.feature
    └── [category]-controls.feature
```

This separation allows:

- Risk controls to evolve independently from implementations
- Multiple implementations to reference same controls
- Clear ownership (compliance office owns risk controls, teams own implementations)

### Feature File Format

```gherkin
Feature: Authentication Controls
  # Compliance risk controls for user authentication and access management.
  #
  # Source:
  #   - ISO 27001:2022 A.8.5 (Secure authentication)
  #   - GDPR Article 32 (Security of processing)
  #
  # Assessment: RISK-ASSESS-2025-01
  # Date: 2025-01-15

  @risk1
  Scenario: RC-001 - Authentication required for all access
    Given a system with protected resources
    Then all user access MUST be authenticated
    And authentication MUST occur before granting access
    And failed authentication attempts MUST be logged

  @risk2
  Scenario: RC-002 - Multi-factor authentication for privileged access
    Given a system with privileged administrative functions
    Then privileged access MUST use multi-factor authentication
    And MFA enforcement MUST be logged
    And MFA failures MUST trigger security alerts
```

**Key elements**:

- **Feature description**: High-level category of controls
- **Source attribution**: Specific regulatory references
- **Assessment ID**: Links to risk assessment documentation
- **Date**: When controls were defined or reviewed
- **Scenarios**: Individual control requirements

### Tagging Convention

**Sequential numbering**: @risk1, @risk2, @risk3...

- Simple and clear
- Easy to reference in conversation
- Works well for smaller sets (<100 controls)

**Categorical numbering**: @risk10-19 (authentication), @risk20-29 (data protection)

- Groups related controls
- Scales to larger control sets
- Easier to understand control domain from tag

**One tag per requirement**: Each risk control gets exactly one @risk tag for clear 1:1 traceability.

### Source Attribution

Always document where requirements originate:

**ISO 27001 Example**:

```text
Source: ISO 27001:2022 A.9.2.1 (Access control policy)
```

**GDPR Example**:

```text
Source: GDPR Article 32 (Security of processing)
```

**FDA 21 CFR Part 11 Example**:

```text
Source: FDA 21 CFR Part 11 §11.10(a) (Validation of systems)
```

**EMA Annex 11 Example**:

```text
Source: EMA Annex 11 Section 4.8 (Data)
```

**Starting Point**: Use existing SOPs if available, otherwise consult regulations directly. SOPs provide organization-specific context; regulations provide baseline requirements.

---

## Implementation Specifications

### Where They Live

Implementation specifications live with the code they test:

```text
specs/
└── [module]/
    └── [feature]/
        └── specification.feature
```

Example:

```text
specs/cli/user-authentication/specification.feature
```

This co-location ensures:

- Specifications evolve with features
- Clear ownership by feature teams
- Easy to find specifications for a feature

### Linking to Risk Controls

Link implementation scenarios to risk controls using @risk tags:

```gherkin
@cli @critical @security
Feature: cli_user-authentication

  As a system administrator
  I want secure user authentication
  So that only authorized users can access the system

  Rule: All access requires valid authentication

    @success @ac1 @risk1
    Scenario: Valid credentials grant access
      Given I have valid credentials
      When I run "r2r login --user admin"
      Then I should be authenticated
      And my session should be active
      And my login should be logged

    @error @ac1 @risk1
    Scenario: Invalid credentials deny access
      Given I have invalid credentials
      When I run "r2r login --user admin"
      Then I should not be authenticated
      And the failure should be logged
```

The `@risk1` tag creates the link from this implementation to RC-001 (Authentication required).

---

## Validation and Coverage

### Validation

Risk control specifications should be validated for:

**Format**: Proper Gherkin syntax

- Scenarios use Given-When-Then structure
- Tags follow convention (@risk[N])
- RFC 2119 language used correctly (MUST/SHOULD/MAY)

**Uniqueness**: No duplicate @risk IDs

- Each control has exactly one tag
- Tags are sequential or follow naming scheme
- No conflicts or reuse

**Attribution**: Source standard documented

- Every scenario references specific regulation
- Section numbers included
- Assessment ID links to risk assessment

**Coverage**: All requirements have implementations

- Every @risk tag has at least one implementing scenario
- Traceability matrix shows complete coverage
- No orphan requirements without tests

**Automation Note**: Validation requires tooling. Ready-to-Release (r2r) CLI tries to help with validation commands.

---

## Best Practices

### ✅ Do

1. **Atomic Requirements**: One clear requirement per scenario
   - Good: "Authentication MUST be required"
   - Bad: "System must be secure" (too vague)

2. **Clear Language**: Use RFC 2119 keywords consistently
   - MUST: Absolute requirement
   - SHOULD: Recommended but not mandatory
   - MAY: Optional

3. **Testable**: Requirements must be verifiable
   - Good: "Response time MUST be under 2 seconds"
   - Bad: "System should be fast" (not measurable)

4. **Source Attribution**: Always document regulatory source
   - Include standard name and section number
   - Links risk control to compliance obligation

5. **Version Control**: Track changes in Git
   - Pull request workflow for changes
   - Compliance officer approval required
   - Complete audit trail from Git history

6. **Review Process**: Compliance officer approval via PR
   - Risk controls reviewed before merge
   - Changes require compliance sign-off
   - No direct commits to main branch

7. **Tagging Discipline**: Consistent @risk ID usage
   - Sequential or categorical numbering
   - One tag per requirement
   - No duplicate or reused tags

### ❌ Don't

1. **Don't mix implementation details in risk controls**
   - Risk control: "Data MUST be encrypted"
   - NOT: "Data MUST be encrypted using AES-256" (implementation choice)

2. **Don't create untestable requirements**
   - Bad: "System should be user-friendly" (subjective)
   - Good: "Help documentation MUST be provided for all commands"

3. **Don't skip source attribution**
   - Every requirement needs regulatory reference
   - Auditors need to trace back to source

4. **Don't create duplicate controls**
   - Review existing controls before adding new
   - Use existing @risk tags when applicable

5. **Don't bypass review process**
   - All risk control changes need compliance approval
   - No emergency bypasses

---

## Integration with Existing Documentation

Risk control specifications integrate with the broader testing approach:

- **Specifications**: [Three-Layer Testing Approach](../specifications/three-layer-approach.md) explains how ATDD/BDD/TDD work together
- **Testing Strategy**: [Testing Strategy Overview](../continuous-delivery/testing/testing-strategy-overview.md) shows where risk controls fit in overall testing

---

## Next Steps

To implement risk control specifications:

1. **Learn shift-left** - Read [Shift-Left Compliance](shift-left-compliance.md) to understand early validation
2. **Review templates** - Check [Template Catalog](https://github.com/ready-to-release/eac/blob/main/templates/index.md) for example patterns

---

## Related Documentation

- [Compliance as Code](compliance-as-code.md) - Core principles including executable specifications
- [Shift-Left Compliance](shift-left-compliance.md) - When and where to validate
- [Transformation Framework](transformation-framework.md) - How to implement organization-wide
- [Template Catalog](https://github.com/ready-to-release/eac/blob/main/templates/index.md) - Example templates
