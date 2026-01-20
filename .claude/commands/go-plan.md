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
   - **Note**: go-architect uses extended thinking mode for deeper analysis

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

5. **Save plan to `out/` folder**:
   - All plan documents MUST be saved to the `out/` folder
   - Use descriptive filenames: `out/<feature-name>-plan.md`
   - Example: `out/validate-config-command-plan.md`
   - The `out/` folder is for intermediate and planning documents
   - Never save plans to module directories or repository root

6. **Suggest clearing context**:
   - After completing the plan, explicitly suggest the user clear context
   - Prompt: "The plan is complete and saved to `out/<filename>.md`. I recommend clearing context now using `/clear` so you can start implementation with a fresh context window while keeping the plan. This helps maintain focus during implementation."
   - This is a best practice: Research/Plan → Clear → Implement

## Why Extended Thinking for Planning

The go-architect agent uses **extended thinking mode** because:

- **Planning is foundational**: Good architectural decisions prevent technical debt and make implementation smoother
- **Trade-off analysis**: Extended thinking helps evaluate multiple design approaches thoroughly
- **Impact anticipation**: Deeper reasoning identifies downstream effects and edge cases
- **Cost-effective**: More time thinking during planning saves significant time during implementation and debugging

**When extended thinking adds most value**:

- ✅ Cross-module changes with complex dependencies
- ✅ New abstractions or interface design
- ✅ Significant refactoring with potential breaking changes
- ✅ Performance-critical architecture decisions
- ✅ Security-sensitive design patterns

**When quick planning is sufficient**:

- Small, localized bug fixes
- Minor utility function additions
- Documentation-only changes
- Simple configuration updates

## Example Usage

User: `/go:plan add a new 'validate-config' command that checks configuration files`

## Output Format

Provide a clear, actionable plan that can guide implementation.
