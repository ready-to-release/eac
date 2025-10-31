# Quick Commit Generation

> **Feature ID**: src_mcp_vscode_quick_commit_generation
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, commit-gen

## User Story

* As a developer making simple changes
* I want to generate a commit message quickly without full review pipeline
* So that I can commit straightforward changes faster

## Acceptance Criteria

* Accepts quick-commit tool call with no required parameters
* Gathers git context from staged changes
* Generates commit message using single generator agent
* Skips reviewer and approver agents for speed
* Returns formatted commit message
* Validates basic commit structure

## Acceptance Tests

### AC1: Accepts quick-commit tool call with no required parameters
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, tools

* Send tools/call request with tool name "quick-commit"
* Verify request is accepted without any required parameters
* Verify process begins execution
* Verify tool schema has no required parameters

### AC2: Gathers git context from staged changes
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: git

* Call quick-commit tool
* Verify git status is collected
* Verify git diff --staged is collected
* Verify file changes are parsed
* Verify module attribution is performed

### AC3: Generates commit using single generator agent
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: critical, agents

* Call quick-commit tool
* Verify generator agent is loaded
* Verify generator is called with git context
* Verify commit message is produced
* Verify reviewer agent is NOT called
* Verify approver agent is NOT called

### AC4: Returns formatted commit message
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: critical, output

* Call quick-commit successfully
* Verify tool result contains content array
* Verify content has type "text"
* Verify text contains commit message
* Verify commit has semantic format
* Verify commit has module sections

### AC5: Validates basic commit structure
**Validated by**: behavior.feature -> @ac5 scenarios

Tags: validation

* Generate commit with quick-commit
* Verify top-level heading is present
* Verify semantic format: <module>: <type>: <description>
* Verify file table completeness
* Verify module sections are correct

### AC6: Faster execution than full pipeline
**Validated by**: behavior.feature -> @ac6 scenarios

Tags: performance

* Call quick-commit tool
* Verify fewer progress notifications than execute-agent
* Verify no review stage notification
* Verify no approval stage notification
* Verify completion is faster than full pipeline
