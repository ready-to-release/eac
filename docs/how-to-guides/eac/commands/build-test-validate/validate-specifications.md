# Validate Specifications

{{ page_breadcrumb() }}

## What You'll Accomplish

Check Gherkin BDD specifications for quality and detect unused step definitions.

## Prerequisites

- Gherkin feature files in repository
- Step definitions implemented

## Steps

### 1. Validate Spec Quality

```bash
r2r eac validate specs
```

**What happens**: Checks Gherkin files against quality standards

### 2. Find Unused Steps

```bash
r2r eac get specs-unused-steps
```

**What happens**: Identifies step definitions that are never used

### 3. Review Issues

Fix reported issues:

- Missing scenario descriptions
- Undefined steps
- Duplicate scenarios
- Quality violations

## Validation Rules

Specifications are checked for:

- **Proper Gherkin syntax**
- **Complete scenarios** (Given/When/Then)
- **Clear descriptions**
- **Step definition coverage**
- **No undefined steps**

## Example Scenario

After adding new authentication specs:

```bash
# Validate specifications
r2r eac validate specs

# Output:
# ✓ features/auth/login.feature
# ✗ features/auth/register.feature
#   Line 12: Undefined step "Given user has valid email"
#   Line 15: Missing scenario description
#
# ✓ 8 files valid, ✗ 1 file with errors

# Check for unused steps
r2r eac get specs-unused-steps

# Output:
# {
#   "unused_steps": [
#     "steps/auth/old_login_steps.go:25: When user clicks legacy button"
#   ]
# }

# Fix issues
# 1. Implement missing step definition
# 2. Add scenario description
# 3. Remove unused step definition

# Validate again
r2r eac validate specs
# ✓ All specifications valid
```

## CI Integration

```bash
# In pre-commit hook
r2r eac validate specs || exit 1
```

## Common Issues

| Problem | Solution |
|---------|----------|
| Undefined steps | Implement step definition |
| Duplicate scenarios | Remove or rename scenarios |
| Quality violations | Follow Gherkin best practices |

## Next Steps

- [Create Specifications](../documentation/create-specifications.md) → Generate new specs

## Related Commands

- [`validate specs`](../../../../reference/commands/validate/specs.md) - Validate Gherkin
- [`get specs-unused-steps`](../../../../reference/commands/get/specs-unused-steps.md) - Find unused

{{ diataxis_footer() }}
