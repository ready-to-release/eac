# Specifications Configuration

This guide covers configuration options for EAC's BDD specification system, including AI prompts, tag contracts, and validation rules.

## Configuration Files

| File                        | Purpose                          |
| --------------------------- | -------------------------------- |
| `.r2r/eac/testing-tags.yml` | Tag definitions and skip reasons |
| `.r2r/eac/ai/specs/`        | AI prompt templates              |
| `.r2r/eac/specs/config.yml` | Specification settings           |

## Tag Configuration

### Testing Tags Contract

`.r2r/eac/testing-tags.yml`:

```yaml
# Tag definitions
tags:
  # Test categories
  - tag: "@smoke"
    description: "Quick sanity tests, run on every commit"

  - tag: "@integration"
    description: "Integration tests requiring external services"

  - tag: "@e2e"
    description: "End-to-end tests covering full workflows"

  - tag: "@unit"
    description: "Unit tests for isolated components"

  - tag: "@performance"
    description: "Performance and load tests"

  # Feature flags
  - tag: "@wip"
    description: "Work in progress, not ready for CI"

  - tag: "@manual"
    description: "Requires manual execution"

# Skip reasons
skip_reasons:
  - code: "not-implemented"
    description: "Feature not yet implemented"

  - code: "flaky"
    description: "Test is flaky, needs investigation"

  - code: "external-dependency"
    description: "Depends on external service unavailable in CI"

  - code: "windows-only"
    description: "Only runs on Windows platform"

  - code: "linux-only"
    description: "Only runs on Linux platform"
```

### Pattern Tags

Special tags with dynamic values:

```yaml
# Pattern tag definitions
pattern_tags:
  # Skip with reason
  - pattern: "@skip:{reason}"
    description: "Skip test with specified reason"
    validates_against: skip_reasons

  # Environment requirement
  - pattern: "@env:{environment}"
    description: "Requires specific environment"
    validates_against: environments.yml

  # Module dependency
  - pattern: "@depm:{module}"
    description: "Depends on specific module"
    validates_against: modules.yml

  # System dependency
  - pattern: "@deps:{dependency}"
    description: "Requires system dependency"
    validates_against: system-dependencies.yml
```

### Tag Usage Examples

```gherkin
@smoke @auth
Feature: User Authentication

  @integration
  Scenario: Successful login
    Given a registered user
    When they provide valid credentials
    Then they should be authenticated

  @skip:not-implemented
  Scenario: MFA login
    Given a user with MFA enabled
    When they provide valid credentials
    Then they should receive MFA challenge

  @env:staging @deps:postgres
  Scenario: Database authentication
    Given a database connection
    When credentials are validated
    Then access should be granted
```

## AI Configuration

### Prompt Templates

Location: `.r2r/eac/ai/specs/`

```text
.r2r/eac/ai/specs/
├── create-feature.md       # Feature file generation
├── create-scenario.md      # Individual scenario generation
├── analyze-requirements.md # Requirements analysis
└── suggest-tags.md         # Tag suggestion
```

### Feature Generation Prompt

```markdown
# Gherkin Feature Generation

## Context
Generate a Gherkin feature file from natural language requirements.

## Requirements
{{.Requirements}}

## Project Context
- Module: {{.Module.Moniker}}
- Existing Features: {{.ExistingFeatures}}
- Available Tags: {{.Tags}}

## Guidelines

### Structure
1. Start with Feature keyword and clear title
2. Add feature description explaining the capability
3. Group scenarios under Rules when appropriate
4. Use Background for common setup

### Scenarios
1. One behavior per scenario
2. Use Given/When/Then format
3. Keep steps declarative, not imperative
4. Use concrete examples in steps

### Tags
1. Add appropriate category tags (@smoke, @integration, etc.)
2. Use @skip:reason for unimplemented scenarios
3. Add @env:name for environment-specific tests

### Language
1. Use business language, not technical jargon
2. Write from user perspective
3. Be specific but concise
4. Avoid implementation details

## Output
Complete feature file content.
```

### Scenario Generation Prompt

```markdown
# Scenario Generation

## Feature Context
{{.FeatureTitle}}
{{.FeatureDescription}}

## Existing Scenarios
{{range .ExistingScenarios}}
- {{.Name}}
{{end}}

## New Requirement
{{.Requirement}}

## Guidelines
1. Follow existing feature style
2. Ensure scenario is unique
3. Use appropriate tags
4. Include edge cases if relevant

## Output
Single scenario with steps.
```

## Specification Settings

### Global Settings

`.r2r/eac/specs/config.yml`:

```yaml
# Directory settings
paths:
  # Root directory for specs
  root: specs/

  # Pattern for feature files
  pattern: "**/*.feature"

  # Step definitions location
  steps: go/*/steps/

# Validation settings
validation:
  # Require feature description
  require_description: true

  # Require scenario descriptions
  require_scenario_description: false

  # Maximum scenarios per feature
  max_scenarios: 20

  # Maximum steps per scenario
  max_steps: 10

  # Require at least one tag
  require_tags: true

# Generation settings
generation:
  # Default tags for new features
  default_tags:
    - "@wip"

  # Include example scenarios
  include_examples: true

  # Generate data tables
  generate_tables: true
```

### Module-Specific Settings

In module contracts:

```yaml
# modules.yml
modules:
  - moniker: src-auth
    type: go-library
    specs:
      path: specs/src-auth/
      default_tags:
        - "@auth"
      require_integration_tag: true
```

## Step Definition Configuration

### Step Registry

```yaml
# .r2r/eac/specs/steps.yml
steps:
  # Common step patterns
  patterns:
    - pattern: "a user is logged in"
      implementation: "steps/auth/login_steps.go"

    - pattern: "the response status is {int}"
      implementation: "steps/api/response_steps.go"

  # Step aliases
  aliases:
    - from: "I am logged in"
      to: "a user is logged in"

    - from: "the status code is {int}"
      to: "the response status is {int}"
```

### Unused Step Detection

```yaml
# Configuration for unused step detection
unused_steps:
  # Directories to scan for step definitions
  step_dirs:
    - go/*/steps/
    - test/steps/

  # Feature directories to scan
  feature_dirs:
    - specs/

  # Ignore patterns
  ignore:
    - "*_helper.go"
    - "common_steps.go"
```

## Test Suite Configuration

### Suite Definitions

`.r2r/eac/test-suites.yml`:

```yaml
suites:
  - name: smoke
    description: "Quick sanity tests"
    tags:
      include: ["@smoke"]
      exclude: ["@skip:*", "@wip"]
    timeout: 5m
    parallel: true

  - name: integration
    description: "Integration tests"
    tags:
      include: ["@integration"]
      exclude: ["@skip:*"]
    timeout: 30m
    parallel: false
    environment:
      requires: ["postgres", "redis"]

  - name: e2e
    description: "End-to-end tests"
    tags:
      include: ["@e2e"]
      exclude: ["@skip:*", "@manual"]
    timeout: 60m
    parallel: false

  - name: commit
    description: "Tests to run on every commit"
    tags:
      include: ["@smoke", "@unit"]
      exclude: ["@skip:*", "@wip", "@manual"]
    timeout: 10m
    parallel: true
```

### Suite Selection

```bash
# Run specific suite
r2r eac test --suite smoke

# Run multiple suites
r2r eac test --suite smoke --suite unit
```

## Environment Configuration

### Environment Requirements

Reference environments from `environments.yml`:

```yaml
# environments.yml
environments:
  - moniker: local
    description: "Local development"
    variables:
      DB_HOST: localhost
      DB_PORT: 5432

  - moniker: staging
    description: "Staging environment"
    variables:
      DB_HOST: staging-db.example.com
      DB_PORT: 5432

  - moniker: production
    description: "Production environment"
    variables:
      DB_HOST: prod-db.example.com
      DB_PORT: 5432
```

### Using Environment Tags

```gherkin
@env:staging
Scenario: Staging-specific test
  Given the staging database is available
  ...

@env:local @env:staging
Scenario: Works in local and staging
  Given any database is available
  ...
```

## Validation Rules

### Feature Validation

```bash
r2r eac validate specs
```

Validates:

- Gherkin syntax correctness
- Required elements present
- Tag definitions exist
- Step patterns match

### Tag Validation

```bash
r2r eac validate test-tags
```

Validates:

- All tags defined in contract
- Skip reasons are valid
- Environment references exist
- Module references exist

### Common Validation Errors

| Error                 | Cause                     | Fix                     |
| --------------------- | ------------------------- | ----------------------- |
| `Undefined tag`       | Tag not in contract       | Add to testing-tags.yml |
| `Invalid skip reason` | Unknown skip code         | Add to skip_reasons     |
| `Missing description` | Feature lacks description | Add description         |
| `Too many scenarios`  | Exceeds max_scenarios     | Split feature file      |

## CI Integration

### GitHub Actions

```yaml
- name: Validate specifications
  run: r2r eac validate specs

- name: Validate test tags
  run: r2r eac validate test-tags

- name: Check unused steps
  run: r2r eac get specs unused-steps

- name: Run smoke tests
  run: r2r eac test --suite smoke
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Validate specs
r2r eac validate specs || exit 1

# Validate tags
r2r eac validate test-tags || exit 1

echo "✅ Specification validation passed"
```

## Example Configurations

### Minimal Configuration

```yaml
# testing-tags.yml
tags:
  - tag: "@smoke"
    description: "Quick tests"
  - tag: "@integration"
    description: "Integration tests"

skip_reasons:
  - code: "not-implemented"
    description: "Not yet implemented"
```

### Comprehensive Configuration

```yaml
# testing-tags.yml
tags:
  # Categories
  - tag: "@smoke"
    description: "Quick sanity tests (< 1 min total)"
  - tag: "@unit"
    description: "Unit tests for isolated components"
  - tag: "@integration"
    description: "Integration tests with external services"
  - tag: "@e2e"
    description: "End-to-end user journey tests"
  - tag: "@performance"
    description: "Performance and load tests"
  - tag: "@security"
    description: "Security-focused tests"

  # Priorities
  - tag: "@critical"
    description: "Critical path tests, must always pass"
  - tag: "@regression"
    description: "Regression tests for known issues"

  # Status
  - tag: "@wip"
    description: "Work in progress"
  - tag: "@manual"
    description: "Requires manual execution"

skip_reasons:
  - code: "not-implemented"
    description: "Feature not yet implemented"
  - code: "blocked"
    description: "Blocked by external dependency"
  - code: "flaky"
    description: "Test is flaky, needs investigation"
  - code: "infrastructure"
    description: "Infrastructure not available"
  - code: "windows-only"
    description: "Only applicable on Windows"
  - code: "linux-only"
    description: "Only applicable on Linux"
  - code: "deprecated"
    description: "Feature deprecated, test kept for reference"

pattern_tags:
  - pattern: "@skip:{reason}"
    validates_against: skip_reasons
  - pattern: "@env:{environment}"
    validates_against: environments.yml
  - pattern: "@depm:{module}"
    validates_against: modules.yml
  - pattern: "@deps:{dependency}"
    validates_against: system-dependencies.yml
  - pattern: "@jira:{ticket}"
    description: "Links to JIRA ticket"
```

## Related Documentation

- [Specifications Overview](specifications-overview.md) - Concepts and workflows
- [Specifications Commands](specifications-commands.md) - Command reference
- [Validate Commands](validate-overview.md) - Tag validation
