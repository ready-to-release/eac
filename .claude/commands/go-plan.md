# Plan

```text
description: "Plan a Go CLI feature or change"
```

You are helping plan a Go CLI feature or change in this repository.

## Process

1. **Understand the request**:
   - Ask clarifying questions if needed
   - Identify whether this is a new feature, bug fix, or refactoring

2. **Delegate to go-architect agent**:
   - Use the Task tool with subagent_type="go-architect"
   - Provide the feature description or problem statement
   - Request architecture design and impact analysis

3. **Use MCP tools for context**:
   - `get-modules` to understand module structure (only if needed, avoid during boot)
   - `get-dependencies` to identify affected modules (only if needed)
   - `get-files-by-module` to locate relevant code (only if needed for ownership)

4. **Output a plan** with:
   - Feature/change summary
   - Affected modules and files
   - Architecture decisions
   - Step-by-step implementation plan
   - Testing strategy
   - Documentation updates needed

## Example Usage

User: `/go:plan add a new 'validate-config' command that checks configuration files`

## Output Format

Provide a clear, actionable plan that can guide implementation.
