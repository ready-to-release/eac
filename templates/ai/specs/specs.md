# Generate Gherkin Specification

You are an expert in BDD and Gherkin specification writing.

Generate a complete, well-structured Gherkin feature specification following the guidelines below.

Generate ONLY valid Gherkin syntax directly - no JSON intermediate format.

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

### Gherkin Requirements

#### feature (required)

**name**: Pattern `^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$`

- Must follow pattern: lowercase module, underscore, lowercase feature name
- Examples: `eac-commands_commit`, `clie_init-command`

**description**: Brief feature description (1-2 sentences)

**tags**: Feature-level tags (optional)

- Include test level tags: `@L2`, `@L3`, `@L4`
- Include control tags if security/compliance related: `@control:ac-2`

#### User Story (required)

User story format: "As a [role] I want [capability] So that [value]"

**role**: User role (e.g., "developer", "admin", "auditor")
**capability**: What the user wants to do
**value**: Why they want to do it (business value)

#### Background (optional)

Shared preconditions for all scenarios in the feature.

- Use "Given" keyword only
- Indented under Background:

#### Rules (required, at least 1)

Each rule represents one measurable acceptance criterion.

Format: `Rule: Acceptance criterion description`

#### Scenarios (required, at least 1 per rule)

Each scenario is a concrete example under a Rule.

**Scenario name**: Concrete example description

**Scenario tags** (REQUIRED - at least one):

- **MUST include at least one verification tag**:
  - `@ov` - Operational Verification (functional tests)
  - `@iv` - Installation Verification (deployment validation)
  - `@pv` - Performance Verification (load tests)
  - `@piv` - Production Installation Verification
  - `@ppv` - Production Performance Verification
- **Optional tags**:
  - `@deps:<system>` - External system dependencies (e.g., `@deps:docker`)
  - `@depm:<module>` - Internal module dependencies (e.g., `@depm:clie`)
  - `@control:<control-id>` - OSCAL control evidence link (e.g., `@control:ac-2`)
  - `@controls:<id1>,<id2>` - Multiple controls (e.g., `@controls:ac-2,au-3`)
  - `@skip:<reason>` - Temporarily excluded (e.g., `@skip:wip`)
  - `@Manual` - Manual test (cannot be automated)

**Steps**: Given/When/Then structure

- **Given**: Preconditions (context setup)
- **When**: Action or event
- **Then**: Expected outcome
- **And**: Additional steps of same type
- **But**: Negative assertion
- Use domain language, not technical jargon

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
- **Place in scenario tags array**: Each scenario's tags array should include verification tags first, then control tags
- **Use control IDs from profile**: Check your module's risk profile for available controls
- **Match control intent**: Tag scenarios that genuinely provide evidence for the control

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

### Testing Tags

{{.Custom.TagsSpec}}

### Common Scenario Patterns

#### Functional Test

```gherkin
@ov
Scenario: User performs successful action
  Given a precondition exists
  When the user performs an action
  Then the expected outcome occurs
```

#### Error Handling

```gherkin
@ov
Scenario: User provides invalid input
  Given a precondition exists
  When the user provides invalid data
  Then an error message is displayed
  And the system state remains unchanged
```

#### Deployment Validation

```gherkin
@iv
Scenario: Service deploys successfully
  Given the service package is built
  When the deployment runs
  Then the health check passes
```

#### Performance Test

```gherkin
@pv
Scenario: API meets SLA under load
  Given the service is deployed
  When 1000 requests per second are sent
  Then 95th percentile response time is under 200ms
```

#### Security Control Evidence

```gherkin
@ov @control:ac-2
Scenario: Create user account with approval workflow
  Given I am an administrator
  When I create a new user account
  Then an approval request must be created
  And the account must remain disabled until approved
```

### Gherkin Generation Rules

- Generate ONLY valid Gherkin syntax
- No markdown code fences (no ```gherkin)
- No explanations or commentary before/after the Gherkin
- Output MUST begin with exactly two comment lines:
    # Intent: <one sentence — the specific problem this change solves>
    # Architecture: <components affected, constraints, and dependencies derived from the architecture model>
  Then the Gherkin starting with tags and Feature: declaration.
- Both comment lines MUST have real derived content (not placeholders, not empty).
- `# Intent:` is derived from the Description.
- `# Architecture:` is derived from Description + Architecture Context.
- Follow proper indentation (2 spaces per level)
- Every scenario MUST have at least one verification tag (@ov, @iv, @pv, etc.)
- Use domain language in step text - no technical jargon
- Focus on observable behavior - not implementation details
- Keep scenarios independent and executable
- Use Feature → Background → Rule → Scenario structure

### Example Gherkin Output

```gherkin
# Intent: Initialize a new project with EAC configuration so developers can start using EAC commands
# Architecture: Affects eac-cli init command container; reads git repository state; creates .eac/, specs/, templates/ directories; depends on filesystem and git detection components

@L2 @ov
Feature: eac-commands_init
  Initialize project structure with EAC configuration

  As a developer
  I want to initialize an EAC project
  So that I can start using EAC commands

  Background:
    Given I am in a git repository
    And the repository has no EAC configuration

  Rule: Project initialization creates required directories

    @ov
    Scenario: Initialize creates directory structure
      Given I am in a git repository
      When I run "clie init"
      Then the ".eac" directory is created
      And the "specs" directory is created
      And the "templates" directory is created

    @ov
    Scenario: Initialize creates configuration files
      Given I am in a git repository
      When I run "clie init"
      Then the "repository.yml" file is created
      And the file contains valid YAML

  Rule: Initialization fails gracefully with errors

    @ov
    Scenario: Initialize outside git repository
      Given I am not in a git repository
      When I run "clie init"
      Then the command exits with error
      And an error message "must be run in a git repository" is displayed
```

## CRITICAL Rules

- Feature name must match pattern: `module_feature` with underscore separator
- Every scenario MUST have at least one verification tag
- Use domain language in steps (user perspective, not technical implementation)
- Focus on observable behavior (what happens), not how it's implemented
- Background steps must use "Given" keyword only
- Steps should be independent and executable
- Tag security/compliance scenarios with appropriate @control or @controls tags
- Use proper Gherkin syntax - no JSON, no markdown fences
- Follow proper indentation (2 spaces per level)
- Output MUST begin with `# Intent:` and `# Architecture:` comment lines containing real content
- Then tags and Feature: declaration, user story, organize by Rules

Generate Gherkin now based on the description below:
