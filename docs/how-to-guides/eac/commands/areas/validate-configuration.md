<!-- EDITOR
# Editor: how-to-guides/commands/areas/validate-configuration.md

## Soul

Configuration reference for validation system including contract schemas (modules, environments, testing-tags), pattern tag validation rules, file ownership patterns, markdown rules, and JSON Schema format.

## Sections

1. Contract Configuration
   - Module Contract Schema
   - Supported Module Types
   - Environment Contract
2. Test Tag Configuration
   - Tag Contract
   - Pattern Tags
   - Skip Reason Configuration
3. Test Suite Configuration
   - Suite Definition
4. File Ownership Configuration
   - Module File Patterns
   - Pattern Rules
5. Markdown Validation Configuration
   - Validated Rules
   - Example Valid Structure
6. Go Tidiness Configuration
   - What Gets Validated
   - Validation Command
7. CI/CD Configuration
   - GitHub Actions
   - Pre-commit Hook
   - Makefile Integration
8. Schema Files
   - Schema Location
   - Schema Format
9. Troubleshooting
10. Related Documentation
-->

# Validate Configuration

This guide covers configuration options for EAC's validation system, including contract schemas, tag patterns, and file ownership rules.

## Contract Configuration

### Module Contract Schema

```yaml
# .r2r/eac/modules.yml
modules:
  - moniker: eac-core
    type: go
    description: Core EAC functionality
    files:
      root: go/eac/core
      patterns:
        - "**/*.go"
        - "go.mod"
        - "go.sum"
    depends_on:
      - eac-config
```

### Supported Module Types

| Type         | Description                       | Validated              |
| ------------ | --------------------------------- | ---------------------- |
| `go`         | Go module (library, exe, or test) | Dependencies, tidiness |
| `container`  | Docker container module           | Dockerfile             |
| `typescript` | TypeScript/npm module             | Package.json           |
| `static`     | Static files (no build)           | File existence         |

### Environment Contract

```yaml
# .r2r/eac/environments.yml
environments:
  - moniker: local
    description: Local development
    type: development

  - moniker: staging
    description: Staging environment
    type: staging

  - moniker: prod
    description: Production environment
    type: production
```

## Test Tag Configuration

### Tag Contract

```yaml
# .r2r/eac/testing-tags.yml
tags:
  - tag: "@commit"
    description: "Tests that run on every commit"

  - tag: "@integration"
    description: "Integration tests requiring external services"

  - tag: "@e2e"
    description: "End-to-end tests"

  - tag: "@smoke"
    description: "Quick health check tests"

  - tag: "@slow"
    description: "Tests that take longer than 5 seconds"

skip_reasons:
  - code: "flaky"
    description: "Test is flaky and needs investigation"

  - code: "wip"
    description: "Work in progress"

  - code: "ci-only"
    description: "Only runs in CI environment"
```

### Pattern Tags

Pattern tags are validated against their respective contracts:

| Pattern          | Validates Against                  | Example          |
| ---------------- | ---------------------------------- | ---------------- |
| `@skip:<reason>` | `skip_reasons` in testing-tags.yml | `@skip:flaky`    |
| `@deps:<name>`   | system-dependencies.yml            | `@deps:docker`   |
| `@env:<moniker>` | environments.yml                   | `@env:staging`   |
| `@depm:<module>` | modules.yml                        | `@depm:eac-core` |

### Skip Reason Configuration

```yaml
skip_reasons:
  - code: "flaky"
    description: "Test is flaky and needs investigation"

  - code: "wip"
    description: "Work in progress"

  - code: "ci-only"
    description: "Only runs in CI environment"

  - code: "manual"
    description: "Requires manual verification"
```

## Test Suite Configuration

### Suite Definition

```yaml
# .r2r/eac/test-suites.yml
suites:
  - name: commit
    description: Pre-commit validation tests
    tags:
      - "@commit"
      - "@unit"
    timeout: 300

  - name: integration
    description: Integration tests
    tags:
      - "@integration"
    timeout: 900

  - name: e2e
    description: End-to-end tests
    tags:
      - "@e2e"
    timeout: 1800
```

## File Ownership Configuration

### Module File Patterns

```yaml
modules:
  - moniker: eac-commands
    files:
      root: go/eac/commands
      patterns:
        - "**/*.go"
        - "go.mod"
        - "go.sum"

  - moniker: docs
    files:
      root: docs
      patterns:
        - "**/*.md"
        - "**/*.yml"
```

### Pattern Rules

- Each file should belong to exactly one module
- Patterns are glob-style (supports `**`, `*`, `?`)
- Root is relative to repository root
- No overlapping patterns between modules

## Markdown Validation Configuration

### Validated Rules

| Rule              | Description                      |
| ----------------- | -------------------------------- |
| Heading hierarchy | No skipped levels (h1 → h3)      |
| Code block syntax | Valid JSON/YAML in fenced blocks |
| Link syntax       | Proper markdown link format      |
| List formatting   | Consistent list markers          |

### Example Valid Structure

```markdown
# Main Title (h1)

## Section (h2)

### Subsection (h3)

Content here...

\```json
{"valid": "json"}
\```

```

## Go Tidiness Configuration

### What Gets Validated

- go.mod file syntax
- go.sum synchronization
- Dependency version consistency
- No missing dependencies
- No unused dependencies

### Validation Command

Internally runs:

```bash
go mod tidy -diff
````

Returns non-zero if any changes would be made.

## CI/CD Configuration

### GitHub Actions

```yaml
name: Validate

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Validate contracts
        run: r2r eac validate contracts

      - name: Validate dependencies
        run: r2r eac validate dependencies

      - name: Validate test tags
        run: r2r eac validate test-tags

      - name: Validate module hierarchy
        run: r2r eac validate module-hierarchy

      - name: Validate module files
        run: r2r eac validate module-files

      - name: Validate markdown
        run: r2r eac validate markdown

      - name: Validate Go tidiness
        run: r2r eac validate go-tidy
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running validations..."

r2r eac validate contracts || exit 1
r2r eac validate dependencies || exit 1
r2r eac validate test-tags || exit 1
r2r eac validate go-tidy || exit 1

echo "All validations passed"
```

### Makefile Integration

```makefile
.PHONY: validate validate-all

validate:
  r2r eac validate contracts
  r2r eac validate dependencies
  r2r eac validate go-tidy

validate-all:
  r2r eac validate contracts
  r2r eac validate dependencies
  r2r eac validate test-tags
  r2r eac validate module-hierarchy
  r2r eac validate module-files
  r2r eac validate markdown
  r2r eac validate go-tidy
```

## Schema Files

### Schema Location

```text
.r2r/eac/schemas/
├── modules.schema.json
├── environments.schema.json
├── testing-tags.schema.json
└── test-suites.schema.json
```

### Schema Format

Schemas use JSON Schema (Draft 7):

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["modules"],
  "properties": {
    "modules": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["moniker", "type"],
        "properties": {
          "moniker": {"type": "string"},
          "type": {"enum": ["go", "container", "typescript", "static"]}
        }
      }
    }
  }
}
```

## Troubleshooting

| Issue            | Cause                       | Solution                          |
| ---------------- | --------------------------- | --------------------------------- |
| Schema not found | Missing schemas directory   | Ensure `.r2r/eac/schemas/` exists |
| Invalid enum     | Wrong module type           | Check supported types             |
| Pattern overlap  | Multiple modules claim file | Adjust glob patterns              |
| Tag undefined    | Missing from contract       | Add to testing-tags.yml           |
| Go not tidy      | go.mod out of sync          | Run `go mod tidy`                 |

## Related Documentation

- [Validate Overview](validate-overview.md) - Validation concepts
- [Validate Commands](validate-commands.md) - Command reference
- [Test Configuration](test-configuration.md) - Test tag integration
