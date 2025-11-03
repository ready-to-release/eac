# Requirements Templates

This folder contains templates for creating feature specifications using a **three-layer testing approach**: ATDD (Acceptance Test-Driven Development), BDD (Behavior-Driven Development), and TDD (Test-Driven Development).

## What's Included

| Template | Purpose | Tool | Format |
|----------|---------|------|--------|
| [acceptance.spec](./acceptance.spec) | Business requirements and acceptance criteria | Gauge | Markdown |
| [behavior.feature](./behavior.feature) | Executable scenarios in Gherkin format | Godog/Cucumber | Gherkin |

These templates help you create comprehensive, testable specifications that bridge the gap between business requirements and technical implementation.

---

## The Three-Layer Testing Approach

### Layer 1: ATDD (Acceptance Test-Driven Development)
**Tool**: Gauge
**File**: `acceptance.spec`
**Purpose**: Define business value and acceptance criteria

- Written in natural language (markdown)
- Documents user stories and acceptance criteria
- Captures Example Mapping workshop results
- Defines what "done" means for the feature
- Readable by all stakeholders

### Layer 2: BDD (Behavior-Driven Development)
**Tool**: Godog (Go) or Cucumber (Ruby/Java/etc)
**File**: `behavior.feature`
**Purpose**: Specify executable scenarios in Given/When/Then format

- Written in Gherkin syntax
- Converts acceptance criteria into testable scenarios
- Bridges business language and technical implementation
- Automated through step definitions
- Tagged by verification type (IV/OV/PV)

### Layer 3: TDD (Test-Driven Development)
**Tool**: Language-specific test frameworks (Go test, pytest, Jest, etc.)
**Purpose**: Drive implementation through unit tests

- Written in your programming language
- Tests individual components and functions
- Fast, isolated, repeatable
- Written before implementation code

---

## Quick Start

### 1. Set Up Your Project Structure

Create a `requirements/` folder in your project root:

```bash
mkdir -p requirements/<component>/<feature_name>
```

**Examples**:
- `requirements/api/user_authentication/`
- `requirements/ui/dashboard_widgets/`
- `requirements/service/notification_delivery/`

### 2. Copy Templates

Copy the templates to your feature folder:

```bash
cp /path/to/templates/acceptance.spec requirements/<component>/<feature_name>/
cp /path/to/templates/behavior.feature requirements/<component>/<feature_name>/
```

### 3. Fill In the Templates

Replace all placeholders:
- `<component>`: Your component/module name
- `<feature_name>`: Your feature identifier
- `<Component Name>`: Full component display name
- `[Feature Name]`: Human-readable feature title
- `[descriptive text]`: Your specific content

### 4. Install Required Tools

**For ATDD (Gauge)**:
```bash
# Install Gauge
curl -SsL https://downloads.gauge.org/stable | sh

# Or using package managers
brew install gauge          # macOS
choco install gauge         # Windows
apt-get install gauge       # Linux
```

**For BDD (Godog for Go projects)**:
```bash
go install github.com/cucumber/godog/cmd/godog@latest
```

**For BDD (Cucumber for other languages)**:
```bash
# Ruby
gem install cucumber

# JavaScript
npm install --save-dev @cucumber/cucumber

# Java
# Add to pom.xml or build.gradle
```

---

## Verification Types Explained

The templates use three **verification types** to categorize tests for audit, regulatory, and implementation reporting:

### Installation Verification (IV)

**Tag**: `@IV` (in behavior.feature)
**Purpose**: Verify system installation, setup, and configuration

**Use for**:
- Installation and deployment procedures
- System configuration and setup
- Version verification
- Environment baseline checks
- Initial system readiness

**Example scenarios**:
```gherkin
@success @ac1 @IV
Scenario: System installs successfully on clean environment
  Given a clean system environment
  When installation is performed
  Then the system should be installed correctly
  And configuration files should exist
```

---

### Operational Verification (OV)

**Tag**: (none - this is the default)
**Purpose**: Verify functional behavior and business logic

**Use for**:
- Functional requirements
- Business logic
- Error handling
- Data processing
- User workflows
- Integration with external systems

**Example scenarios**:
```gherkin
@success @ac2
Scenario: User can perform primary operation successfully
  Given the system is ready
  When user performs the operation
  Then the operation should succeed
  And expected results should be produced
```

---

### Performance Verification (PV)

**Tag**: `@PV` (in behavior.feature)
**Purpose**: Verify performance requirements and resource usage

**Use for**:
- Response time requirements
- Resource usage limits (memory, CPU, disk)
- Throughput and scalability
- Performance under load
- Concurrent operation handling

**Example scenarios**:
```gherkin
@success @ac3 @PV
Scenario: Operation completes within performance threshold
  Given typical operating conditions
  When operation is performed
  Then it should complete within 2 seconds
  And resource usage should remain acceptable
```

---

## Feature ID Convention

Every feature has a unique Feature ID that appears in all related files:

**Pattern**: `<component>_<feature_name>`

**Examples**:
- `api_user_authentication`
- `ui_dashboard_widgets`
- `service_notification_delivery`
- `backend_data_export`
- `cli_project_initialization`

**Where it appears**:
- `acceptance.spec`: `> **Feature ID**: api_user_authentication`
- `behavior.feature`: `# Feature ID: api_user_authentication`
- Test files: `// Feature: api_user_authentication` or `# Feature: api_user_authentication`

This enables **traceability** from requirements through implementation to test results.

---

## Acceptance Criteria Linking

Link scenarios to acceptance criteria using `@ac` tags:

**In acceptance.spec**:
```markdown
### AC1: System creates user account
**Validated by**: behavior.feature -> @ac1 scenarios

### AC2: System validates user input
**Validated by**: behavior.feature -> @ac2 scenarios
```

**In behavior.feature**:
```gherkin
@success @ac1
Scenario: Create user account with valid data
  # Tests AC1

@error @ac2
Scenario: Reject invalid email format
  # Tests AC2
```

Each acceptance criterion typically has 2-4 scenarios testing different aspects.

---

## Example Mapping Workshop

Before writing specifications, conduct an **Example Mapping** workshop:

### The Card System

**Yellow Card** (1 per feature):
- The user story
- "As a... I want... So that..."

**Blue Cards** (2-5 per feature):
- Rules / Acceptance Criteria
- What must be true for the feature to work

**Green Cards** (2-4 per Blue Card):
- Concrete examples
- Specific test cases
- Format: `[Context] → [Action] → [Result]`

**Red Cards** (as needed):
- Questions and unknowns
- Track these in `issues.md`

### Example

**Yellow Card**:
> As a user, I want to export data to CSV, so that I can analyze it in Excel

**Blue Card 1**: Generates valid CSV format
- **Green Card 1a**: Empty data → export → valid CSV with headers only
- **Green Card 1b**: 100 rows → export → CSV with all rows
- **Green Card 1c**: Special chars → export → properly escaped CSV

**Blue Card 2**: Handles large datasets efficiently
- **Green Card 2a**: 10K rows → export → completes within 5 seconds
- **Green Card 2b**: 100K rows → export → completes within 30 seconds

**Red Card**: What's the maximum row limit?

---

## Directory Structure

Recommended structure for features:

```
requirements/
├── <component_1>/
│   ├── <feature_1>/
│   │   ├── acceptance.spec          # ATDD - business requirements
│   │   ├── behavior.feature         # BDD - executable scenarios
│   │   ├── acceptance_test.go       # Gauge step implementations
│   │   ├── step_definitions_test.go # Godog step definitions
│   │   └── issues.md               # Optional: open questions
│   └── <feature_2>/
│       ├── acceptance.spec
│       └── behavior.feature
└── <component_2>/
    └── <feature_3>/
        ├── acceptance.spec
        └── behavior.feature
```

---

## Writing Effective Scenarios

### Good Scenario (Clear and Testable)

```gherkin
@success @ac1
Scenario: User successfully logs in with valid credentials
  Given a registered user account exists with email "user@example.com"
  And the password is "ValidPass123!"
  When the user submits login credentials
  Then the user should be authenticated
  And a session token should be created
  And the user should be redirected to the dashboard
```

**Why it's good**:
- Specific, concrete example
- Clear preconditions
- Observable outcomes
- Can be automated

### Poor Scenario (Vague and Untestable)

```gherkin
Scenario: Login works
  When user logs in
  Then everything is fine
```

**Why it's poor**:
- No specific details
- No preconditions
- Vague outcomes
- Can't be automated

---

## Scenario Count Guidelines

Keep feature files manageable:

| Scenario Count | Status | Action |
|----------------|--------|--------|
| 10-15 | ✅ Ideal | Optimal size |
| 15-20 | ✅ Acceptable | Still manageable |
| 20-30 | ⚠️ Large | Consider splitting |
| 30+ | ❌ Too Large | Must split |

**When to split**: Create multiple `.feature` files for sub-features:
- `user_authentication_login.feature`
- `user_authentication_registration.feature`
- `user_authentication_password_reset.feature`

---

## Running Tests

### Run ATDD Tests (Gauge)

```bash
# Run all acceptance tests
gauge run requirements/

# Run specific component
gauge run requirements/api/

# Run specific feature
gauge run requirements/api/user_authentication/
```

### Run BDD Tests (Godog)

```bash
# Run all behavior tests
godog requirements/**/behavior.feature

# Run by verification type
godog --tags="@IV" requirements/**/behavior.feature    # Installation only
godog --tags="@PV" requirements/**/behavior.feature    # Performance only
godog --tags="~@IV && ~@PV" requirements/**/behavior.feature  # Operational only

# Run specific feature
godog requirements/api/user_authentication/behavior.feature

# Run with tags
godog --tags="@success" requirements/**/behavior.feature
godog --tags="@ac1" requirements/**/behavior.feature
```

### Run BDD Tests (Cucumber - Other Languages)

```bash
# Ruby
cucumber features/

# JavaScript
npx cucumber-js

# Java
mvn test
```

### Generate Reports for Audit

```bash
# Generate separate reports by verification type
godog --tags="@IV" \
  --format=junit:test-results/iv-tests.xml \
  requirements/**/behavior.feature

godog --tags="@PV" \
  --format=junit:test-results/pv-tests.xml \
  requirements/**/behavior.feature

godog --tags="~@IV && ~@PV" \
  --format=junit:test-results/ov-tests.xml \
  requirements/**/behavior.feature
```

These reports can be used in implementation reports for regulatory/audit documentation.

---

## Common Tags

### Feature-Level Tags
Apply to entire feature (all scenarios):

| Tag | Purpose | Example Use Case |
|-----|---------|------------------|
| `@<component>` | Component identifier | `@api`, `@ui`, `@service` |
| `@critical` | Business-critical | Must pass for release |
| `@integration` | External systems | Database, API, third-party |

### Scenario-Level Tags
Apply to individual scenarios:

| Tag | Purpose | Example Use Case |
|-----|---------|------------------|
| `@success` | Happy path | Normal operation succeeds |
| `@error` | Error case | Invalid input, failures |
| `@IV` | Installation Verification | Setup, config, deployment |
| `@PV` | Performance Verification | Response time, resources |
| `@ac1`, `@ac2`, ... | Acceptance criteria link | Maps to acceptance.spec |
| `@wip` | Work in progress | Exclude from CI runs |
| `@security` | Security scenarios | Auth, injection, access control |

---

## Workflow

### 1. Planning Phase
- Conduct Example Mapping workshop
- Create Yellow Card (user story)
- Define Blue Cards (acceptance criteria)
- Generate Green Cards (examples)
- Note Red Cards (questions)

### 2. Specification Phase
- Copy templates to feature folder
- Fill in `acceptance.spec` with workshop results
- Convert Green Cards to Gherkin scenarios in `behavior.feature`
- Tag scenarios appropriately (@IV, @PV, @ac tags)
- Ensure Feature ID is consistent

### 3. Implementation Phase
- Write step definitions for scenarios
- Write failing unit tests (TDD)
- Implement feature code
- Make tests pass
- Refactor

### 4. Verification Phase
- Run all three test layers (ATDD, BDD, TDD)
- Generate reports
- Verify all acceptance criteria are met
- Update documentation as needed

---

## Best Practices

### 1. Start with the User Story
Always begin with "As a... I want... So that..." to clarify the business value.

### 2. Keep Scenarios Focused
One scenario = One behavior. Don't try to test everything in one scenario.

### 3. Use Background for Common Setup
Put shared preconditions in the `Background` section.

### 4. Make Scenarios Readable
Write scenarios that non-technical stakeholders can understand.

### 5. Use Scenario Outlines for Repetition
Don't copy-paste similar scenarios. Use `Scenario Outline` with `Examples` tables.

### 6. Tag Appropriately
Use verification tags (@IV, @PV) and acceptance criteria tags (@ac1, @ac2).

### 7. Keep Feature Files Manageable
Split large feature files (30+ scenarios) into focused sub-features.

### 8. Link Everything
Maintain traceability: User Story → Acceptance Criteria → Scenarios → Implementation → Tests.

---

## Language-Specific Adaptations

### Go Projects
- BDD: Use Godog
- TDD: Use `go test`
- Example: `requirements/<component>/<feature>/step_definitions_test.go`

### Python Projects
- BDD: Use behave or pytest-bdd
- TDD: Use pytest
- Example: `requirements/<component>/<feature>/steps.py`

### JavaScript/TypeScript Projects
- BDD: Use @cucumber/cucumber
- TDD: Use Jest or Mocha
- Example: `requirements/<component>/<feature>/steps.ts`

### Java Projects
- BDD: Use Cucumber-JVM
- TDD: Use JUnit or TestNG
- Example: `requirements/<component>/<feature>/StepDefinitions.java`

---

## Customization

### Adapting Templates for Your Project

1. **Modify placeholder names**: Change `<component>` to match your architecture
2. **Adjust verification types**: Add custom tags if needed (e.g., `@smoke`, `@regression`)
3. **Update tool references**: If using different BDD/ATDD tools
4. **Add project-specific sections**: Security requirements, compliance, etc.
5. **Create project conventions**: Document in your project's README

### Adding Custom Verification Types

If IV/OV/PV don't fit your needs, you can define custom types:

```markdown
### Smoke Verification (SV)
**Tag**: @SV
**Purpose**: Quick sanity checks after deployment
```

Document your custom types clearly so the team uses them consistently.

---

## Benefits of This Approach

### For Business Stakeholders
- ✅ Clear visibility into what's being built
- ✅ Specifications in natural language
- ✅ Easy to review and approve requirements
- ✅ Living documentation that stays up-to-date

### For Developers
- ✅ Clear acceptance criteria before coding
- ✅ Automated tests from requirements
- ✅ Confidence in implementation correctness
- ✅ Regression protection

### For QA/Testers
- ✅ Test cases derived from requirements
- ✅ Automated test execution
- ✅ Clear verification types for testing strategy
- ✅ Traceability from requirements to tests

### For Compliance/Audit
- ✅ Documented requirements with verification
- ✅ Test results grouped by verification type
- ✅ Traceability from requirements through implementation
- ✅ Audit-ready reports (IV/OV/PV)

---

## Troubleshooting

### "Scenarios are too technical"
→ Focus on **what** the system does, not **how** it does it. Avoid implementation details.

### "Too many scenarios in one file"
→ Split into multiple `.feature` files by sub-feature or verification type.

### "Scenarios are redundant with unit tests"
→ BDD scenarios test **behavior from user perspective**. Unit tests test **implementation details**.

### "Stakeholders won't read Gherkin"
→ Start with `acceptance.spec` (natural language). Gherkin is for developers/QA.

### "Example Mapping takes too long"
→ Timebox to 25 minutes. If you can't finish, the feature is too large - split it.

---

## Additional Resources

### Learn More About the Methodology

- **Example Mapping**: https://cucumber.io/blog/bdd/example-mapping-introduction/
- **BDD**: https://cucumber.io/docs/bdd/
- **ATDD**: https://gauge.org/
- **Gherkin Syntax**: https://cucumber.io/docs/gherkin/reference/

### Tools

- **Gauge** (ATDD): https://gauge.org/
- **Godog** (BDD for Go): https://github.com/cucumber/godog
- **Cucumber** (BDD multi-language): https://cucumber.io/
- **behave** (BDD for Python): https://behave.readthedocs.io/

---

## Questions or Issues?

If you have questions about using these templates:

1. Review the template instructions at the bottom of each file
2. Check existing feature specifications in your project for examples
3. Consult your team's testing documentation
4. Refer to the official tool documentation (Gauge, Godog, Cucumber)

---

**Version**: 1.0
**Last Updated**: 2025-11-03
**License**: Use freely in your projects

These templates are designed to be project-agnostic and adaptable to any software project adopting a specification-driven testing approach.
