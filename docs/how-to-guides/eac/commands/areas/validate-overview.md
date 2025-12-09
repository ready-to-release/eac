# Validation System

{{ page_breadcrumb() }}

The validation system in EAC ensures repository integrity across contracts, dependencies, configurations, and code quality.

## What is the Validation System?

EAC's validation system enables you to:

- **Validate contracts** against JSON schemas
- **Check dependency consistency** between go.mod and contracts
- **Verify test tags** are properly defined
- **Analyze module hierarchy** for cycles and issues
- **Ensure file ownership** is properly assigned
- **Validate markdown** syntax and structure
- **Check Go module tidiness** for dependency sync

The system provides automated quality gates for CI/CD pipelines and local development.

## When to Use Validate Commands

Use validate commands when you need:

| Scenario           | Command                     |
| ------------------ | --------------------------- |
| Schema validation  | `validate contracts`        |
| Dependency check   | `validate dependencies`     |
| Tag verification   | `validate test-tags`        |
| Hierarchy analysis | `validate module-hierarchy` |
| Ownership check    | `validate module-files`     |
| Markdown lint      | `validate markdown`         |
| Go mod sync        | `validate go-tidy`          |

### Common Use Cases

- **Pre-commit hooks** - Catch errors before commit
- **CI/CD pipelines** - Automated quality gates
- **Post-refactor** - Verify structural integrity
- **Pull requests** - Ensure changes meet standards

## Key Concepts

### Validation Types

| Type        | Purpose              | Validates           |
| ----------- | -------------------- | ------------------- |
| Schema      | Contract validity    | YAML/JSON structure |
| Consistency | Implementation match | go.mod vs contracts |
| Tags        | Test infrastructure  | Feature file tags   |
| Structure   | Architecture         | Dependency graph    |
| Ownership   | Organization         | File assignment     |
| Syntax      | Documentation        | Markdown format     |
| Tidiness    | Dependencies         | go.mod/go.sum sync  |

### Exit Codes

All validate subcommands use consistent exit codes:

| Code | Description       |
| ---- | ----------------- |
| 0    | Validation passed |
| 1    | Validation failed |

### Contract Files

Validation checks these contract files:

```text
.r2r/eac/
├── modules.yml         # Module contracts
├── environments.yml    # Environment definitions
├── testing-tags.yml    # Test tag definitions
└── test-suites.yml     # Test suite configurations
```

## Workflow Overview

### Recommended Validation Order

Run validations in sequence (each builds on previous):

```bash
# 1. Foundation - must pass first
r2r eac validate contracts

# 2. Relies on contract validity
r2r eac validate dependencies

# 3. Checks contract consistency
r2r eac validate module-hierarchy

# 4. Validates file organization
r2r eac validate module-files

# 5. Validates test infrastructure
r2r eac validate test-tags

# 6. Documentation quality
r2r eac validate markdown

# 7. Final dependency check
r2r eac validate go-tidy
```

### When to Run Each

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

## Integration Points

### With Build

```bash
# Validate before building
r2r eac validate contracts
r2r eac build eac-core
```

### With Test

```bash
# Validate test infrastructure before testing
r2r eac validate test-tags
r2r eac test suite commit
```

### Full CI Pipeline

```bash
# Complete validation pipeline
r2r eac validate contracts && \
r2r eac validate dependencies && \
r2r eac validate module-hierarchy && \
r2r eac validate module-files && \
r2r eac validate test-tags && \
r2r eac validate markdown && \
r2r eac validate go-tidy
```

## Best Practices

### Do's

- **Run contracts first** - Foundation for other validations
- **Use in CI/CD** - Automated quality gates
- **Fail fast** - Stop pipeline on first error
- **Run go-tidy locally** - Prevents CI failures

### Don'ts

- **Don't skip validations** - All have purpose
- **Don't ignore failures** - Fix root causes
- **Don't run out of order** - Dependencies matter

## Troubleshooting

| Problem             | Solution                     |
| ------------------- | ---------------------------- |
| Contract fails      | Check YAML syntax and schema |
| Dependency mismatch | Update contract or go.mod    |
| Undefined tags      | Add to testing-tags.yml      |
| Circular dependency | Refactor to break cycle      |
| Files unordered     | Update module patterns       |
| Markdown errors     | Fix syntax and headings      |
| Go not tidy         | Run `go mod tidy`            |

## Next Steps

- [Validate Configuration](validate-configuration.md) - Configure validation settings
- [Validate Commands](validate-commands.md) - Full command reference

## Related Areas

- [Build](build-overview.md) - Build after validating
- [Test](test-overview.md) - Test after validating
- [Specifications](specifications-overview.md) - Validate test tags

{{ diataxis_footer() }}
