# Specifications (BDD)

{{ page_breadcrumb() }}

Specifications in EAC use Gherkin syntax to define behavior-driven development (BDD) scenarios that serve as both documentation and executable tests.

## What are Specifications?

EAC's specification system enables you to:

- **Generate Gherkin specs** from natural language using AI
- **Validate specification structure** against quality contracts
- **Detect unused step definitions** in test code
- **Maintain living documentation** that executes as tests

The system uses AI to transform requirements into properly structured Gherkin feature files following project conventions.

## When to Use Specifications

Use specification commands when you need:

| Scenario                          | Commands             |
| --------------------------------- | -------------------- |
| Creating new feature specs        | `create spec`        |
| Validating spec quality           | `validate specs`     |
| Finding orphaned step definitions | `get specs unused-steps` |

### Common Use Cases

- **TDD workflow** - Write specs before implementation
- **Acceptance criteria** - Define done for features
- **Documentation** - Living docs that prove behavior
- **Stakeholder communication** - Readable by non-developers
- **Regression testing** - Automated behavior verification

## Key Concepts

### Gherkin Syntax

Gherkin uses a structured format:

```gherkin
Feature: User Authentication
  Users can securely log into the system.

  Rule: Valid credentials grant access

    Scenario: Successful login with valid credentials
      Given a registered user with email "user@example.com"
      And the user has password "SecurePass123"
      When the user submits login credentials
      Then the user should be authenticated
      And a session token should be issued

    Scenario: Failed login with invalid password
      Given a registered user with email "user@example.com"
      When the user submits wrong password "WrongPass"
      Then authentication should fail
      And an error message should be displayed
```

### Feature Structure

| Element      | Purpose                | Example                                 |
| ------------ | ---------------------- | --------------------------------------- |
| **Feature**  | High-level capability  | `Feature: User Authentication`          |
| **Rule**     | Business rule grouping | `Rule: Valid credentials grant access`  |
| **Scenario** | Specific test case     | `Scenario: Successful login`            |
| **Given**    | Preconditions          | `Given a registered user`               |
| **When**     | Action under test      | `When the user submits credentials`     |
| **Then**     | Expected outcome       | `Then the user should be authenticated` |

### Tags

Tags control test execution and categorization:

```gherkin
@smoke @auth
Feature: User Authentication

  @integration
  Scenario: Login flow
    ...

  @skip:not-implemented
  Scenario: MFA login
    ...
```

Common tags:

- `@smoke` - Quick sanity tests
- `@integration` - Integration tests
- `@e2e` - End-to-end tests
- `@skip:<reason>` - Skip with reason
- `@env:<environment>` - Environment-specific

### Step Definitions

Gherkin steps map to code:

```go
// steps/auth_steps.go
func (s *AuthSteps) aRegisteredUserWithEmail(email string) error {
    s.user = createUser(email)
    return nil
}

func init() {
    godog.Given(`^a registered user with email "([^"]*)"$`, s.aRegisteredUserWithEmail)
}
```

## Workflow Overview

### TDD Workflow (Spec First)

```bash
# 1. Generate spec from requirements
r2r eac create spec "User can reset password via email"

# 2. Review generated spec
cat specs/src-auth/password-reset.feature

# 3. Run tests (should fail - no implementation)
r2r eac test src-auth

# 4. Implement step definitions
# ... write code ...

# 5. Run tests (should pass)
r2r eac test src-auth

# 6. Commit
r2r eac work commit --all
```

### Validation Workflow

```bash
# 1. Validate all specs meet quality standards
r2r eac validate specs

# 2. Check for undefined tags
r2r eac validate test-tags

# 3. Find unused step definitions
r2r eac get specs unused-steps
```

### AI Generation

When generating specs, provide clear requirements:

```bash
# Simple feature
r2r eac create spec "Users can update their profile picture"

# Detailed requirements
r2r eac create spec "Users can update their profile picture. \
  They should be able to upload JPG or PNG files up to 5MB. \
  The old picture should be deleted. \
  A thumbnail should be generated automatically."
```

## Spec Organization

### Directory Structure

```text
specs/
├── src-auth/
│   ├── login.feature
│   ├── logout.feature
│   └── password-reset.feature
├── src-api/
│   ├── users.feature
│   └── products.feature
└── src-core/
    └── validation.feature
```

### Naming Conventions

- Feature files: `kebab-case.feature`
- One feature per file
- Group by module/domain
- Match module structure

## Integration Points

### With Testing

Specs execute via godog (Go BDD framework):

```bash
# Run all specs
r2r eac test

# Run specific module specs
r2r eac test src-auth

# Run by tag
r2r eac test --suite smoke
```

### With Validation

Tag validation ensures consistency:

```bash
# Validate tags are defined
r2r eac validate test-tags

# Output:
# ✅ All test tags are defined in the tag contract
#    Validated 47 feature files
#    Contract defines 15 valid tags
```

### With Risk Management

Specs provide evidence for compliance:

```bash
# Link spec to control
@control:AC-2
Scenario: User provisioning
  ...

# Assessment includes spec results
r2r eac create risk-assess
```

### With CI/CD

```yaml
- name: Validate specifications
  run: r2r eac validate specs

- name: Run BDD tests
  run: r2r eac test --as-junit

- name: Check for unused steps
  run: r2r eac get specs unused-steps
```

## Best Practices

### Writing Good Specs

- **Business language** - Avoid technical jargon in Given/When/Then
- **Single behavior** - One scenario tests one thing
- **Independent scenarios** - No dependencies between scenarios
- **Declarative style** - Describe what, not how

### Do's

```gherkin
# Good: Declarative, business-focused
Scenario: Customer receives order confirmation
  Given a customer has items in their cart
  When the customer completes checkout
  Then an order confirmation email should be sent
```

### Don'ts

```gherkin
# Bad: Imperative, technical
Scenario: Send email
  Given I click the cart button
  And I click the checkout button
  And I enter my credit card "4111111111111111"
  And I click submit
  Then the sendEmail function should be called
```

## Next Steps

- [Specifications Configuration](specifications-configuration.md) - Configure AI prompts and tag contracts
- [Specifications Commands](specifications-commands.md) - Full command reference

## Related Areas

- [Build Commands](build-overview.md) - Building modules
- [Test Commands](test-overview.md) - Running specification tests
- [Validate Commands](validate-overview.md) - Tag and spec validation
- [Risk Management](risks-overview.md) - Specs as compliance evidence

{{ diataxis_footer() }}
