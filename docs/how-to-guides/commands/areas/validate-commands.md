<!-- EDITOR
# Editor: how-to-guides/commands/areas/validate-commands.md

## Soul

Command reference for validation system covering seven validate subcommands (contracts, dependencies, test-tags, module-hierarchy, module-files, markdown, go-tidy) with success/failure output examples.

## Sections

1. Quick Reference
2. validate contracts
   - Synopsis
   - Description
   - Files Validated
   - Output
   - Common Errors
   - Exit Codes
3. validate dependencies
4. validate test-tags
   - Pattern Tags Validated
5. validate module-hierarchy
   - What It Validates
6. validate module-files
7. validate markdown
8. validate go-tidy
9. Common Workflows
   - Local Development
   - CI/CD Quality Gate
   - Pre-commit Hook
   - Post-Refactor
10. Integration Patterns
    - GitHub Actions
    - Makefile
11. Troubleshooting
12. Related Documentation
-->

# Validate Commands

Command reference for EAC's validation system.

## Quick Reference

| Command                     | Description                                        |
| --------------------------- | -------------------------------------------------- |
| `validate contracts`        | Validate repository contracts against JSON schemas |
| `validate dependencies`     | Validate module dependencies match contracts       |
| `validate test-tags`        | Validate test tags are defined in contract         |
| `validate module-hierarchy` | Validate module dependency graph structure         |
| `validate module-files`     | Validate module file ownership                     |
| `validate markdown`         | Validate markdown file syntax                      |
| `validate go-tidy`          | Validate Go module dependencies are tidy           |

---

## validate contracts

Validate repository configuration files against JSON schemas.

### Synopsis

```bash
r2r eac validate contracts
```

### Description

Validates all contract files against their JSON schemas to ensure proper structure and values.

### Files Validated

| File               | Description                           |
| ------------------ | ------------------------------------- |
| `modules.yml`      | Module contracts and metadata         |
| `environments.yml` | Environment definitions               |
| `testing-tags.yml` | Test tag definitions and skip reasons |
| `test-suites.yml`  | Test suite configurations             |

### Output

```text
Validating repository contracts...

  modules.yml               OK (23 modules)
  environments.yml          OK (4 environments)
  testing-tags.yml          OK (15 tags, 5 skip reasons)
  test-suites.yml           OK (3 suites)

All 4 contracts validated successfully
```

### Common Errors

- Missing required fields
- Invalid enum values
- Type mismatches (string vs number)
- Schema constraint violations

### Exit Codes

| Code | Description         |
| ---- | ------------------- |
| 0    | All contracts valid |
| 1    | Validation failed   |

---

## validate dependencies

Validate that go.mod files match dependency declarations in module contracts.

### Synopsis

```bash
r2r eac validate dependencies
```

### Description

Checks that actual dependencies in go.mod files match the `depends_on` declarations in module contracts.

### Output (Success)

```text
=== Dependency Validation Report ===

Modules checked: 15
Dependencies validated: 127

All dependencies match their contracts
```

### Output (Failure)

```text
=== Dependency Validation Report ===

Module 'src-api' depends on 'eac-core' but contract doesn't declare it

Fix: Add dependency to module contract
```

### Common Errors

- Missing dependency in contract
- Extra dependency not declared
- Version mismatches

### Exit Codes

| Code | Description               |
| ---- | ------------------------- |
| 0    | All dependencies match    |
| 1    | Dependency mismatch found |

---

## validate test-tags

Validate that all tags used in Gherkin feature files are defined in the tag contract.

### Synopsis

```bash
r2r eac validate test-tags
```

### Description

Scans all feature files for tags and verifies they are defined in `testing-tags.yml`. Also validates pattern tags against their respective contracts.

### Pattern Tags Validated

| Pattern          | Validates Against          |
| ---------------- | -------------------------- |
| `@skip:<reason>` | `skip_reasons` in contract |
| `@deps:<name>`   | `system-dependencies.yml`  |
| `@env:<moniker>` | `environments.yml`         |
| `@depm:<module>` | `modules.yml`              |

### Output (Success)

```text
All test tags are defined in the tag contract
   Validated 47 feature files
   Contract defines 15 valid tags
```

### Output (Failure)

```text
Found 2 undefined tag(s) used in 3 location(s):

  @performance (used in 1 file(s)):
    - specs/core/performance.feature:3

  @experimental (used in 2 file(s)):
    - specs/api/experimental.feature:1
    - specs/ui/experimental.feature:5

Fix: Add missing tags to .r2r/eac/testing-tags.yml
```

### Exit Codes

| Code | Description          |
| ---- | -------------------- |
| 0    | All tags defined     |
| 1    | Undefined tags found |

---

## validate module-hierarchy

Validate the module dependency graph structure for cycles and invalid references.

### Synopsis

```bash
r2r eac validate module-hierarchy
```

### Description

Analyzes the module dependency graph to detect structural issues.

### What It Validates

- No circular dependencies
- All referenced modules exist
- Bidirectional consistency
- All modules are reachable

### Output (Success)

```text
=== Module Hierarchy Validation Report ===

All module hierarchy checks passed!
```

### Output (Failure)

```text
=== Module Hierarchy Validation Report ===

References to Non-Existent Modules (1):
  Module 'src-api' depends on 'src-common', but 'src-common' does not exist

Circular Dependencies (1):
  Circular dependency: src-auth -> eac-core -> src-utils -> src-auth
```

### Common Errors

- Circular dependencies between modules
- References to non-existent modules
- Unreachable modules (orphaned)

### Exit Codes

| Code | Description             |
| ---- | ----------------------- |
| 0    | Hierarchy valid         |
| 1    | Structural issues found |

---

## validate module-files

Validate that files belong to exactly one module and are properly owned.

### Synopsis

```bash
r2r eac validate module-files
```

### Description

Checks file ownership rules are satisfied.

### What It Validates

- Each file belongs to exactly one module
- No files in the "unordered" catch-all module
- No overlapping glob patterns

### Output (Success)

```text
=== Module File Ownership Validation Report ===

All module file ownership checks passed!
```

### Output (Failure)

```text
=== Module File Ownership Validation Report ===

Files in Unordered Module (3):
   These files should be claimed by a proper module:
  - src/shared/util.go
  - src/shared/helper.go
  - src/shared/constants.go

   Fix: Create or update module contracts to claim these files.

Files with Multi-Module Ownership (1):
   Each file should belong to exactly one module:
  - go/eac/core/types.go
    Claimed by: eac-core, src-api

   Fix: Adjust module contract glob patterns to prevent overlap.
```

### Common Errors

- Files not claimed by any module (unordered)
- Files claimed by multiple modules (overlapping patterns)

### Exit Codes

| Code | Description            |
| ---- | ---------------------- |
| 0    | Ownership valid        |
| 1    | Ownership issues found |

---

## validate markdown

Validate markdown file syntax and structure.

### Synopsis

```bash
r2r eac validate markdown
```

### Description

Checks markdown files for syntax errors and structural issues.

### What It Validates

- Valid markdown syntax
- Proper heading hierarchy (no skipped levels)
- Valid code blocks (JSON, YAML syntax)
- Proper link syntax

### Output (Success)

```text
=== Markdown Validation Report ===

All markdown files are valid
   Validated 42 files
```

### Output (Failure)

```text
=== Markdown Validation Report ===

Validation errors in 2 file(s):

  docs/guide.md:
    - Line 15: Heading level skipped (h1 -> h3)
    - Line 42: Invalid JSON in code block

  README.md:
    - Line 5: Malformed link syntax
```

### Common Errors

- Skipped heading levels (h1 -> h3)
- Invalid JSON/YAML in code blocks
- Malformed links or references

### Exit Codes

| Code | Description         |
| ---- | ------------------- |
| 0    | Markdown valid      |
| 1    | Syntax errors found |

---

## validate go-tidy

Validate that Go module dependencies are tidy (go.mod and go.sum are synchronized).

### Synopsis

```bash
r2r eac validate go-tidy
```

### Description

Runs `go mod tidy -diff` on all Go modules to check synchronization.

### Output (Success)

```text
=== Go Module Tidy Validation Report ===

Total Go modules: 8
Tidy modules: 8
Untidy modules: 0

All Go modules have tidy dependencies!
```

### Output (Failure)

```text
=== Go Module Tidy Validation Report ===

Total Go modules: 8
Tidy modules: 7
Untidy modules: 1

Modules with untidy dependencies:

  - go/eac/commands
    Diff:
    -  github.com/example/old v1.0.0
    +  github.com/example/new v2.0.0

To fix, run: go mod tidy
```

### Exit Codes

| Code | Description          |
| ---- | -------------------- |
| 0    | All modules tidy     |
| 1    | Untidy modules found |

---

## Common Workflows

### Local Development

```bash
# Before committing
r2r eac validate contracts
r2r eac validate go-tidy

# Full validation
r2r eac validate contracts && \
r2r eac validate dependencies && \
r2r eac validate test-tags
```

### CI/CD Quality Gate

```bash
# Run all validations
for cmd in contracts dependencies test-tags module-hierarchy module-files markdown go-tidy; do
  echo "Validating $cmd..."
  r2r eac validate $cmd || exit 1
done
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

### Post-Refactor

```bash
# After major restructuring
r2r eac validate module-hierarchy
r2r eac validate module-files
r2r eac validate dependencies
```

---

## Integration Patterns

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

### Makefile

```makefile
.PHONY: validate validate-contracts validate-deps validate-tags

validate: validate-contracts validate-deps validate-tags validate-go-tidy

validate-contracts:
  r2r eac validate contracts

validate-deps:
  r2r eac validate dependencies

validate-tags:
  r2r eac validate test-tags

validate-go-tidy:
  r2r eac validate go-tidy
```

---

## Troubleshooting

| Problem                   | Solution                                     |
| ------------------------- | -------------------------------------------- |
| Contract validation fails | Check JSON schema errors, verify YAML syntax |
| Dependency mismatch       | Update module contract or fix go.mod file    |
| Undefined test tags       | Add tags to testing-tags.yml                 |
| Circular dependency       | Refactor modules to break cycle              |
| Files unordered           | Update module glob patterns                  |
| Markdown errors           | Fix syntax, heading hierarchy, code blocks   |
| Go not tidy               | Run `go mod tidy` in module directory        |
| Schema not found          | Ensure .r2r/eac/schemas/ directory exists    |

---

## Related Documentation

- [Validate Overview](validate-overview.md) - Validation concepts
- [Validate Configuration](validate-configuration.md) - Configuration options
- [Test Commands](test-commands.md) - Run tests after validating
