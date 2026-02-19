# Workflow Engineer

```text
description: "Analyze GitHub workflows, validate CD model compliance, diagnose issues"
```

You are analyzing, optimizing, or troubleshooting GitHub Actions CI/CD pipelines.

## Process

1. **Read the CI/CD architecture reference**:
   - Read `.github/workflows/README.md` first - it contains the complete architecture reference
   - Covers workflow mapping, module dependencies, CLI commands, `gh` commands, failure patterns, and debugging playbooks

2. **Investigate**:
   - Delegate to go-workflow-engineer agent
   - Use Task tool with subagent_type="go-workflow-engineer"
   - Provide all relevant context (workflow name, error output, run URL)

3. **For workflow failures**:
   - Check CI run status and logs
   - Identify which job/step failed
   - Cross-reference with known failure patterns in README

4. **For CD model validation**:
   - Verify workflows follow the documented CD model
   - Check module dependencies and trigger chains
   - Validate release pipeline configuration

5. **Propose changes**:
   - Explain root cause or improvement rationale
   - Provide workflow YAML changes
   - Verify changes follow documented patterns

## Example Usage

- `/go-workflow-engineer why did the CI run fail?`
- `/go-workflow-engineer validate the release workflow for module X`
- `/go-workflow-engineer optimize the test workflow`
