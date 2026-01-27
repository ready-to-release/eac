# Review

```text
description: "Review code changes before commit/PR"
```

You are reviewing Go code changes for quality, clarity, and correctness.

## Process

1. **MANDATORY: Run code-simplifier plugin FIRST**:
   - Run `/plugin code-simplifier` to invoke the plugin
   - This MUST be done before other review steps
   - Review and apply suggested changes
   - Commit simplifications separately

   **Note**: The code-simplifier is a Claude Code plugin that analyzes code and suggests simplifications for clarity and maintainability

2. **Review checklist**:
   - ✅ All tests pass (`go test ./...`)
   - ✅ Code follows Go conventions (`gofmt`, `go vet`)
   - ✅ No race conditions (`go test -race`)
   - ✅ Errors properly wrapped with %w
   - ✅ Functions are small and focused (< 40 lines ideal, < 100 max)
   - ✅ Names are clear and intention-revealing
   - ✅ No unnecessary complexity
   - ✅ Comments explain "why", not "what"
   - ✅ Public APIs have doc comments

3. **Validate with MCP tools**:
   - `validate-specs` if specs were changed
   - `validate-dependencies` if module contracts affected
   - `build` to ensure clean build

4. **Documentation check**:
   - If CLI surface changed, verify help text updated
   - If behavior changed, verify how-to guides updated
   - If architecture changed, verify design docs updated

5. **Security check**:
   - No hardcoded secrets
   - Input validation for user-provided data
   - Proper error handling (don't leak internals)

## Output

Provide a summary:

- Changes reviewed
- Issues found (if any)
- Required actions before commit/PR
- Confirmation that code-simplifier ran

## Example Usage

User: `/go:review the changes I made to the CLI parser`
