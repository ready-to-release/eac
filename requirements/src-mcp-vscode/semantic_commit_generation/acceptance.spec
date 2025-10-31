# Semantic Commit Message Generation

> **Feature ID**: src_mcp_vscode_semantic_commit_generation
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, critical, commit-gen

## User Story

* As a developer
* I want to generate a semantic commit message automatically from staged changes
* So that I can maintain consistent, high-quality commit messages across my repository

## Acceptance Criteria

* Accepts execute-agent tool call with agentFile parameter
* Gathers git context (status, diff, staged files)
* Calls generator agent to create initial commit message
* Calls reviewer agent to provide feedback
* Calls approver agent for final approval
* Handles concerns if flagged by approver
* Calls title generator to create commit title
* Validates final commit message structure
* Auto-fixes validation errors when possible
* Returns complete commit message via tool result
* Sends progress notifications for each stage

## Acceptance Tests

### AC1: Accepts execute-agent tool call with agentFile parameter
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, tools

* Send tools/call request with tool name "execute-agent"
* Include agentFile parameter in arguments
* Verify request is accepted
* Verify process begins execution

### AC2: Gathers git context from repository
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: critical, git

* Call execute-agent with valid agentFile
* Verify git status is collected
* Verify git diff is collected for staged changes
* Verify current HEAD SHA is collected
* Verify file changes are parsed and normalized

### AC3: Executes multi-agent pipeline
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: critical, agents

* Call execute-agent with valid agentFile
* Verify generator agent is called first
* Verify reviewer agent is called with generated commit
* Verify approver agent is called with commit and review
* Verify title generator is called
* Verify fixer agent is called if validation fails

### AC4: Handles approval with concerns
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: agents

* Mock approver response with "Approved (with concerns)"
* Verify concerns handler agent is loaded
* Verify concerns handler is called with commit and approval
* Verify corrected commit is used

### AC5: Validates and auto-fixes commit message
**Validated by**: behavior.feature -> @ac5 scenarios

Tags: validation

* Generate commit with validation errors
* Verify validation is performed
* Verify fixer agent is called with errors
* Verify fixed commit is re-validated
* Verify final commit passes validation

### AC6: Returns complete commit message
**Validated by**: behavior.feature -> @ac6 scenarios

Tags: critical, output

* Call execute-agent successfully
* Verify tool result is returned
* Verify result contains content array
* Verify content has type "text"
* Verify text contains full commit message
* Verify commit has top-level heading
* Verify commit has module sections

### AC7: Sends progress notifications
**Validated by**: behavior.feature -> @ac7 scenarios

Tags: progress

* Call execute-agent
* Verify progress notification sent for git context
* Verify progress notification sent for generator
* Verify progress notification sent for reviewer
* Verify progress notification sent for approver
* Verify progress notification sent for validation
* Verify completion notification sent

### AC8: Handles agent execution errors
**Validated by**: behavior.feature -> @ac8 scenarios

Tags: error

* Mock Claude CLI failure
* Call execute-agent
* Verify error is returned in tool result
* Verify error message starts with "ERROR:"
* Verify meaningful error description is provided

### AC9: Validates file completeness and provides feedback
**Validated by**: behavior.feature -> @ac9 scenarios

Tags: validation, feedback

* Create commit missing files in table
* Verify early validation detects missing files
* Verify feedback prompt is built
* Verify generator is called again with feedback
* Verify regenerated commit includes missing files

### AC10: Stitches multi-agent outputs correctly
**Validated by**: behavior.feature -> @ac10 scenarios

Tags: output

* Execute full pipeline successfully
* Verify final commit has title as top-level heading
* Verify base commit body is included
* Verify "Agent: Approved" status is added
* Verify file ends with newline
* Verify all sections are properly formatted

### AC11: Reads documentation and contracts
**Validated by**: behavior.feature -> @ac11 scenarios

Tags: documentation

* Call execute-agent
* Verify semantic-commits.md is read
* Verify versioning.md is read
* Verify module contracts are loaded from contracts/deployable-units/
* Verify documentation is passed to generator agent
