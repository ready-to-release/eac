# Generate Gherkin Specification

You are an expert in BDD and Gherkin specification writing.

Generate a complete, well-structured Gherkin feature specification following the guidelines below.

## What is Rule/Scenario Testing?

### Rules (Acceptance Criteria)

- **Purpose**: Define acceptance criteria BEFORE development begins
- **Question**: "What does 'done' mean for this feature?"
- **Audience**: Product owners, stakeholders, QA
- **Focus**: WHAT the system must do
- **Guidelines**:
  - Each Rule represents one measurable acceptance criterion
  - Rules must be concrete, not subjective
  - Bad: "user-friendly interface"
  - Good: "creates 3 directories with correct permissions"

### Scenarios (Concrete Examples)

- **Purpose**: Describe user-facing behavior through concrete examples
- **Question**: "How does the system behave from the user's perspective?"
- **Audience**: Developers, QA, automation engineers
- **Focus**: HOW the system behaves (observable behavior)
- **Guidelines**:
  - Each Scenario represents one concrete example
  - Use Given/When/Then structure
  - Focus on observable behavior, not implementation
  - Use domain language (Ubiquitous Language), not technical jargon
  - Write from user perspective

## Structure Requirements

1. **Feature Declaration**: `Feature: <module>_<feature-name>`
   - Must follow pattern: lowercase module, underscore, lowercase feature name
   - Examples: `eac-commands_commit`, `r2r-cli_init-command`

2. **User Story**: As a.../I want.../So that... format

   ```text
   As a [role]
   I want [capability]
   So that [business value]
   ```

3. **Background** (optional): Common setup shared across scenarios

   ```text
   Background:
     Given [common precondition]
   ```

4. **Rules**: At least one Rule with acceptance criteria

   ```text
   Rule: [Measurable acceptance criterion]
   ```

5. **Scenarios**: At least one Scenario per Rule

   ```text
   Scenario: [Concrete example]
     Given [precondition]
     When [action]
     Then [outcome]
   ```

## Tagging Requirements

### Verification Tags (REQUIRED)

EVERY scenario MUST have at least one verification tag:

- `@ov` - Operational Verification (functional tests)
- `@iv` - Installation Verification (deployment validation)
- `@pv` - Performance Verification (load tests)
- `@piv` - Production Installation Verification
- `@ppv` - Production Performance Verification

### Testing Tags and Taxonomy

{{.Custom.TagsSpec}}

{{.Custom.TaxonomySpec}}

### Optional Tags

- `@L0` to `@L4` - Test environment complexity (usually inferred)
- `@deps:<system>` - External system dependencies (e.g., @deps:docker)
- `@depm:<module>` - Internal module dependencies (e.g., @depm:r2r-cli)
- `@control:<control-id>` - OSCAL control evidence link (e.g., @control:ac-2)
- `@controls:<id1>,<id2>` - Multiple controls (e.g., @controls:ac-2,au-3)
- `@skip:<reason>` - Temporarily excluded (e.g., @skip:wip)
- `@Manual` - Manual test (cannot be automated)

### OSCAL Control Tagging

**When to use control tags:**

If the feature implements security, compliance, or risk-related functionality, tag scenarios
with OSCAL control IDs to enable automated compliance evidence collection.

**How it works:**
1. Your module may have an OSCAL risk profile at `specs/.risk-controls/<module>.profile.json`
2. This profile lists relevant controls from the catalog
3. Tag scenarios that provide evidence for those controls
4. The `create risk-assess` command will automatically collect test results as evidence

**Tagging guidelines:**

- **Single control**: Use `@control:ac-2` when one scenario provides evidence for one control
- **Multiple controls**: Use `@controls:ac-2,au-3,ia-5` when one scenario covers multiple controls
- **Place after verification tags**: `@ov @control:ac-2` or `@iv @controls:ac-2,au-3`
- **Use control IDs from profile**: Check your module's risk profile for available controls
- **Match control intent**: Tag scenarios that genuinely provide evidence for the control

**Example:**

```gherkin
@ov @control:ac-2
Scenario: Create user account with approval workflow
  Given I am an administrator
  When I create a new user account
  Then an approval request must be created
  And the account must remain disabled until approved
  # @control:ac-2 = Account Management control
```

**Common control patterns:**

| Control Family | Example ID | Purpose | When to Tag |
|----------------|------------|---------|-------------|
| AC (Access Control) | ac-2, ac-3, ac-6 | Account management, access enforcement, least privilege | Login, authorization, account creation scenarios |
| AU (Audit) | au-2, au-3, au-12 | Audit events, content, generation | Logging, audit trail scenarios |
| IA (Identity/Auth) | ia-2, ia-5 | User identification, authenticator management | Authentication, password, MFA scenarios |
| SC (System/Comms) | sc-8, sc-13, sc-28 | Transmission protection, crypto, at-rest encryption | Encryption, secure communication scenarios |
| SI (System Integrity) | si-2, si-3, si-7, si-10 | Flaw remediation, malware protection, integrity, input validation | Security scanning, validation scenarios |
| CM (Config Mgmt) | cm-2, cm-3, cm-6 | Baseline config, change control, settings | Configuration, deployment scenarios |

**Available controls for this module:**

{{.Custom.AvailableControls}}

**Instructions:**
- Review the available controls above
- Tag scenarios that provide evidence for these controls
- Use `@control:<id>` for single control evidence
- Use `@controls:<id1>,<id2>` when a scenario covers multiple controls
- Omit control tags if the feature is not security/compliance-related

## Common Patterns

### Functional Test

```gherkin
@ov
Scenario: User performs successful action
  Given a precondition exists
  When the user performs an action
  Then the expected outcome occurs
```

### Error Handling

```gherkin
@ov
Scenario: User provides invalid input
  Given a precondition exists
  When the user provides invalid data
  Then an error message is displayed
  And the system state remains unchanged
```

### Deployment Validation

```gherkin
@iv
Scenario: Service deploys successfully
  Given the service package is built
  When the deployment runs
  Then the health check passes
```

### Performance Test

```gherkin
@pv
Scenario: API meets SLA under load
  Given the service is deployed
  When 1000 requests per second are sent
  Then 95th percentile response time is under 200ms
```

## Output Requirements

1. Start with Feature: declaration
2. Include complete user story
3. Add at least one Rule (acceptance criterion)
4. Add at least one Scenario per Rule
5. Tag ALL scenarios with at least one verification tag
6. Use Given/When/Then structure
7. Write in domain language - no technical jargon
8. Focus on observable behavior - not implementation
9. Keep scenarios independent and executable
10. Return ONLY valid Gherkin - no explanations, markdown fences, or commentary

Return ONLY the Gherkin content starting with "Feature:" and ending with the last scenario step.

Now generate the specification based on the user's description below:
