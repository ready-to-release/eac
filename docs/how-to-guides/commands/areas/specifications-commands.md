<!-- EDITOR
# Editor: how-to-guides/commands/areas/specifications-commands.md

## Soul

Command reference for BDD specifications covering create spec (AI generation), validate specs (quality checks), and get specs unused-steps (dead code detection).

## Sections

1. Quick Reference
2. create spec
   - Synopsis
   - Description
   - Arguments
   - Flags
   - Examples
   - Output
   - Exit Codes
3. validate specs
   - Synopsis
   - Description
   - Flags
   - Examples
   - Output (Success)
   - Output (Failures)
   - Exit Codes
4. get specs unused-steps
   - Synopsis
   - Description
   - Flags
   - Examples
   - Output
   - JSON Output
   - Exit Codes
5. Common Workflows
   - TDD Workflow
   - Spec Review
   - CI Integration
   - Generating Multiple Specs
6. Integration with Other Commands
   - With Validation
   - With Testing
   - With Risk Management
7. Related Documentation
-->

# Specifications Commands

Command reference for EAC's BDD specification system.

## Quick Reference

| Command              | Description                                                        |
| -------------------- | ------------------------------------------------------------------ |
| `create spec`        | Generate Gherkin specifications from natural language descriptions |
| `validate specs`     | Validate Gherkin specifications against quality contracts          |
| `get specs unused-steps` | Detect unused godog step definitions                               |

---

## create spec

Generate Gherkin feature files from natural language descriptions using AI.

### Synopsis

```bash
r2r eac create spec "<description>" [options]
```

### Description

Transforms natural language requirements into properly structured Gherkin feature files. The AI generates:

- Feature title and description
- Rules for grouping related scenarios
- Scenarios with Given/When/Then steps
- Appropriate tags based on content
- Example data tables when relevant

### Arguments

| Argument      | Required | Description                                 |
| ------------- | -------- | ------------------------------------------- |
| `description` | Yes      | Natural language description of the feature |

### Flags

| Flag       | Short | Type     | Default           | Description                   |
| ---------- | ----- | -------- | ----------------- | ----------------------------- |
| `--module` | `-m`  | string   | -                 | Target module for the spec    |
| `--output` | `-o`  | string   | `specs/<module>/` | Output directory              |
| `--tags`   | `-t`  | string[] | -                 | Additional tags to include    |
| `--debug`  | `-d`  | bool     | `false`           | Save AI prompts and responses |

### Examples

```bash
# Simple feature
r2r eac create spec "Users can log in with email and password"

# Detailed requirements
r2r eac create spec "Users can reset their password via email. \
  They receive a link that expires in 24 hours. \
  The link can only be used once."

# Specify module
r2r eac create spec "API rate limiting" --module src-api

# Add tags
r2r eac create spec "User authentication" --tags "@auth" --tags "@security"

# Custom output
r2r eac create spec "Shopping cart" --output specs/e-commerce/

# Debug mode
r2r eac create spec "Payment processing" --debug
```

### Output

```text
Generating specification...

Analyzing requirements...
  ✓ Feature: User Authentication
  ✓ 2 rules identified
  ✓ 5 scenarios generated

Generated specification:
────────────────────────────────────────
@auth @integration
Feature: User Authentication
  Users can securely log into the system using their credentials.

  Rule: Valid credentials grant access

    Scenario: Successful login with valid email and password
      Given a registered user with email "user@example.com"
      And the user has password "SecurePass123"
      When the user submits login credentials
      Then the user should be authenticated
      And a session token should be issued

    Scenario: Failed login with incorrect password
      Given a registered user with email "user@example.com"
      When the user submits password "WrongPassword"
      Then authentication should fail
      And an error message "Invalid credentials" should be displayed
────────────────────────────────────────

✓ Saved: specs/src-auth/user-authentication.feature
```

### Exit Codes

| Code | Description                          |
| ---- | ------------------------------------ |
| 0    | Specification generated successfully |
| 1    | Error generating specification       |
| 2    | Invalid module specified             |
| 3    | AI generation failed                 |

---

## validate specs

Validate Gherkin specifications against quality contracts.

### Synopsis

```bash
r2r eac validate specs [options]
```

### Description

Validates all Gherkin feature files for:

- Syntax correctness
- Required elements (Feature, Scenario)
- Tag definitions in contract
- Step structure (Given/When/Then)
- Naming conventions

### Flags

| Flag       | Short | Type   | Default | Description                   |
| ---------- | ----- | ------ | ------- | ----------------------------- |
| `--module` | `-m`  | string | -       | Validate specific module only |
| `--strict` |       | bool   | `false` | Fail on warnings              |
| `--fix`    |       | bool   | `false` | Auto-fix simple issues        |

### Examples

```bash
# Validate all specs
r2r eac validate specs

# Validate specific module
r2r eac validate specs --module src-auth

# Strict mode
r2r eac validate specs --strict

# Auto-fix issues
r2r eac validate specs --fix
```

### Output (Success)

```text
Validating specifications...

Scanning feature files...
  ✓ specs/src-auth/login.feature
  ✓ specs/src-auth/logout.feature
  ✓ specs/src-api/users.feature
  ✓ specs/src-api/products.feature

Validation Results:
  Files validated: 4
  Scenarios: 23
  Steps: 89

✓ All specifications valid
```

### Output (Failures)

```text
Validating specifications...

Scanning feature files...
  ✓ specs/src-auth/login.feature
  ✗ specs/src-auth/password-reset.feature
  ⚠ specs/src-api/users.feature
  ✓ specs/src-api/products.feature

Errors:
────────────────────────────────────────
specs/src-auth/password-reset.feature:
  Line 5: Missing Feature description
  Line 12: Scenario has no Then step
  Line 18: Undefined tag @experimental

Warnings:
────────────────────────────────────────
specs/src-api/users.feature:
  Line 1: Feature description exceeds 200 characters
  Line 25: Scenario name could be more descriptive

Validation Results:
  Files validated: 4
  Errors: 3
  Warnings: 2

✗ Validation failed
```

### Exit Codes

| Code | Description                                   |
| ---- | --------------------------------------------- |
| 0    | Validation passed                             |
| 1    | Validation failed (errors)                    |
| 2    | Validation passed with warnings (strict mode) |

---

## get specs unused-steps

Detect step definitions that are not used by any feature file.

### Synopsis

```bash
r2r eac get specs unused-steps [options]
```

### Description

Scans step definition files and feature files to identify:

- Step definitions not matched by any feature
- Potentially dead code
- Duplicate step patterns

Helps maintain clean test code by identifying unused steps for removal.

### Flags

| Flag                | Short | Type   | Default | Description                |
| ------------------- | ----- | ------ | ------- | -------------------------- |
| `--module`          | `-m`  | string | -       | Check specific module only |
| `--include-helpers` |       | bool   | `false` | Include helper functions   |
| `--json`            |       | bool   | `false` | Output as JSON             |

### Examples

```bash
# Check all modules
r2r eac get specs unused-steps

# Check specific module
r2r eac get specs unused-steps --module src-auth

# Include helper functions
r2r eac get specs unused-steps --include-helpers

# JSON output
r2r eac get specs unused-steps --json
```

### Output

```text
Scanning for unused step definitions...

Step Definitions Scanned:
  go/eac/auth/steps/: 15 steps
  go/eac/api/steps/: 23 steps
  go/eac/core/steps/: 8 steps

Feature Files Scanned:
  specs/: 12 features, 45 scenarios

Unused Step Definitions:
────────────────────────────────────────

go/eac/auth/steps/auth_steps.go:
  Line 45: "^the user is an administrator$"
  Line 67: "^the session has expired$"
  Line 89: "^two-factor authentication is enabled$"

go/eac/api/steps/api_steps.go:
  Line 123: "^the API rate limit is (\d+) requests per minute$"

Summary:
  Total steps: 46
  Used steps: 42
  Unused steps: 4 (8.7%)

Consider removing unused steps or adding feature coverage.
```

### JSON Output

```json
{
  "total_steps": 46,
  "used_steps": 42,
  "unused_steps": [
    {
      "file": "go/eac/auth/steps/auth_steps.go",
      "line": 45,
      "pattern": "^the user is an administrator$",
      "function": "theUserIsAnAdministrator"
    },
    {
      "file": "go/eac/auth/steps/auth_steps.go",
      "line": 67,
      "pattern": "^the session has expired$",
      "function": "theSessionHasExpired"
    }
  ]
}
```

### Exit Codes

| Code | Description           |
| ---- | --------------------- |
| 0    | No unused steps found |
| 1    | Unused steps found    |
| 2    | Error scanning files  |

---

## Common Workflows

### TDD Workflow

```bash
# 1. Generate spec from requirements
r2r eac create spec "User can update profile picture"

# 2. Validate the generated spec
r2r eac validate specs

# 3. Run tests (should fail - no implementation)
r2r eac test src-users

# 4. Implement step definitions
# ... write code ...

# 5. Run tests (should pass)
r2r eac test src-users

# 6. Check for unused steps
r2r eac get specs unused-steps
```

### Spec Review

```bash
# Validate all specs
r2r eac validate specs

# Check for undefined tags
r2r eac validate test-tags

# Find unused steps
r2r eac get specs unused-steps

# Review and clean up
```

### CI Integration

```bash
# In CI pipeline
r2r eac validate specs --strict || exit 1
r2r eac validate test-tags || exit 1

# Warn on unused steps (don't fail)
r2r eac get specs unused-steps || echo "Warning: unused steps found"

# Run tests
r2r eac test --suite commit
```

### Generating Multiple Specs

```bash
# Generate specs for multiple features
for feature in "User login" "User logout" "Password reset"; do
  r2r eac create spec "$feature" --module src-auth
done

# Validate all
r2r eac validate specs --module src-auth
```

---

## Integration with Other Commands

### With Validation

```bash
# Validate specs structure
r2r eac validate specs

# Validate tags are defined
r2r eac validate test-tags

# Both should pass before commit
```

### With Testing

```bash
# Run spec tests
r2r eac test --suite integration

# Debug test failures
r2r eac test debug
```

### With Risk Management

```bash
# Specs provide evidence for controls
# Tag specs with control references

@control:AC-2
Feature: User Account Management
  ...

# Then run assessment
r2r eac create risk-assess
```

---

## Related Documentation

- [Specifications Overview](specifications-overview.md) - Concepts and workflows
- [Specifications Configuration](specifications-configuration.md) - Configuration reference
- [Validate Commands](validate-overview.md) - Tag validation
