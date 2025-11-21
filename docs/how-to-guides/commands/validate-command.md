# Validate Command

**Problem**: Module dependencies can become inconsistent, circular, or violate architectural constraints.

**Solution**: Use `validate` to check module dependencies against contracts and detect violations.

## Key Benefits

- Detects circular dependencies
- Validates dependency contracts
- Ensures architectural integrity
- Catches dependency issues early
- Provides clear violation reports

## Quick Start

```bash
# Validate all module dependencies
r2r eac validate dependencies

# Check after adding new dependencies
r2r eac validate dependencies
```

## Command Reference

### validate dependencies

Validate module dependency contracts.

```bash
r2r eac validate dependencies

# Example output:
# Validating module dependencies...
#
# ✅ src-commands
# ✅ src-core
# ✅ src-auth
# ✅ src-api
#
# Summary:
#   Total modules: 15
#   Valid: 15
#   Violations: 0
#
# ✅ All dependency contracts are valid
```

**What it checks:**
- Circular dependencies
- Dependency contract compliance
- Module isolation rules
- Layered architecture constraints
- Cross-module boundary violations

## Typical Workflows

### Before Committing

```bash
# Make changes to module dependencies
# ... edit go.mod or imports ...

# Validate before commit
r2r eac validate dependencies

# Commit if valid
r2r eac work commit --all
```

### During Code Review

```bash
# Reviewer checks dependency changes
r2r eac validate dependencies

# Check dependency graph
r2r eac show dependencies
```

### Continuous Integration

```bash
# CI pipeline validation
r2r eac build modules
r2r eac test modules
r2r eac validate dependencies

# Fail build if violations found
```

## Dependency Violations

### Circular Dependencies

```
❌ Circular dependency detected:
   src-auth → src-api → src-auth

Fix: Refactor to break the cycle:
- Extract shared code to src-core
- Use dependency inversion
- Introduce interface boundaries
```

### Contract Violations

```
❌ Contract violation in src-api:
   Depends on src-internal (not declared in contract)

Fix: Update module contract in src/api/module.yml:
dependencies:
  - src-core
  - src-internal
```

### Layer Violations

```
❌ Layer violation:
   src-api (infrastructure) → src-core (domain)
   Infrastructure cannot depend on domain

Fix: Reverse dependency using dependency inversion principle
```

## Integration Patterns

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Validating dependencies..."
r2r eac validate dependencies || exit 1
```

### GitHub Actions

```yaml
- name: Validate Dependencies
  run: |
    r2r eac validate dependencies
    if [ $? -ne 0 ]; then
      echo "Dependency validation failed"
      exit 1
    fi
```

### Makefile

```makefile
validate:
	r2r eac validate dependencies

ci: build test validate
```

## Best Practices

- **Validate frequently**: Run before every commit
- **CI enforcement**: Include in CI pipeline
- **Fix violations immediately**: Don't accumulate technical debt
- **Review dependency changes**: Carefully review PRs adding dependencies
- **Use contracts**: Define allowed dependencies in module contracts
- **Visualize dependencies**: Use `show dependencies` to understand graph

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Circular dependency | Extract shared code, use interfaces, break cycle |
| Undeclared dependency | Add to module contract `dependencies:` list |
| Layer violation | Refactor to respect architectural boundaries |
| Module not found | Check module contract exists and is valid |

## Summary

1. **Validate**: `r2r eac validate dependencies`
2. **Fix violations**: Refactor code or update contracts
3. **Revalidate**: Ensure fixes resolve issues
4. **Commit**: Commit only when validation passes

Dependency validation maintains clean architecture and prevents technical debt accumulation.
