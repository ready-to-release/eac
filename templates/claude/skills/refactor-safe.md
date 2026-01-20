---
name: refactor-safe
description: Safe refactoring workflow with MCP validation
model: claude-sonnet-4-5
thinking: extended
color: orange
---

# Safe Refactoring Workflow

Refactor code safely using MCP tools to validate correctness at each step.

## Principles

**Safe refactoring means**:
- Small, incremental changes
- Tests pass after each change
- Behavior preserved
- Dependencies verified

**MCP tools validate safety at each step**

## Workflow Steps

### 1. Establish Baseline

**Before any changes**:

```bash
# Capture current state
get-modules
get-dependencies <module>
get-files-by-module <module>

# Verify current state passes
validate-module-hierarchy
validate-contracts
build <module>
test <module>
show-test-results
```

**Requirements**:
- ✅ All tests must pass
- ✅ Clean build
- ✅ No validation errors

**If baseline fails**: Fix issues before refactoring.

---

### 2. Plan Refactoring

**Input**: Code to refactor

**MCP Discovery**:
```bash
get-files-by-module <module>       # Locate files
get-dependencies <module>          # Check impacts
```

**Plan**:
1. Break refactoring into small steps
2. Define verification at each step
3. Identify affected modules (from MCP)

**Output**: Step-by-step plan

---

### 3. Incremental Refactoring

**For each refactoring step**:

1. **Make small change**
   - Rename variable
   - Extract function
   - Move code
   - Simplify logic

2. **Verify immediately**:
   ```bash
   build <module>              # Must pass
   test <module>               # Must pass
   validate-module-hierarchy   # Check dependencies
   ```

3. **If verification fails**:
   - Revert change
   - Identify issue
   - Try smaller step

4. **If verification passes**:
   - Commit change
   - Move to next step

**Never**: Make multiple changes before verifying.

---

### 4. Validate Complete Refactoring

**After all steps**:

```bash
# Verify final state
validate-contracts         # Interfaces unchanged
validate-module-hierarchy  # Dependencies valid
validate-artifacts         # Build outputs correct
build <module>             # Clean build
test <module>              # All tests pass
show-test-results          # No regressions
```

**Compare to baseline**:
- ✅ Same test count
- ✅ Same test outcomes
- ✅ Same dependencies
- ✅ Same contracts

---

## Example Refactoring

**Goal**: Extract caching logic into separate module

### Step 1: Baseline
```bash
get-modules                     # Current: api-module
get-dependencies api-module     # Depends on: db-module
test api-module                 # 25 tests pass
```

### Step 2: Plan
1. Extract cache interface
2. Extract cache implementation
3. Create cache-module
4. Update api-module to use cache-module

### Step 3: Execute (Incremental)

**Change 1**: Extract cache interface in api-module
```bash
build api-module
test api-module                 # 25 tests still pass ✅
git commit -m "Extract cache interface"
```

**Change 2**: Extract cache implementation in api-module
```bash
build api-module
test api-module                 # 25 tests still pass ✅
git commit -m "Extract cache implementation"
```

**Change 3**: Move cache code to new cache-module
```bash
build cache-module              # New module builds ✅
build api-module                # Still builds ✅
test cache-module               # Tests pass ✅
test api-module                 # 25 tests still pass ✅
validate-module-hierarchy       # No circular deps ✅
git commit -m "Create cache-module"
```

### Step 4: Final Validation
```bash
validate-contracts              # PASS ✅
validate-module-hierarchy       # PASS ✅
build api-module                # PASS ✅
build cache-module              # PASS ✅
test api-module                 # 25 tests PASS ✅
test cache-module               # Tests PASS ✅
```

**Outcome**: Safe refactoring complete, behavior preserved.

---

## Safety Checklist

**Before Starting**:
- ✅ All tests pass
- ✅ Clean build
- ✅ Validations pass
- ✅ Plan created

**During Each Step**:
- ✅ Change is small
- ✅ `build <module>` passes
- ✅ `test <module>` passes
- ✅ Committed before next step

**After Completion**:
- ✅ All validations pass
- ✅ Test count unchanged
- ✅ No regressions
- ✅ Dependencies valid

## Common Refactorings

### Extract Function
1. Identify code to extract
2. Create new function
3. Build & test ✅
4. Replace usage
5. Build & test ✅

### Rename Variable/Function
1. Find all usages (`get-files-by-module`)
2. Rename in one place
3. Build & test ✅
4. Rename in next place
5. Build & test ✅
6. Repeat

### Move Code Between Modules
1. Copy code to new location
2. Build both modules ✅
3. Update references
4. Build & test ✅
5. Remove old code
6. Build & test ✅
7. Validate dependencies ✅

### Simplify Logic
1. Ensure tests cover behavior
2. Simplify small section
3. Build & test ✅
4. Simplify next section
5. Build & test ✅
6. Repeat

## Recovery Steps

**If build fails**:
1. Revert last change
2. Verify build passes
3. Try smaller change

**If tests fail**:
1. Review test output (`show-test-results`)
2. Identify breaking change
3. Revert change
4. Fix issue or use different approach

**If validation fails**:
```bash
validate-module-hierarchy       # Check for circular deps
validate-contracts              # Check interface changes
```
1. Fix validation issue
2. Re-verify with MCP tools

## MCP Commands for Refactoring

**Discovery**:
- `get-modules` - Find modules to refactor
- `get-files-by-module <module>` - Locate files
- `get-dependencies <module>` - Check impacts

**Continuous Verification**:
- `build <module>` - After EVERY change
- `test <module>` - After EVERY change
- `validate-module-hierarchy` - Check dependencies

**Final Validation**:
- `validate-contracts` - Verify interfaces
- `validate-artifacts` - Check build outputs
- `show-test-results` - Confirm no regressions
