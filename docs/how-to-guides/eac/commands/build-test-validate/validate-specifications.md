# Validate Specifications

## What You'll Accomplish

Check Gherkin BDD specifications for quality and detect unused step definitions.

## Prerequisites

### Required Knowledge

**New to specifications?** Learn these concepts first:

- [BDD Fundamentals](../../../../explanation/specifications/concepts/bdd-fundamentals.md) - Understand specification structure and quality standards

### Required Setup

- Gherkin feature files in repository
- Step definitions implemented

## Steps

### 1. Validate Spec Quality

```bash
eac validate specs
```

**What happens**: Checks Gherkin files against quality standards

### 2. Find Unused Steps

```bash
eac get specs-unused-steps
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
- **Required tags** - Feature level (`@deps:`, `@depm:`, `@env:`) and scenario level (`@L0`-`@L4`, `@ov`, etc.)
- **Feature naming** - Must follow `<module>_<feature-name>` convention
- **Clear descriptions**
- **Step definition coverage**
- **No undefined steps**

**Tagging Requirements**: All tags must comply with `.eac/testing-tags.yml`

## Example Scenario

After adding new authentication specs:

```bash
# Validate specifications
eac validate specs

# Output:
# ✓ specs/r2r-cli/user-login/specification.feature
# ✗ specs/r2r-cli/user-registration/specification.feature
#   Line 12: Undefined step "Given user has valid email"
#   Line 15: Missing scenario description
#   Line 8: Missing required tag @L0-@L4 on scenario
#
# ✓ 8 files valid, ✗ 1 file with errors

# Check for unused steps
eac get specs-unused-steps

# Output:
# {
#   "unused_steps": [
#     "go/specs/impl/r2r-cli/steps_user_login.go:25: When user clicks legacy button"
#   ]
# }

# Fix issues
# 1. Implement missing step definition
# 2. Add scenario description
# 3. Remove unused step definition

# Validate again
eac validate specs
# ✓ All specifications valid
```

## CI Integration

```bash
# In pre-commit hook
eac validate specs || exit 1
```

## Common Issues

| Problem             | Solution                      |
| ------------------- | ----------------------------- |
| Undefined steps     | Implement step definition     |
| Duplicate scenarios | Remove or rename scenarios    |
| Quality violations  | Follow Gherkin best practices |

## Next Steps

- [Create Specifications](../documentation/create-specifications.md) → Generate new specs

## Related Commands

- [`validate specs`](../../../../reference/eac/commands/validate/specs.md) - Validate Gherkin
- [`get specs-unused-steps`](../../../../reference/eac/commands/get/specs-unused-steps.md) - Find unused
