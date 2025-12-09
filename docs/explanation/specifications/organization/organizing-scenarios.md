# Organizing Scenarios

{{ page_breadcrumb() }}

> **Scenario coverage strategy and independence**

Learn how to write effective scenarios that thoroughly test each Rule.

---

## Scenario Coverage Strategy

Each Rule needs to be clearly defined by:

1. **Happy path** - Primary use case
2. **Error cases** - How system handles failures
3. **Edge cases** - Boundary conditions, special states

---

## Coverage Example

```gherkin
Rule: Configuration file must contain valid module definitions

  @ov
  Scenario: Valid configuration with all required fields
    Given I have a configuration with module name and path
    When I validate the configuration
    Then validation should succeed

  @ov
  Scenario: Configuration missing required module name
    Given I have a configuration without a module name
    When I validate the configuration
    Then I should see error "module.name is required"

  @ov
  Scenario: Configuration with invalid module path
    Given I have a configuration with non-existent path
    When I validate the configuration
    Then I should see error "module.path does not exist: src/invalid"
```

**Coverage breakdown**:
- **Scenario 1**: Happy path - valid configuration
- **Scenario 2**: Error case - missing required field
- **Scenario 3**: Error case - invalid field value

---

## Scenario Independence

**Critical principle**: Each scenario must be **completely independent**.

### Why Independence Matters

**Parallel Execution**:
- Scenarios may run in **any order**
- Test runners execute scenarios concurrently for speed
- No guaranteed execution order

**Filtered Runs**:
- You may run subset of scenarios (by tag, by feature, etc.)
- Scenarios cannot rely on others having run first

**Failure Isolation**:
- Scenario failure doesn't cascade to others
- Each scenario can be debugged independently

**Self-Documentation**:
- Each scenario is self-documenting
- Complete context visible within the scenario
- No need to read other scenarios to understand this one

---

## Achieving Independence

### Each Scenario Has Complete Setup

**Bad** (depends on previous scenario):

```gherkin
Scenario: Create user account
  Given I am on the registration page
  When I submit valid registration
  Then account should be created

Scenario: Login with new account
  # ❌ Assumes previous scenario ran and created account
  When I login with the credentials
  Then I should be logged in
```

**Good** (independent scenarios):

```gherkin
Scenario: Create user account
  Given I am on the registration page
  When I submit valid registration
  Then account should be created

Scenario: Login with existing account
  Given a user account exists with email "test@example.com"
  And I am on the login page
  When I login with "test@example.com" and "password"
  Then I should be logged in
```

### Use Background for Common Setup

If **all scenarios** in a Rule need the same setup, use Background:

```gherkin
Rule: User authentication

  Background:
    Given a user account exists with email "test@example.com"
    And I am on the login page

  @ov
  Scenario: Login with valid credentials
    When I login with "test@example.com" and "password"
    Then I should be logged in

  @ov
  Scenario: Login with invalid password
    When I login with "test@example.com" and "wrongpassword"
    Then I should see error "Invalid credentials"
```

**Important**: Only use Background if **all scenarios** need it. If only some scenarios need setup, include it in those specific scenarios.

---

## Reusability in Gherkin

Gherkin is a formal language designed to enable reusability.

### Step Reusability

Design each step so it can be reused in many places:

**Good** (reusable steps):

```gherkin
Given a user account exists with email "test@example.com"
When I login with "test@example.com" and "password"
Then I should see message "Welcome back"
```

These steps can be used across many scenarios and features.

**Bad** (too specific):

```gherkin
Given the test user account is created
When I perform the login action
Then the welcome screen appears
```

These steps are vague and tied to specific context.

### Benefits of Reusability

**Consistent Language**:
- Team discovers similarities in the domain
- Extracts and makes them explicit
- Deepens the ubiquitous language

**Cleaner Implementation**:
- Each step gets a clean, reusable implementation
- Less code duplication in step definitions
- Easier to maintain

**Domain Understanding**:
- Reusable steps reveal domain patterns
- Explicit context in specifications
- Better communication between stakeholders and developers

---

## Learn More About Scenario Independence

To learn more about this critical principle, read [Effective Behaviour Driven Development](../references.md#effective-behavior-driven-development) by Gáspár Nagy and Seb Rose.

---

## Best Practices

### DO

- ✅ Make each scenario completely independent
- ✅ Include all necessary setup in the scenario
- ✅ Use Background only for setup needed by ALL scenarios
- ✅ Write reusable steps
- ✅ Test scenarios in random order to verify independence
- ✅ Cover happy path, error cases, and edge cases

### DON'T

- ❌ Assume scenarios run in specific order
- ❌ Make scenarios depend on each other
- ❌ Use Background for setup needed by only some scenarios
- ❌ Write overly specific steps that can't be reused
- ❌ Skip error cases and edge cases
- ❌ Create scenarios with implicit dependencies

---

## Scenario Independence Checklist

Use this checklist for each scenario:

- [ ] Can this scenario run first in the test suite?
- [ ] Can this scenario run last in the test suite?
- [ ] Can this scenario run alone (all other scenarios skipped)?
- [ ] Does this scenario have all necessary setup?
- [ ] Will this scenario pass regardless of what other scenarios do?
- [ ] Is the scenario outcome deterministic?

If you answer "no" to any question, the scenario is not independent.

---

## Example: Complete Rule with Independent Scenarios

```gherkin
Rule: Creates project directory structure

  @ov
  Scenario: Initialize in empty directory
    Given I am in an empty folder
    When I run "r2r init"
    Then a file named "r2r.yaml" should be created
    And a directory named "src" should be created

  @ov
  Scenario: Initialize in existing project
    Given I am in a directory with "r2r.yaml"
    When I run "r2r init"
    Then the command should fail
    And I should see error "Project already initialized"

  @ov
  Scenario: Initialize with custom config
    Given I am in an empty folder
    When I run "r2r init --config custom.yaml"
    Then a file named "custom.yaml" should be created
```

**Why these scenarios are independent**:
- Each scenario sets up its own context (empty folder, existing project, etc.)
- No scenario depends on another having run
- Each can run in any order
- Each has complete setup and verification

---

## Related Documentation

- [Organizing Rules](./organizing-rules.md) - Creating acceptance criteria
- [File Structure](./file-structure.md) - Where scenarios live
- [Size Guidelines](./size-guidelines.md) - How many scenarios per Rule

{{ diataxis_footer() }}
