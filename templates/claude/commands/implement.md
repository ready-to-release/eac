---
name: implement
description: Implement features using plan and MCP verification
model: claude-sonnet-4-5
thinking: enabled
color: blue
---

# Implementation Command

Implement features following a plan and using MCP tools for verification.

## What This Does

1. **Read plan** from `out/<feature-name>-plan.md`
2. **Implement** changes incrementally
3. **Verify** each step using MCP tools
4. **Test** implementation continuously

## MCP Tools Used

This command uses EAC MCP tools for verification:

```bash
build <module>                # Build incrementally
test <module>                 # Run tests
validate-module-hierarchy     # Check dependencies
validate-contracts            # Verify interfaces
get-files-by-module <module>  # Locate files to modify
```

## When to Use

- After creating an implementation plan
- When implementing planned features
- For incremental development with verification

## Workflow

### Step 1: Review Plan
Read plan document from `out/` folder.

### Step 2: Locate Files
```bash
# Find files to modify
get-files-by-module <module-name>
```

### Step 3: Implement Incrementally
Make changes in small steps:

1. Modify code
2. Build: `build <module>`
3. Test: `test <module>`
4. Fix any failures
5. Repeat

### Step 4: Verify Completion
```bash
# Final verification
validate-module-hierarchy    # Check dependencies
validate-contracts           # Verify interfaces
test <module>                # All tests pass
build <module>               # Clean build
```

## Implementation Principles

**Always**:
- Follow the plan
- Make small, verifiable changes
- Run `build <module>` after each change
- Run `test <module>` frequently
- Commit working increments

**Never**:
- Make changes not in the plan (discuss with user first)
- Skip verification steps
- Continue with failing tests
- Make large, unverified changes

## Example Usage

**User**: "Implement the caching plan in out/caching-layer-plan.md"

**Workflow**:

1. **Read Plan**: Load `out/caching-layer-plan.md`

2. **Locate Files**:
```bash
get-files-by-module cache-module
```

3. **Implement Step 1** (from plan):
   - Add cache interface
   - Verify: `build cache-module`

4. **Implement Step 2**:
   - Add cache implementation
   - Verify: `build cache-module`, `test cache-module`

5. **Implement Step 3**:
   - Integrate with API
   - Verify: `build api-module`, `test api-module`

6. **Final Verification**:
```bash
validate-module-hierarchy
validate-contracts
test cache-module
test api-module
```

## Verification Checklist

Before reporting complete:

- ✅ All plan steps implemented
- ✅ `build <module>` succeeds
- ✅ `test <module>` passes
- ✅ `validate-module-hierarchy` passes
- ✅ `validate-contracts` passes
- ✅ Code reviewed
