# Validate Before Commit

{{ page_breadcrumb() }}

## What You'll Accomplish

Run comprehensive quality checks before committing to ensure code meets repository standards.

## Prerequisites

### Required Knowledge
**New to validation?** Learn these concepts first:
- [Building and Testing Changes](../../../../tutorials/core-workflows/building-and-testing.md) - Understand pre-commit validation workflow

### Required Setup
- Working in EAC-managed repository
- Changes ready to commit

## Steps

### 1. Run All Validations

```bash
r2r eac validate
```

**What happens**: Runs all validation checks:

- Contract validation
- Dependency validation
- Specification validation
- Go module tidiness
- Module file ownership

### 2. Fix Any Issues

If validation fails, fix the reported issues:

```bash
# If go.mod is not tidy
go mod tidy

# If specs have issues
r2r eac validate specs

# If dependencies are wrong
r2r eac validate dependencies
```

### 3. Validate Again

```bash
r2r eac validate
```

**What happens**: Re-runs all checks to verify fixes

## Quick Validation Workflow

```bash
# Validate specific areas
r2r eac validate specs          # Check Gherkin specs
r2r eac validate go-tidy        # Check Go modules
r2r eac validate dependencies   # Check module deps
r2r eac validate contracts      # Check contracts
```

## Example Scenario

You're ready to commit but want to ensure quality first:

```bash
# Run validation
r2r eac validate

# Output:
# ✓ Contracts valid
# ✓ Dependencies valid
# ✗ Go modules not tidy (run 'go mod tidy')
# ✓ Specifications valid

# Fix issue
go mod tidy

# Validate again
r2r eac validate
# ✓ All validations passed

# Now safe to commit
git commit -m "feat: add authentication"
```

## Pre-Commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
r2r eac validate || exit 1
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "Go modules not tidy" | Run `go mod tidy` |
| "Contract invalid" | Check module.yml format |
| "Dependency mismatch" | Update dependencies in module.yml |

## Next Steps

- [Scan for Security Issues](./scan-for-security-issues.md) → Security validation
- [Make Commits with AI](../development-workflow/make-commits-with-ai.md) → Commit changes

## Related Commands

- [`validate`](../../../../reference/commands/validate/validate.md) - Run all validations
- [`validate specs`](../../../../reference/commands/validate/specs.md) - Validate specifications
- [`validate dependencies`](../../../../reference/commands/validate/dependencies.md) - Check dependencies

{{ diataxis_footer() }}
