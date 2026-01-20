---
name: review
description: Review changes using MCP validation
model: claude-sonnet-4-5
thinking: enabled
color: yellow
---

# Review Command

Review code changes using MCP tools to validate quality and correctness.

## What This Does

1. **Analyze** changes made
2. **Validate** using MCP tools
3. **Check** quality and correctness
4. **Provide** feedback

## MCP Tools Used

This command uses EAC MCP tools for validation:

```bash
get-modules                  # Understand affected modules
get-dependencies <mod>       # Check dependency impacts
validate-module-hierarchy    # Check for circular deps
validate-contracts           # Verify interfaces
validate-artifacts           # Check build outputs
test <module>                # Verify tests pass
build <module>               # Verify builds
show-test-results            # View test outcomes
```

## When to Use

- Before committing changes
- Before creating pull requests
- After completing features
- During code review process

## Workflow

### Step 1: Understand Changes
Review what was modified.

### Step 2: MCP Discovery
```bash
# Identify affected modules
get-modules

# Check dependencies
get-dependencies <modified-module>

# Find all modified files
get-files-by-module <module>
```

### Step 3: Run Validations
```bash
# Verify structure
validate-module-hierarchy

# Verify interfaces
validate-contracts

# Verify builds
build <module>

# Verify tests
test <module>

# Check artifacts
validate-artifacts
```

### Step 4: Review Results
```bash
# View test outcomes
show-test-results
```

### Step 5: Provide Feedback
Summarize findings:
- ✅ What passes
- ⚠️ What needs attention
- ❌ What must be fixed

## Review Checklist

**Code Quality**:
- ✅ Clear, readable code
- ✅ Appropriate naming
- ✅ Proper error handling
- ✅ No code duplication
- ✅ Follows project patterns

**MCP Validations**:
- ✅ `validate-module-hierarchy` passes
- ✅ `validate-contracts` passes
- ✅ `validate-artifacts` passes
- ✅ `build <module>` succeeds
- ✅ `test <module>` passes

**Testing**:
- ✅ Tests exist for new code
- ✅ Tests cover edge cases
- ✅ Tests pass consistently
- ✅ Good test names

**Documentation**:
- ✅ Clear commit messages
- ✅ Updated documentation
- ✅ Comments where needed

## Example Usage

**User**: "Review my changes to cache-module"

**MCP Validation**:
```bash
# Check structure
validate-module-hierarchy

# Check interfaces
validate-contracts

# Build
build cache-module

# Test
test cache-module

# View results
show-test-results
```

**Review Report**:
```markdown
## Review: cache-module changes

### MCP Validations
✅ validate-module-hierarchy: PASS
✅ validate-contracts: PASS
✅ build cache-module: PASS
✅ test cache-module: PASS (15/15 tests)

### Code Quality
✅ Clear implementation
✅ Good error handling
✅ Well-tested

### Recommendations
- None - ready to commit
```

## Review Principles

**Always**:
- Run all MCP validations
- Check test coverage
- Verify builds succeed
- Review for clarity
- Consider maintainability

**Never**:
- Approve failing tests
- Skip validations
- Ignore warnings
- Rush reviews
