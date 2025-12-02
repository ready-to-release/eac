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
- `@risk-control:<name>-<id>` - Compliance/security requirements
- `@skip:<reason>` - Temporarily excluded (e.g., @skip:wip)
- `@Manual` - Manual test (cannot be automated)

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
