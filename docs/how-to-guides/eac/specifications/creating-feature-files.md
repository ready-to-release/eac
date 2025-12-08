<!-- EDITOR
# Editor: how-to-guides/specifications/creating-feature-files.md

## Soul

Step-by-step guide for creating Gherkin feature files from templates with placeholder replacement, rule addition, and validation.

## Sections

1. Step 1: Copy Template
2. Step 2: Replace Placeholders
3. Before
4. After
5. Step 3: Add Rules from Example Mapping
6. Step 4: Add Scenarios for Each Rule
7. Step 5: Add Required Tags
8. File Structure
9. Validation
10. Related
-->

# Creating Feature Files

How to create new Gherkin specification files from templates.

## Step 1: Copy Template

```bash
mkdir -p specs/<module>/<feature>
cp templates/specs/specification.feature specs/<module>/<feature>/specification.feature
```

Example:

```bash
mkdir -p specs/cli/init-project
cp templates/specs/specification.feature specs/cli/init-project/specification.feature
```

## Step 2: Replace Placeholders

Open the file and replace:

| Placeholder                  | Replace With               | Example                     |
| ---------------------------- | -------------------------- | --------------------------- |
| `[module-name_feature-name]` | Feature name in kebab-case | `cli_init-project`          |
| `[role]`                     | User role                  | `developer`                 |
| `[capability]`               | What user wants            | `initialize a new project`  |
| `[business value]`           | Why it matters             | `start development quickly` |

### Before

```gherkin
Feature: [module-name_feature-name]

  As a [role]
  I want to [capability]
  So that [business value]
```

### After

```gherkin
Feature: cli_init-project

  As a developer
  I want to initialize a new project
  So that I can start development quickly
```

## Step 3: Add Rules from Example Mapping

Convert Blue Cards to Rule blocks:

```gherkin
Rule: Creates project directory structure

Rule: Configuration file contains required fields

Rule: Existing projects are not overwritten
```

## Step 4: Add Scenarios for Each Rule

Convert Green Cards to Scenario blocks:

```gherkin
Rule: Creates project directory structure

  @ov
  Scenario: Initialize in empty directory
    Given I am in an empty folder
    When I run "r2r init"
    Then a file named "r2r.yaml" should be created
    And a directory named "specs/" should be created

  @ov
  Scenario: Initialize in existing project
    Given I am in a directory with "r2r.yaml"
    When I run "r2r init"
    Then the command should fail
    And I should see "Project already initialized"
```

## Step 5: Add Required Tags

Every scenario MUST have:

1. **Verification tag**: `@ov`, `@iv`, `@pv`, `@piv`, or `@ppv`
2. **Acceptance criteria tag**: `@ac1`, `@ac2`, etc.

```gherkin
@ov @ac1
Scenario: Initialize in empty directory
```

## File Structure

```text
specs/
└── cli/
    └── init-project/
        └── specification.feature
```

## Validation

```bash
# Validate specification syntax
eac validate specs

# Run the specification
eac test cli --suite pre-commit
```

## Related

- [Feature Naming](../../../reference/specifications/feature-naming.md)
- [Gherkin Limits](../../../reference/specifications/gherkin-limits.md)
- [Example Mapping](running-example-mapping.md)
