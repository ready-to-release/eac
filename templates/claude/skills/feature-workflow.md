---
name: feature-workflow
description: End-to-end feature development with MCP tools
model: claude-sonnet-4-5
thinking: extended
color: purple
---

# Feature Development Workflow

Complete workflow for implementing a feature using MCP-powered discovery.

## Workflow Steps

### 1. Planning Phase (Use MCP!)

**Input**: Feature description

**MCP Discovery**:
```bash
get-modules                  # Understand structure
get-dependencies <module>    # Check dependency impacts
show-modules                 # View module details
validate-contracts           # Verify current state
```

**Actions**:
1. Use MCP tools to discover current state
2. Design solution based on discovery
3. Create implementation plan
4. Save to `out/<feature-name>-plan.md`

**Output**: Detailed plan document with MCP commands

---

### 2. Specification Phase

**Input**: Plan document

**Actions**:
1. Define acceptance criteria
2. Update module contracts if needed
3. Document expected behavior

**Output**: Specification document

---

### 3. Implementation Phase

**Input**: Plan and specification

**Actions**:
1. Follow implementation steps from plan
2. Write code incrementally
3. Use MCP commands to verify progress:
   ```bash
   build <module>              # Build incrementally
   validate-module-hierarchy   # Check no circular deps
   ```
4. Commit logical chunks

**Output**: Working implementation

---

### 4. Testing Phase

**Input**: Implementation

**MCP Testing**:
```bash
test <module>              # Run module tests
get-tests                  # List all tests
show-test-results          # View test results
```

**Actions**:
1. Write tests
2. Run tests using MCP commands
3. Fix any failures
4. Verify coverage

**Output**: Passing tests

---

### 5. Validation Phase

**Input**: Complete implementation

**MCP Validation**:
```bash
validate-contracts         # Verify contracts
validate-module-hierarchy  # Check dependency graph
validate-artifacts         # Verify build artifacts
test <module>              # Final test run
build <module>             # Final build
```

**Actions**:
1. Run all MCP validation commands
2. Check against plan
3. Self-review code changes
4. Create pull request

**Output**: Ready for merge

---

## Success Criteria

- ✅ Plan document in `out/` folder
- ✅ MCP discovery used for planning
- ✅ Implementation matches plan
- ✅ All tests passing (`test <module>`)
- ✅ Contracts validated (`validate-contracts`)
- ✅ Module hierarchy valid (`validate-module-hierarchy`)
- ✅ Code reviewed
- ✅ Documentation updated

## Example Flow

**User**: "Implement caching layer"

**Workflow**:

1. **Planning with MCP**:
   ```bash
   get-modules                 # Find data access modules
   get-dependencies data-access  # See what depends on it
   ```
   Result: `out/caching-layer-plan.md`

2. **Implementation**:
   - Add cache module
   - Update data access to use cache
   - Verify with `build cache-module`

3. **Testing**:
   ```bash
   test cache-module
   test data-access
   show-test-results
   ```

4. **Validation**:
   ```bash
   validate-contracts
   validate-module-hierarchy
   build cache-module
   ```

5. **Review**: Create PR with passing tests

## MCP Commands Reference

**Discovery**:
- `get-modules`
- `get-dependencies <module>`
- `show-modules`
- `get-files-by-module <module>`

**Build & Test**:
- `build <module>`
- `test <module>`
- `get-tests`
- `show-test-results`

**Validation**:
- `validate-contracts`
- `validate-module-hierarchy`
- `validate-artifacts`
- `validate-dependencies`

## Phase Checklist

**Planning Phase**:
- ✅ Ran `get-modules` to understand structure
- ✅ Ran `get-dependencies` to check impacts
- ✅ Created plan in `out/` folder
- ✅ Plan includes MCP verification commands

**Implementation Phase**:
- ✅ Following plan steps
- ✅ Running `build <module>` after changes
- ✅ Committing working increments
- ✅ No circular dependencies (`validate-module-hierarchy`)

**Testing Phase**:
- ✅ Tests written for new code
- ✅ Tests cover edge cases and errors
- ✅ `test <module>` passes
- ✅ `show-test-results` reviewed

**Validation Phase**:
- ✅ All validations pass
- ✅ Clean build
- ✅ All tests pass
- ✅ Ready for review
