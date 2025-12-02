# Validate Command

**Problem**: Maintaining repository integrity across contracts, dependencies, and configurations in a monorepo.

**Solution**: Use `validate` to verify all repository contracts, dependencies, and configurations meet quality standards.

## Key Benefits

- Automated contract validation against JSON schemas
- Dependency graph integrity checks
- File ownership and organizational validation
- CI/CD integration for quality gates
- Early detection of configuration issues

## Quick Start

```bash
# Validate all contracts against schemas
r2r eac validate contracts

# Validate module dependencies match contracts
r2r eac validate dependencies

# Validate test tags are defined
r2r eac validate test-tags

# Validate dependency graph structure
r2r eac validate module-hierarchy

# Validate file ownership
r2r eac validate module-files

# Validate markdown syntax
r2r eac validate markdown

# Validate Go modules are tidy
r2r eac validate go-tidy
```

## Command Reference

### Quick Reference Table

| Subcommand         | Purpose                | Validates                                                        |
| ------------------ | ---------------------- | ---------------------------------------------------------------- |
| `contracts`        | Schema validation      | modules.yml, environments.yml, testing-tags.yml, test-suites.yml |
| `dependencies`     | Dependency consistency | go.mod files match module contracts                              |
| `test-tags`        | Tag definitions        | All test tags defined in tag contract                            |
| `module-hierarchy` | Graph structure        | No cycles, valid references, reachability                        |
| `module-files`     | File ownership         | Single ownership, no unordered files                             |
| `markdown`         | Markdown syntax        | Valid syntax, heading hierarchy, code blocks                     |
| `go-tidy`          | Go modules             | Dependencies are tidy (go.mod/go.sum sync)                       |

All subcommands return exit code 0 on success, 1 on failure.

### validate contracts

Validate repository configuration files against JSON schemas.

```bash
r2r eac validate contracts
```

**What it validates:**

- `modules.yml` - Module contracts and metadata
- `environments.yml` - Environment definitions
- `testing-tags.yml` - Test tag definitions and skip reasons
- `test-suites.yml` - Test suite configurations

**Output:**

```text
Validating repository contracts...

  modules.yml               OK (23 modules)
  environments.yml          OK (4 environments)
  testing-tags.yml          OK (15 tags, 5 skip reasons)
  test-suites.yml           OK (3 suites)

All 4 contracts validated successfully
```

**Common errors:**

- Missing required fields
- Invalid enum values
- Type mismatches (string vs number)
- Schema constraint violations

### validate dependencies

Validate that go.mod files match dependency declarations in module contracts.

```bash
r2r eac validate dependencies
```

**What it validates:**

- Actual dependencies in go.mod files
- Contract declarations in modules.yml
- Consistency between implementation and contract

**Output:**

```text
=== Dependency Validation Report ===

Modules checked: 15
Dependencies validated: 127

✅ All dependencies match their contracts
```

**Common errors:**

- Missing dependency in contract
- Extra dependency not declared
- Version mismatches

**Fix:** Update module contract or fix go.mod file.

### validate test-tags

Validate that all tags used in Gherkin feature files are defined in the tag contract.

```bash
r2r eac validate test-tags
```

**What it validates:**

- Tags in feature files (@smoke, @integration, etc.)
- Pattern tags (@skip:\<reason>, @deps:\<name>, @env:\<moniker>, @depm:\<module>)
- Tag definitions in testing-tags.yml
- Skip reason codes
- Environment monikers
- Module names
- System dependencies

**Output (success):**

```text
✅ All test tags are defined in the tag contract
   Validated 47 feature files
   Contract defines 15 valid tags
```

**Output (failures):**

```text
❌ Found 2 undefined tag(s) used in 3 location(s):

  @performance (used in 1 file(s)):
    - specs/core/performance.feature:3

  @experimental (used in 2 file(s)):
    - specs/api/experimental.feature:1
    - specs/ui/experimental.feature:5

Fix: Add missing tags to .r2r/eac/testing-tags.yml
```

**Pattern tag validation:**

- `@skip:<reason>` - Validates against skip_reasons in contract
- `@deps:<name>` - Validates against system-dependencies.yml and OS platforms
- `@env:<moniker>` - Validates against environments.yml
- `@depm:<module>` - Validates against modules.yml

### validate module-hierarchy

Validate the module dependency graph structure for cycles and invalid references.

```bash
r2r eac validate module-hierarchy
```

**What it validates:**

- No circular dependencies
- All referenced modules exist
- Bidirectional consistency (depends_on relationships)
- All modules are reachable

**Output (success):**

```text
=== Module Hierarchy Validation Report ===

✅ All module hierarchy checks passed!
```

**Output (failures):**

```text
=== Module Hierarchy Validation Report ===

❌ References to Non-Existent Modules (1):
  • Module 'src-api' depends on 'src-common', but 'src-common' does not exist

❌ Circular Dependencies (1):
  • Circular dependency: src-auth -> eac-core -> src-utils -> src-auth
```

**Common errors:**

- Circular dependencies between modules
- References to non-existent modules
- Unreachable modules (orphaned)

**Fix:** Update module contracts to remove cycles or add missing modules.

### validate module-files

Validate that files belong to exactly one module and are properly owned.

```bash
r2r eac validate module-files
```

**What it validates:**

- Each file belongs to exactly one module
- No files in the "unordered" catch-all module
- No overlapping glob patterns

**Output (success):**

```text
=== Module File Ownership Validation Report ===

✅ All module file ownership checks passed!
```

**Output (failures):**

```text
=== Module File Ownership Validation Report ===

❌ Files in Unordered Module (3):
   These files should be claimed by a proper module:
  • src/shared/util.go
  • src/shared/helper.go
  • src/shared/constants.go

   Fix: Create or update module contracts to claim these files.

❌ Files with Multi-Module Ownership (1):
   Each file should belong to exactly one module:
  • go/eac/core/types.go
    Claimed by: eac-core, src-api

   Fix: Adjust module contract glob patterns to prevent overlap.
```

**Common errors:**

- Files not claimed by any module (unordered)
- Files claimed by multiple modules (overlapping patterns)

**Fix:** Update module contract glob patterns.

### validate markdown

Validate markdown file syntax and structure.

```bash
r2r eac validate markdown
```

**What it validates:**

- Valid markdown syntax
- Proper heading hierarchy (no skipped levels)
- Valid code blocks (JSON, YAML syntax)

**Output (success):**

```text
=== Markdown Validation Report ===

✅ All markdown files are valid
   Validated 42 files
```

**Output (failures):**

```text
=== Markdown Validation Report ===

❌ Validation errors in 2 file(s):

  docs/guide.md:
    - Line 15: Heading level skipped (h1 -> h3)
    - Line 42: Invalid JSON in code block

  README.md:
    - Line 5: Malformed link syntax
```

**Common errors:**

- Skipped heading levels (h1 -> h3)
- Invalid JSON/YAML in code blocks
- Malformed links or references

**Fix:** Correct markdown syntax errors.

### validate go-tidy

Validate that Go module dependencies are tidy (go.mod and go.sum are synchronized).

```bash
r2r eac validate go-tidy
```

**What it validates:**

- Runs `go mod tidy -diff` on all Go modules
- Checks go.mod and go.sum are synchronized
- Ensures no missing or extra dependencies

**Output (success):**

```text
=== Go Module Tidy Validation Report ===

Total Go modules: 8
Tidy modules: 8
Untidy modules: 0

✅ All Go modules have tidy dependencies!
```

**Output (failures):**

```text
=== Go Module Tidy Validation Report ===

Total Go modules: 8
Tidy modules: 7
Untidy modules: 1

❌ Modules with untidy dependencies:

  • go/eac/commands
    Diff:
    -  github.com/example/old v1.0.0
    +  github.com/example/new v2.0.0

To fix, run: go mod tidy
```

**Common errors:**

- go.mod and go.sum out of sync
- Missing dependencies in go.mod
- Unused dependencies in go.mod

**Fix:** Run `go mod tidy` in the affected module.

## Usage in CI Pipelines

### GitHub Actions

```yaml
name: Validate Repository

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
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

# Validate contracts
r2r eac validate contracts || exit 1

# Validate dependencies
r2r eac validate dependencies || exit 1

# Validate test tags
r2r eac validate test-tags || exit 1

# Validate Go tidiness
r2r eac validate go-tidy || exit 1

echo "✅ All validations passed"
```

### Make Integration

```makefile
.PHONY: validate validate-contracts validate-deps validate-tags validate-hierarchy validate-files validate-markdown validate-go-tidy

validate: validate-contracts validate-deps validate-tags validate-hierarchy validate-files validate-markdown validate-go-tidy

validate-contracts:
  r2r eac validate contracts

validate-deps:
  r2r eac validate dependencies

validate-tags:
  r2r eac validate test-tags

validate-hierarchy:
  r2r eac validate module-hierarchy

validate-files:
  r2r eac validate module-files

validate-markdown:
  r2r eac validate markdown

validate-go-tidy:
  r2r eac validate go-tidy
```

## Common Validation Errors and Fixes

### Contract Validation Errors

**Error:** Invalid enum value in modules.yml

```text
Error: modules.yml validation failed:
  Module 'src-api' has invalid type 'go-server' (expected: go-cli, go-library, go-commands, etc.)
```

**Fix:** Use a valid module type from the schema.

```yaml
# modules.yml
modules:
  - moniker: src-api
    type: go-library  # Use valid type
```

### Dependency Validation Errors

**Error:** Missing dependency in contract

```text
Error: Module 'src-api' depends on 'eac-core' but contract doesn't declare it
```

**Fix:** Add dependency to module contract.

```yaml
# modules.yml
modules:
  - moniker: src-api
    depends_on:
      - eac-core  # Add missing dependency
```

### Test Tag Errors

**Error:** Undefined tag in feature file

```text
❌ Found 1 undefined tag(s):
  @slow (used in 1 file(s)):
    - specs/performance.feature:3
```

**Fix:** Add tag to testing-tags.yml.

```yaml
# testing-tags.yml
tags:
  - tag: "@slow"
    description: "Tests that take longer than 5 seconds"
```

### Hierarchy Errors

**Error:** Circular dependency detected

```text
❌ Circular Dependencies (1):
  • Circular dependency: src-auth -> eac-core -> src-auth
```

**Fix:** Refactor to break the cycle, extract common code to a new module.

### File Ownership Errors

**Error:** Files in unordered module

```text
❌ Files in Unordered Module (2):
  • src/util/helper.go
  • src/util/types.go
```

**Fix:** Create or update module to claim these files.

```yaml
# modules.yml
modules:
  - moniker: src-utils
    type: go-library
    files:
      root: src/util
      patterns:
        - "**/*.go"
```

### Markdown Errors

**Error:** Skipped heading level

```text
docs/guide.md:
  - Line 15: Heading level skipped (h1 -> h3)
```

**Fix:** Use proper heading hierarchy.

```markdown
# Main Title (h1)
## Section (h2)
### Subsection (h3)  ✅ Correct hierarchy
```

### Go Tidy Errors

**Error:** go.mod out of sync

```text
❌ Modules with untidy dependencies:
  • go/eac/commands
    Diff:
    -  github.com/old/package v1.0.0
```

**Fix:** Run go mod tidy.

```bash
cd go/eac/commands
go mod tidy
```

## Best Practices

### When to Run Each Validation

**Always (Pre-commit):**

- `validate contracts` - Ensures config files are valid
- `validate go-tidy` - Catches dependency issues early

**Before Commit (Recommended):**

- `validate dependencies` - Ensures contracts match reality
- `validate test-tags` - Prevents tag filtering issues

**CI/CD Pipeline:**

- All validations - Comprehensive quality gate

**Periodic (Weekly):**

- `validate module-hierarchy` - Detect architectural drift
- `validate module-files` - Ensure clean organization
- `validate markdown` - Maintain documentation quality

### Validation Order

1. `contracts` - Foundation, must pass first
2. `dependencies` - Relies on contract validity
3. `module-hierarchy` - Checks contract consistency
4. `module-files` - Validates file organization
5. `test-tags` - Validates test infrastructure
6. `markdown` - Documentation quality
7. `go-tidy` - Final dependency check

### Integration Tips

**Run in sequence:** Each validation builds on the previous one.

```bash
r2r eac validate contracts && \
r2r eac validate dependencies && \
r2r eac validate module-hierarchy && \
r2r eac validate module-files && \
r2r eac validate test-tags && \
r2r eac validate markdown && \
r2r eac validate go-tidy
```

**Fail fast:** Stop on first error.

**Verbose output:** Capture full output in CI logs.

**Cache results:** Skip validations if no relevant changes.

## Typical Workflows

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

### Pull Request Checks

```bash
# Validate only changed modules
CHANGED_MODULES=$(r2r eac get-changed-modules)

# Always validate contracts
r2r eac validate contracts || exit 1

# Validate dependencies for changed modules
if [[ -n "$CHANGED_MODULES" ]]; then
  r2r eac validate dependencies || exit 1
  r2r eac validate go-tidy || exit 1
fi
```

### Post-Refactor

```bash
# After major restructuring
r2r eac validate module-hierarchy
r2r eac validate module-files
r2r eac validate dependencies
```

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
| Permission denied         | Check file permissions, may need sudo        |
| Schema not found          | Ensure .r2r/eac/schemas/ directory exists    |

## Advanced Usage

### Selective Validation

```bash
# Validate specific contract
r2r eac validate contracts  # All contracts

# Validate specific module type
r2r eac get-modules --type go-* | while read module; do
  # Check go.mod for each Go module
  r2r eac validate go-tidy
done
```

### Custom Validation Scripts

```bash
#!/bin/bash
# validate-all.sh

VALIDATIONS=(
  "contracts"
  "dependencies"
  "test-tags"
  "module-hierarchy"
  "module-files"
  "markdown"
  "go-tidy"
)

FAILED=0

for validation in "${VALIDATIONS[@]}"; do
  echo "Running: validate $validation"
  if ! r2r eac validate $validation; then
    FAILED=$((FAILED + 1))
    echo "❌ Failed: $validation"
  else
    echo "✅ Passed: $validation"
  fi
  echo ""
done

if [[ $FAILED -gt 0 ]]; then
  echo "❌ $FAILED validation(s) failed"
  exit 1
else
  echo "✅ All validations passed"
  exit 0
fi
```

### Validation Reports

```bash
# Generate validation report
{
  echo "# Validation Report"
  echo "Date: $(date)"
  echo ""

  for cmd in contracts dependencies test-tags module-hierarchy module-files markdown go-tidy; do
    echo "## validate $cmd"
    r2r eac validate $cmd 2>&1
    echo ""
  done
} > validation-report.md
```

## Summary

The `validate` command ensures repository integrity through comprehensive validation:

1. **contracts** - Schema validation for configuration files
2. **dependencies** - Consistency between go.mod and contracts
3. **test-tags** - All tags defined in tag contract
4. **module-hierarchy** - No cycles, valid references
5. **module-files** - Single ownership, proper organization
6. **markdown** - Valid syntax and structure
7. **go-tidy** - Go modules are synchronized

**Exit codes:** 0 on success, 1 on failure

**Best practice:** Run all validations in CI/CD pipelines as quality gates. Run contracts and go-tidy validations before every commit.

Use validation to catch configuration errors early, maintain contract consistency, and ensure repository quality standards.
