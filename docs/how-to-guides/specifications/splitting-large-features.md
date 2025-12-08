# Splitting Large Features

How to refactor features that have grown too large.

## When to Split

| Indicator      | Threshold | Action             |
| -------------- | --------- | ------------------ |
| Rule count     | >10       | Must split         |
| Rule count     | 7-10      | Consider splitting |
| Scenario count | >30       | Must split         |
| Scenario count | 20-30     | Should split       |

## Strategy 1: Split by Rule

If you have 8+ Rules, create separate features by grouping related Rules.

### Before (12 Rules)

```text
Feature: project_validation
  Rule: File structure must be valid
  Rule: Configuration must be valid
  Rule: Dependencies must be valid
  Rule: Module boundaries must be respected
  Rule: Naming conventions must be followed
  Rule: Documentation must exist
  Rule: Tests must exist
  Rule: Coverage must meet threshold
  Rule: Performance benchmarks must pass
  Rule: Security checks must pass
  Rule: Linting must pass
  Rule: Type checking must pass
```

### After (3 features × 4 Rules)

```text
Feature: project_structure-validation
  Rule: File structure must be valid
  Rule: Configuration must be valid
  Rule: Module boundaries must be respected
  Rule: Naming conventions must be followed

Feature: project_quality-validation
  Rule: Tests must exist with minimum coverage
  Rule: Documentation must exist for public APIs
  Rule: Performance benchmarks must pass
  Rule: Linting must pass

Feature: project_security-validation
  Rule: Dependencies must have no known vulnerabilities
  Rule: Security checks must pass (SAST)
  Rule: Secrets must not be committed
  Rule: Authentication must be enforced
```

## Strategy 2: Split by Workflow

If scenarios cluster around different user workflows, split by workflow.

### Before (35 scenarios)

```text
Feature: user-management
  6 Rules covering registration, login, profile
```

### After (3 features)

```text
Feature: user-management_registration
  Rule: Users can create accounts
  Rule: Email verification required

Feature: user-management_authentication
  Rule: Users can log in
  Rule: Sessions expire correctly

Feature: user-management_profile
  Rule: Users can update profile
  Rule: Users can change password
```

## Strategy 3: Split by Scenario Type

If you have many error scenarios, separate them.

### Before (30 scenarios)

```text
Feature: config-validation
  15 happy path scenarios
  15 error scenarios
```

### After (2 features)

```text
Feature: config-validation
  15 happy path scenarios

Feature: config-validation_errors
  15 error scenarios
```

## Refactoring Steps

1. **Identify split points**: Group related Rules/Scenarios
2. **Create new feature files**: One per group
3. **Move content**: Cut and paste Rules with their Scenarios
4. **Update feature names**: Follow naming convention
5. **Update tags**: Ensure all scenarios have required tags
6. **Update step definitions**: Check for shared steps
7. **Run tests**: Verify nothing broke

## Naming After Split

Original: `project_validation`

Split into:

- `project_structure-validation`
- `project_quality-validation`
- `project_security-validation`

## Related

- [Gherkin Limits](../../reference/specifications/gherkin-limits.md)
- [Feature Naming](../../reference/specifications/feature-naming.md)
- [Creating Feature Files](creating-feature-files.md)
