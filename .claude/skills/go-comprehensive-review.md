---
name: go-comprehensive-review
description: Multi-perspective code review with aggregated findings
---

# Comprehensive Code Review Skill

Conduct a thorough, multi-perspective code review by coordinating multiple specialized agents and aggregating their findings.

## Purpose

Provide comprehensive code quality assessment from multiple angles:

- **Architecture**: Design patterns, module boundaries, interfaces
- **Testing**: Coverage, test quality, edge cases
- **Security**: Vulnerabilities, compliance, best practices
- **UX**: CLI ergonomics, error messages, help text
- **Simplicity**: Code clarity, complexity reduction

## When to Use This Skill

**Use for critical code paths**:
- Pre-PR reviews for important features
- Security-sensitive changes
- Complex refactorings
- Public API changes
- Before major releases

**Don't use for**:
- Simple bug fixes
- Documentation-only changes
- Trivial updates
- Draft/WIP code

## How It Works

### Phase 1: Context Loading (1-2 min)

1. Check for cached session context (`out/claude/session-context.json`)
2. If missing, gather project context once
3. Share context with all agents

### Phase 2: Multi-Agent Review (5-10 min)

Run 5 specialized agents sequentially:

1. **go-architect** (2-3 min)
   - Task: "Review architecture and design patterns in the changed code"
   - Focus: Module boundaries, interface design, dependency management
   - Output: `out/go-architect-<timestamp>.json`

2. **go-test-engineer** (2-3 min)
   - Task: "Analyze test coverage and test quality for the changes"
   - Focus: Missing tests, edge cases, test clarity
   - Output: `out/go-test-engineer-<timestamp>.json`

3. **go-security-release** (2-3 min)
   - Task: "Scan for security vulnerabilities and compliance issues"
   - Focus: Security risks, dependency vulnerabilities, hardcoded secrets
   - Output: `out/go-security-release-<timestamp>.json`

4. **go-cli-ux** (1-2 min) - If CLI changes detected
   - Task: "Review CLI user experience and error handling"
   - Focus: Command ergonomics, help text, error messages
   - Output: `out/go-cli-ux-<timestamp>.json`

5. **code-simplifier** (2-3 min)
   - Task: "Analyze code complexity and identify simplification opportunities"
   - Focus: Code clarity, naming, unnecessary complexity
   - Output: Inline recommendations

### Phase 3: Aggregation (1-2 min)

1. **Load all JSON reports**:
   - Read all `out/*-<timestamp>.json` files from Phase 2
   - Parse findings from each agent

2. **Deduplicate findings**:
   - Group by location (file:line)
   - Merge duplicate observations
   - Preserve agent attribution

3. **Prioritize**:
   - Sort by severity: critical → high → medium → low → info
   - Group by category: security, correctness, testing, architecture, ux, style

4. **Generate consolidated report**:
   - Save to `out/comprehensive-review-<timestamp>.md`
   - Include summary, prioritized findings, next steps

### Phase 4: Presentation (1 min)

Present unified findings:

```markdown
# Comprehensive Code Review Report

**Date**: <timestamp>
**Scope**: <files or modules reviewed>
**Agents**: 5 specialized agents

---

## Executive Summary

- **Total Findings**: X
- **Critical**: X (MUST FIX before merge)
- **High**: X (SHOULD FIX before merge)
- **Medium**: X (Consider fixing)
- **Low/Info**: X (Optional improvements)

---

## Critical Issues (MUST FIX)

### [Category] Issue Title
**Location**: `file.go:123`
**Agent**: go-security-release
**Severity**: Critical

Description of the issue...

**Recommendation**: Specific fix steps...

---

## High Priority (SHOULD FIX)

[Same format as Critical]

---

## Medium Priority (Consider)

[Same format]

---

## Suggestions (Optional)

[Same format]

---

## Metrics by Agent

| Agent | Findings | Critical | High | Medium | Low |
|-------|----------|----------|------|--------|-----|
| go-architect | 8 | 0 | 2 | 4 | 2 |
| go-test-engineer | 12 | 0 | 3 | 6 | 3 |
| go-security-release | 5 | 1 | 1 | 2 | 1 |
| go-cli-ux | 4 | 0 | 0 | 2 | 2 |
| code-simplifier | 10 | 0 | 0 | 5 | 5 |
| **TOTAL** | **39** | **1** | **6** | **19** | **13** |

---

## Next Steps

1. **Address Critical issues immediately** (cannot merge until fixed)
2. **Review High priority findings** with team
3. **Create follow-up issues** for Medium/Low improvements
4. **Run comprehensive review again** after fixes

---

## Agent Reports

Full agent reports available at:
- Architecture: `out/go-architect-<timestamp>.json`
- Testing: `out/go-test-engineer-<timestamp>.json`
- Security: `out/go-security-release-<timestamp>.json`
- UX: `out/go-cli-ux-<timestamp>.json`
```

## Usage Examples

### Example 1: Pre-PR Review

```text
User: /go:comprehensive-review
Assistant: Starting comprehensive code review...

Phase 1: Loading context... (checking out/claude/session-context.json)
Phase 2: Running 5 specialized agents...
  - go-architect reviewing architecture... (2m 15s)
  - go-test-engineer analyzing tests... (2m 42s)
  - go-security-release scanning security... (1m 58s)
  - go-cli-ux reviewing CLI... (1m 23s)
  - code-simplifier analyzing complexity... (2m 10s)
Phase 3: Aggregating findings... (1m 05s)
Phase 4: Generating report...

Review complete! Found:
- 1 Critical issue (security)
- 6 High priority findings
- 19 Medium priority suggestions
- 13 Low priority improvements

Report saved to: out/comprehensive-review-2026-01-22-143052.md

Critical Issue Summary:
❌ Hardcoded API key in config.go:45 (go-security-release)

High Priority Summary:
⚠️ Missing error handling in 3 locations (go-architect)
⚠️ Untested edge case in parser (go-test-engineer)
⚠️ Confusing error message for users (go-cli-ux)
...

Next Steps:
1. Fix hardcoded API key immediately
2. Review 6 high priority findings
3. Create issues for deferred improvements
```

### Example 2: Selective Review (Specific Files)

```text
User: Run comprehensive review on go/cli/eac/impl/release/*.go
Assistant: Starting targeted comprehensive review...

Scope: go/cli/eac/impl/release/*.go (8 files)
[Same phases as above, but focused on specified files]
```

## Configuration

### Minimum Review Scope

- At least 1 changed file
- Or explicit file/module specification

### Agent Selection Logic

**Always run**:
- go-architect (design patterns)
- go-test-engineer (test coverage)
- go-security-release (security)
- code-simplifier (complexity)

**Conditional**:
- go-cli-ux: Only if CLI code detected (cobra, commands, flags)
- go-workflow-engineer: Only if workflow files changed (.github/workflows/*.yaml)

### Time Budget

- **Quick review** (< 5 files): 5-8 minutes
- **Standard review** (5-20 files): 8-15 minutes
- **Comprehensive review** (20+ files): 15-25 minutes

## Output Artifacts

All outputs saved to `out/` directory:

1. **Individual agent reports** (JSON):
   - `out/go-architect-<timestamp>.json`
   - `out/go-test-engineer-<timestamp>.json`
   - `out/go-security-release-<timestamp>.json`
   - `out/go-cli-ux-<timestamp>.json`

2. **Consolidated report** (Markdown):
   - `out/comprehensive-review-<timestamp>.md`

3. **Summary for user** (displayed in chat):
   - Critical issues count
   - High priority findings
   - Next steps

## Integration with Existing Skills

### Use Before These Skills

- `go-cli-feature`: Before starting implementation
- `go-cli-refactor-safe`: Before major refactoring

### Use After These Skills

- After implementing a feature (before creating PR)
- After completing refactoring (verify no regressions)

### Complements These Commands

- `/go:review`: Single-agent review (faster, less thorough)
- `/go:release`: Release readiness (includes security scan)

## Quality Bars

A comprehensive review is complete when:

- ✅ All 5 agents executed successfully
- ✅ All findings aggregated and deduplicated
- ✅ Critical issues clearly identified
- ✅ Consolidated report generated
- ✅ Next steps provided to user

## Benefits Over Single-Agent Review

| Aspect | Single Agent | Comprehensive (This Skill) |
|--------|--------------|----------------------------|
| Coverage | 1 perspective | 5 perspectives |
| Time | 2-3 min | 8-15 min |
| Findings | 5-10 issues | 20-50 issues |
| Depth | Surface-level | Multi-layered |
| Prioritization | Manual | Automated aggregation |
| Best for | Simple changes | Critical code |

## Future Enhancements

When Claude Code supports parallel agent execution:

- **Phase 2 becomes parallel**: 10 min → 3 min (70% reduction)
- **All agents run simultaneously**
- **Aggregation starts immediately after fastest agent completes**
- **Total time**: 8-15 min → 4-6 min

## Notes

- This skill runs agents **sequentially** (no parallel execution yet)
- Each agent leverages shared context cache for efficiency
- Findings are automatically deduplicated by location
- Severity levels follow standard conventions (critical → info)
- All reports follow `.claude/schemas/agent-result.json` schema
