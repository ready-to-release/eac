---
name: comprehensive-review
description: Multi-perspective code review with aggregated findings
---

# Comprehensive Code Review Skill

Conduct a thorough, multi-perspective code review by coordinating multiple specialized agents and aggregating their findings.

## Purpose

Provide comprehensive code quality assessment from multiple angles:

- **Architecture**: Design patterns, module boundaries, interfaces
- **Testing**: Coverage, test quality, edge cases
- **Debugging**: Potential issues, error handling
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

1. Gather project context using MCP tools
2. Identify scope of review (changed files or specific targets)

### Phase 2: Multi-Agent Review (5-10 min)

Run specialized agents sequentially:

1. **architect** (2-3 min)
   - Task: "Review architecture and design patterns in the changed code"
   - Focus: Module boundaries, interface design, dependency management
   - Output: Findings and recommendations

2. **test-engineer** (2-3 min)
   - Task: "Analyze test coverage and test quality for the changes"
   - Focus: Missing tests, edge cases, test clarity
   - Output: Test quality analysis

3. **debugger** (1-2 min)
   - Task: "Identify potential bugs and error handling gaps"
   - Focus: Error paths, edge cases, defensive coding
   - Output: Bug analysis

### Phase 3: Aggregation (1-2 min)

1. **Collect findings from all agents**:
   - Review each agent's output from Phase 2
   - Extract key findings and recommendations

2. **Organize findings**:
   - Group by severity and category
   - Note which agent provided each finding
   - Preserve agent attribution

3. **Prioritize**:
   - Sort by severity: critical → high → medium → low → info
   - Group by category: security, correctness, testing, architecture, style

4. **Generate consolidated report**:
   - Save to `out/comprehensive-review-<timestamp>.md`
   - Include summary, prioritized findings, next steps

### Phase 4: Presentation (1 min)

Present unified findings:

```markdown
# Comprehensive Code Review Report

**Date**: <timestamp>
**Scope**: <files or modules reviewed>
**Agents**: 3 specialized agents

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
**Location**: `file:line`
**Agent**: <agent-name>
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
| architect | 8 | 0 | 2 | 4 | 2 |
| test-engineer | 12 | 0 | 3 | 6 | 3 |
| debugger | 5 | 1 | 1 | 2 | 1 |
| **TOTAL** | **25** | **1** | **6** | **12** | **6** |

---

## Next Steps

1. **Address Critical issues immediately** (cannot merge until fixed)
2. **Review High priority findings** with team
3. **Create follow-up issues** for Medium/Low improvements
4. **Run comprehensive review again** after fixes

---

## Agent Reports

Full agent reports available at:
- Architecture: `out/architect-<timestamp>.json`
- Testing: `out/test-engineer-<timestamp>.json`
- Debugging: `out/debugger-<timestamp>.json`
```

## Configuration

### Minimum Review Scope

- At least 1 changed file
- Or explicit file/module specification

### Agent Selection Logic

**Always run**:
- architect (design patterns)
- test-engineer (test coverage)
- debugger (potential issues)

**Extensible**: Add project-specific agents as needed

### Time Budget

- **Quick review** (< 5 files): 5-8 minutes
- **Standard review** (5-20 files): 8-15 minutes
- **Comprehensive review** (20+ files): 15-25 minutes

## Output Artifacts

All outputs saved to `out/` directory:

1. **Individual agent reports** (JSON):
   - `out/architect-<timestamp>.json`
   - `out/test-engineer-<timestamp>.json`
   - `out/debugger-<timestamp>.json`

2. **Consolidated report** (Markdown):
   - `out/comprehensive-review-<timestamp>.md`

3. **Summary for user** (displayed in chat):
   - Critical issues count
   - High priority findings
   - Next steps

## Quality Bars

A comprehensive review is complete when:

- ✅ All agents executed successfully
- ✅ All findings aggregated and deduplicated
- ✅ Critical issues clearly identified
- ✅ Consolidated report generated
- ✅ Next steps provided to user

## Benefits Over Single-Agent Review

| Aspect | Single Agent | Comprehensive (This Skill) |
|--------|--------------|----------------------------|
| Coverage | 1 perspective | 3+ perspectives |
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
- Findings are organized by severity and category
- Severity levels follow standard conventions (critical → info)
- The consolidated report provides a complete multi-perspective assessment
