# Translating Example Maps to Gherkin

How to convert Example Mapping cards into Gherkin specifications.

## Card to Gherkin Mapping

| Card Color      | Gherkin Element     | Location            |
| --------------- | ------------------- | ------------------- |
| Yellow (Story)  | Feature description | `specs/`            |
| Blue (Rule)     | `Rule:` block       | `specs/`            |
| Green (Example) | `Scenario:` block   | `specs/`            |
| Red (Question)  | Issues to resolve   | `specs/` or backlog |

## Step-by-Step Translation

### Step 1: Yellow Card → Feature

**Yellow Card**:

```text
As a developer
I want to initialize a project
So that I can start coding quickly
```

**Gherkin**:

```gherkin
Feature: cli_init-project

  As a developer
  I want to initialize a project
  So that I can start coding quickly
```

### Step 2: Blue Cards → Rules

**Blue Cards**:

- "Creates project structure"
- "Configuration has defaults"
- "Prevents overwriting existing projects"

**Gherkin**:

```gherkin
Rule: Creates project directory structure

Rule: Configuration file contains default values

Rule: Existing projects are not overwritten
```

### Step 3: Green Cards → Scenarios

**Green Cards under "Creates project structure"**:

- "Empty directory → creates files"
- "With subdirectory → creates in current"

**Gherkin**:

```gherkin
Rule: Creates project directory structure

  @ov @ac1
  Scenario: Initialize in empty directory
    Given I am in an empty folder
    When I run "r2r init"
    Then a file named "r2r.yaml" should be created
    And a directory named "specs/" should be created

  @ov @ac1
  Scenario: Initialize with existing subdirectories
    Given I am in a directory with "src/"
    When I run "r2r init"
    Then "r2r.yaml" should be created in current directory
    And "src/" should remain unchanged
```

### Step 4: Red Cards → Actions

**Red Cards**:

- "What if no write permission?"
- "What about .gitignore?"

**Actions**:

1. Resolve before implementation (ask product owner)
2. Add as new scenarios once resolved
3. Or document in issues.md if deferred

## Complete Example

### Cards

```text
YELLOW: User registration

BLUE 1: Email must be validated
  GREEN: Valid email succeeds
  GREEN: Invalid format rejected

BLUE 2: Duplicate emails prevented
  GREEN: New email → account created
  GREEN: Existing email → error shown

RED: What about disposable emails?
```

### Gherkin Output

```gherkin
Feature: user_registration

  As a visitor
  I want to register an account
  So that I can access the application

  Rule: Email address must be validated

    @ov @ac1
    Scenario: Valid email succeeds
      Given I am on the registration page
      When I enter email "user@example.com"
      And I submit the form
      Then I should see "Check your email"

    @ov @ac1
    Scenario: Invalid email format rejected
      Given I am on the registration page
      When I enter email "not-an-email"
      And I submit the form
      Then I should see "Invalid email format"

  Rule: Duplicate email addresses are prevented

    @ov @ac2
    Scenario: New email creates account
      Given no account exists with "new@example.com"
      When I register with "new@example.com"
      Then my account should be created

    @ov @ac2
    Scenario: Existing email shows error
      Given an account exists with "existing@example.com"
      When I register with "existing@example.com"
      Then I should see "Email already registered"
```

## Checklist

- [ ] One Feature per file
- [ ] Feature name follows `module_feature-name` format
- [ ] User story in Feature description
- [ ] One Rule per Blue card
- [ ] Scenarios under their Rule
- [ ] All scenarios have `@ov` (or other verification tag)
- [ ] All scenarios have `@ac<n>` linking to Rule
- [ ] Red card questions resolved or documented

## Related

- [Running Example Mapping](running-example-mapping.md)
- [Creating Feature Files](creating-feature-files.md)
- [Feature Naming](../../reference/specifications/feature-naming.md)
