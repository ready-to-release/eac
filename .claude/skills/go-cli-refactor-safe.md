---
name: go-cli-refactor-safe
description: Safe refactoring workflow with continuous validation
---

# Go CLI Safe Refactoring Skill

This skill orchestrates safe refactoring with continuous testing to prevent breaking changes.

## When to Use This Skill

- Refactoring existing code
- Improving code clarity or structure
- Reducing complexity
- Need to ensure no behavior changes

## Core Principle

**Refactoring changes structure, NOT behavior**:

Every step must maintain passing tests. If tests fail, either:

1. The refactoring broke something (fix it), OR
2. The tests were testing implementation details (update tests)

## Workflow

### Step 1: Baseline (go-test-engineer)

**Action**: Establish current state

- Run all tests to establish baseline: `go test ./...`
- Record coverage: `go test -cover ./...`
- Document current test pass rate
- Note any existing issues

**Output**: Baseline metrics

```text
Tests: 245 passed, 0 failed
Coverage: 84%
Race detector: Clean
```

**Blocker**: If tests failing at baseline, fix them BEFORE refactoring

### Step 2: Plan Refactoring (go-architect)

**Action**: Design target structure

- Identify code to refactor
- Design improved structure
- Plan incremental steps (small changes)
- Identify risks

**Output**: Refactoring plan with steps

**Example**:

```markdown
## Refactoring Plan

**Target**: Simplify config validation logic

**Current Issues**:
- 200-line function (too large)
- Nested if statements (hard to follow)
- Mixed concerns (validation + formatting)

**Target Design**:
- Extract validation into separate functions
- Separate formatting logic
- Use early returns to reduce nesting

**Steps**:
1. Extract email validation
2. Extract phone validation
3. Extract formatting
4. Simplify main function
```

### Step 3: Refactor Incrementally

**Action**: Make ONE change at a time

**Process for EACH step**:

1. Make ONE small change
2. Run tests: `go test ./...`
3. If tests pass: Commit
4. If tests fail: Fix immediately or revert
5. Repeat

**Important**:

- NEVER make multiple changes before testing
- Keep diffs small and reviewable
- Commit frequently
- Each commit should have passing tests

**Example step-by-step**:

```go
// Step 1: Extract email validation
func validateEmail(email string) error {
    // validation logic
}

// Run tests: go test ./...
// ✅ All pass - commit

// Step 2: Extract phone validation
func validatePhone(phone string) error {
    // validation logic
}

// Run tests: go test ./...
// ✅ All pass - commit

// Continue...
```

### Step 4: Verify After Each Step (go-test-engineer)

**Action**: Run comprehensive tests after EACH change

- `go test ./...` - All tests must pass
- `go test -race ./...` - No new race conditions
- Verify behavior unchanged
- If any failure: Fix or revert immediately

**Never move to next step until tests pass**:

### Step 5: Simplify (code-simplifier)

**Action**: Run code-simplifier plugin

- Use Task tool with subagent_type="code-simplifier"
- Review suggested changes
- Apply simplifications
- Run tests after applying changes
- Commit simplifications separately

**Output**: Cleaner, simpler code

### Step 6: Final Validation

**Action**: Comprehensive validation

- Run full test suite: `go test ./...`
- Race detector: `go test -race ./...`
- Coverage check: Verify not decreased
- Review all diffs for clarity improvements
- Lint check: `golangci-lint run` (if available)

**Comparison to baseline**:

```text
Before:
- Tests: 245 passed
- Coverage: 84%
- File: 200 lines, cyclomatic complexity 15

After:
- Tests: 245 passed (same ✅)
- Coverage: 86% (improved ✅)
- File: 150 lines, cyclomatic complexity 8 (improved ✅)
```

## Verification Checklist

Refactoring complete when ALL are ✅:

- ✅ All tests still pass (same pass rate as baseline)
- ✅ No behavior changes (tests prove it)
- ✅ Code is simpler and clearer
- ✅ Coverage maintained or improved
- ✅ No new race conditions
- ✅ All commits have passing tests
- ✅ Diffs are small and reviewable

## Common Refactoring Patterns

### Extract Function

```go
// Before: Long function
func Process(data Data) error {
    // 50 lines of validation
    // 30 lines of transformation
    // 20 lines of persistence
}

// After: Extracted functions
func Process(data Data) error {
    if err := validate(data); err != nil {
        return err
    }
    transformed := transform(data)
    return persist(transformed)
}

func validate(data Data) error { /* ... */ }
func transform(data Data) Data { /* ... */ }
func persist(data Data) error { /* ... */ }
```

### Reduce Nesting

```go
// Before: Nested ifs
func Process(data Data) error {
    if data.IsValid() {
        if data.HasPermission() {
            if data.IsReady() {
                // process
            } else {
                return ErrNotReady
            }
        } else {
            return ErrNoPermission
        }
    } else {
        return ErrInvalid
    }
}

// After: Early returns
func Process(data Data) error {
    if !data.IsValid() {
        return ErrInvalid
    }
    if !data.HasPermission() {
        return ErrNoPermission
    }
    if !data.IsReady() {
        return ErrNotReady
    }
    // process
}
```

### Extract Interface

```go
// Before: Tight coupling
type Service struct {
    db *sql.DB
}

// After: Interface for flexibility
type Repository interface {
    Get(id string) (*Data, error)
    Save(data *Data) error
}

type Service struct {
    repo Repository
}
```

## Safety Rules

**Always**:

- Run tests after EVERY change
- Keep changes small (< 50 lines per commit)
- Commit frequently with passing tests
- Use version control (easy to revert)
- Focus on one concern at a time

**Never**:

- Make multiple changes before testing
- Skip tests "just this once"
- Refactor and add features simultaneously
- Make changes without understanding code
- Ignore failing tests

## When Things Go Wrong

**If tests fail after refactoring**:

1. **Read failure carefully**: What broke?
2. **Identify cause**: Logic error or test issue?
3. **Fix or revert**: Fix quickly or revert the change
4. **Rethink approach**: Maybe smaller steps needed

**If coverage decreases**:

- Identify what's no longer covered
- Add tests for uncovered paths
- Don't proceed until coverage restored

## Example Usage

**User request**: "Refactor the config parser - it's 300 lines and hard to understand"

**Skill execution**:

1. **Baseline**: Run tests, record metrics (245 tests, 84% coverage)
2. **Plan**: Break parser into stages (validate → parse → transform)
3. **Refactor**:
   - Step 1: Extract validation (commit, tests pass)
   - Step 2: Extract parsing (commit, tests pass)
   - Step 3: Extract transformation (commit, tests pass)
   - Step 4: Simplify main function (commit, tests pass)
4. **Verify**: All 245 tests pass, coverage now 87%
5. **Simplify**: Run code-simplifier, commit changes
6. **Final**: File now 180 lines, clearer structure, all tests passing

**Outcome**: ✅ Code refactored safely, all tests passing, improved clarity

## Success Criteria

Refactoring successful when:

- ✅ All tests passing (same as before)
- ✅ Code is clearer/simpler
- ✅ Behavior unchanged
- ✅ Coverage maintained/improved
- ✅ Small, reviewable commits

This skill ensures refactoring is safe, incremental, and continuously validated.
