---
name: test
description: Write tests using MCP-powered discovery
model: claude-sonnet-4-5
thinking: enabled
color: green
---

# Testing Command

Write comprehensive tests using MCP tools to discover structure and verify coverage.

## What This Does

1. **Discover** code structure using MCP tools
2. **Plan** test scenarios
3. **Write** tests
4. **Run** tests using MCP commands
5. **Verify** all scenarios covered

## MCP Tools Used

This command leverages EAC MCP tools:

```bash
get-modules                  # Understand structure
get-files-by-module <mod>    # Locate code to test
get-dependencies <mod>       # Plan integration tests
get-tests                    # List existing tests
test <module>                # Run tests
show-test-results            # View results
```

## When to Use

- Writing tests for new features
- Improving test coverage
- After implementing functionality
- Before refactoring

## Workflow

### Step 1: MCP Discovery
```bash
# Understand what to test
get-modules
get-files-by-module <module-name>
get-dependencies <module-name>

# Check existing tests
get-tests
```

### Step 2: Plan Test Scenarios
Based on discovery:
- **Happy path**: Normal operations
- **Edge cases**: Boundary conditions
- **Error cases**: Invalid inputs, failures
- **Integration**: Dependencies (from `get-dependencies`)

### Step 3: Write Tests
Create tests following best practices:
- Clear test names
- Arrange-Act-Assert structure
- Independent tests
- Good assertions

### Step 4: Run and Verify
```bash
# Run tests
test <module>

# Check results
show-test-results
```

### Step 5: Iterate
Fix failures and improve coverage.

## Test Structure

**Good Test Naming**:
```
Test_Component_Scenario_ExpectedOutcome
```

**Examples**:
```
Test_Cache_GetExistingKey_ReturnsValue
Test_Cache_GetMissingKey_ReturnsError
Test_Cache_EvictFull_RemovesOldest
```

## Test Types

**Unit Tests**:
- Test individual functions
- Mock dependencies
- Fast execution

**Integration Tests**:
- Test module interactions
- Use real dependencies (from `get-dependencies`)
- Verify contracts

**Example MCP Discovery for Integration**:
```bash
get-dependencies api-module
# Output: Depends on auth-module, db-module
# Write integration tests for these interactions
```

## Example Usage

**User**: "Write tests for cache-module"

**MCP Discovery**:
```bash
get-modules                      # Confirm cache-module exists
get-files-by-module cache-module # See: cache.go, eviction.go
get-dependencies cache-module    # Depends on storage-module
get-tests                        # No existing tests
```

**Test Plan**:
1. Unit: Cache operations (set, get, delete)
2. Unit: Eviction policy
3. Integration: Storage interactions
4. Error: Invalid inputs

**Implementation**:
Write tests for all scenarios.

**Verification**:
```bash
test cache-module
show-test-results
```

## Testing Principles

**Always**:
- Use MCP discovery before writing tests
- Test behavior, not implementation
- Cover happy path and errors
- Make tests readable
- Run tests frequently

**Never**:
- Guess what to test (use MCP tools!)
- Skip error cases
- Make tests depend on order
- Ignore test failures
