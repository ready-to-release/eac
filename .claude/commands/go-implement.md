# Implement

```text
description: "Implement a Go CLI feature using TDD"
```

You are implementing a Go CLI feature using Test-Driven Development.

## Process

1. **Review the plan** (if exists):
   - Read any existing plan or design documents
   - Clarify requirements if needed

2. **Create specifications** (if not exists):
   - Use MCP `create-spec` (or CLI `go run ./go/eac/commands create spec`) to generate Gherkin spec
   - Validate with MCP `validate-specs` (or CLI `go run ./go/eac/commands validate specs`)

3. **Write tests FIRST** (delegate to go-test-engineer):
   - Use Task tool with subagent_type="go-test-engineer"
   - Write failing unit tests (`*_test.go`)
   - Write Gherkin step implementations if doing BDD
   - Verify tests fail for the right reasons

4. **Implement to pass tests**:
   - Write minimal code to make tests pass
   - Follow "Three Rules of Vibe Coding":
     - **Easy to understand**: Clear, simple, explicit
     - **Easy to change**: Small functions, stable boundaries
     - **Hard to break**: Input validation, error handling, tests

5. **Run tests continuously**:
   - After EACH code change: `go test ./...`
   - Fix failures immediately
   - Use MCP `test` for module-specific tests

6. **Refactor for clarity**:
   - Improve names, extract functions, reduce complexity
   - Run tests after each refactoring
   - Keep diffs small and reviewable

## Quality Bars

Before reporting complete:

- ✅ All tests pass
- ✅ Code follows Go conventions (gofmt, go vet)
- ✅ No race conditions (verify with `go test -race`)
- ✅ Errors properly wrapped with context

## Example Usage

User: `/go:implement the validate-config command we planned`
