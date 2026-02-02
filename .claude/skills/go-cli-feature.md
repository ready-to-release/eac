---
name: go-cli-feature
description: End-to-end workflow for developing a new CLI feature
---

# Go CLI Feature Development Skill

This skill orchestrates the complete workflow for developing a new CLI feature following TDD and the Three Rules of Vibe Coding.

## When to Use This Skill

- Developing a new CLI command or feature
- Need end-to-end guidance from planning to completion
- Want to follow best practices automatically

## Workflow

### Step 1: Plan (go-architect)

**Action**: Delegate to go-architect agent

- Understand feature requirements
- Design command structure and interfaces
- Identify affected modules using MCP `get-modules`, `get-dependencies`
- Create architecture plan

**Output**: Architecture design with interfaces and implementation plan

### Step 2: Specify

**Action**: Create Gherkin specifications

- Use MCP `create-spec` to generate specification from description
- Create `.feature` file in `specs/` directory
- Validate with MCP `validate-specs`

**Output**: Validated specification file

### Step 3: Design CLI UX (go-cli-ux)

**Action**: Delegate to go-cli-ux agent

- Define command syntax, flags, output format
- Create help text and usage examples
- Design error messages
- Plan output formatting (tables, colors, etc.)

**Output**: Complete CLI command specification

### Step 4: Test First (go-test-engineer)

**Action**: Delegate to go-test-engineer agent

- Write failing unit tests (`*_test.go`)
- Implement Gherkin step definitions if doing BDD
- Verify tests fail for the right reasons
- Create test fixtures and helpers

**Output**: Complete test suite (failing)

### Step 5: Implement

**Action**: Write minimal code to pass tests

- Follow Three Rules of Vibe Coding:
  - **Easy to understand**: Clear names, simple flow, minimal abstraction
  - **Easy to change**: Small functions, stable boundaries, pure where possible
  - **Hard to break**: Input validation, error wrapping, comprehensive tests
- Write complete, compilable code (no stubs/placeholders)
- Use context.Context for I/O operations
- Wrap errors with %w for context

**Output**: Complete implementation

### Step 6: Verify (go-test-engineer)

**Action**: Run all tests

- Run unit tests: `go test ./...` or MCP `test <module>`
- Fix any failures immediately
- Verify with race detector: `go test -race ./...`
- Check coverage: `go test -cover ./...`

**Output**: All tests passing

### Step 7: Refine (code-simplifier)

**Action**: Run code-simplifier agent

- Use Task tool with subagent_type="code-simplifier:code-simplifier"
- Review suggested changes
- Apply simplifications that improve clarity
- Re-run tests to ensure nothing broke

**Output**: Simplified, cleaner code with tests still passing

**Note**: The code-simplifier is a Claude Code plugin that analyzes code and suggests simplifications

### Step 8: Document

**Action**: Update documentation

- Update command help text
- Add usage examples
- If CLI surface changed, add/update how-to guide in `docs/how-to-guides/`
- Update README if needed

**Output**: Complete documentation

## Verification Checklist

Before marking feature complete, verify:

- ✅ All tests pass (`go test ./...`)
- ✅ No race conditions (`go test -race ./...`)
- ✅ Specs validated (MCP `validate-specs`)
- ✅ Code simplified (code-simplifier ran)
- ✅ Documentation updated
- ✅ Lint passing (if golangci-lint available)
- ✅ Ready for commit/PR

## Code-Simplifier Integration

**When**: After implementation (Step 7)
**How**: Automatically via Task tool
**Why**: Ensures code stays simple and maintainable

## Quality Standards

Every feature must meet:

### Tests

- Table-driven unit tests
- Edge cases covered
- Error cases tested
- Mocks for external dependencies

### Code

- Functions < 40 lines (ideal), < 100 lines (max)
- Clear, intention-revealing names
- Errors wrapped with %w
- No unnecessary complexity

### CLI

- Consistent flag naming
- Clear help text with examples
- Proper exit codes
- User-friendly error messages

### Documentation

- Command help text
- Usage examples
- How-to guide if new command

## Example Usage

**User request**: "Add a new `validate config` command that checks configuration files"

**Workflow execution**:

1. **Plan**: Architecture for config validation (go-architect)
2. **Specify**: Create `validate-config.feature` specification
3. **Design UX**: Command structure, flags (--file, --strict), output format
4. **Test**: Write failing tests for validation logic
5. **Implement**: Config parser and validator
6. **Verify**: All tests pass
7. **Refine**: Run code-simplifier
8. **Document**: Update CLI help and add how-to guide

**Deliverables**:

- `validate_config.go` - Implementation
- `validate_config_test.go` - Tests
- `specs/validate-config.feature` - Specification
- `docs/how-to-guides/eac/commands/validate-config.md` - Documentation

## Tips for Success

- Don't skip steps (especially testing!)
- Keep commits small and focused
- Run tests after each change
- Use MCP tools for context discovery
- Delegate to specialized agents
- Always run code-simplifier at the end

This skill ensures you follow best practices and deliver high-quality CLI features every time.
