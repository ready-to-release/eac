# Feature ID: src_mcp_vscode_semantic_commit_generation
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @critical @commit-gen
Feature: Semantic Commit Message Generation

  Background:
    Given the MCP server is initialized
    And the workspace has a valid git repository
    And there are staged changes in the repository

  @success @ac1
  Scenario: Accept execute-agent tool call with valid agentFile
    When I send a "tools/call" request with tool name "execute-agent"
    And the arguments include agentFile ".claude/agents/commit-gen.md"
    Then the request is accepted
    And the tool execution begins

  @error @ac1
  Scenario: Reject execute-agent call without agentFile parameter
    When I send a "tools/call" request with tool name "execute-agent"
    And the arguments do not include agentFile
    Then the server returns an error
    And the error indicates missing required parameter

  @success @ac2
  Scenario: Gather complete git context
    Given I call execute-agent with valid agentFile
    When the git context gathering stage begins
    Then git status is collected
    And git diff --staged is collected
    And current HEAD SHA is collected
    And file changes are parsed into FileChange objects
    And each FileChange has module attribution

  @success @ac2
  Scenario: Gather git context with multiple modules
    Given I have staged files from 3 different modules
    When the git context gathering stage begins
    Then the file changes include entries for all 3 modules
    And each file is correctly attributed to its module

  @error @ac2
  Scenario: Handle git context gathering failure
    Given the workspace is not a git repository
    When I call execute-agent with valid agentFile
    Then the server returns an error
    And the error message indicates git context failure

  @success @ac3
  Scenario: Execute full multi-agent pipeline
    Given I call execute-agent with valid agentFile
    When the pipeline executes
    Then the generator agent is called first
    And the generator produces an initial commit message
    And the reviewer agent is called with the generated commit
    And the reviewer produces feedback
    And the approver agent is called with commit and review
    And the approver produces approval decision
    And the title generator is called
    And the title generator produces a commit title

  @success @ac3
  Scenario: Pipeline passes correct context between agents
    Given I call execute-agent with valid agentFile
    When the pipeline executes
    Then the generator receives git context and documentation
    And the reviewer receives the generated commit message
    And the approver receives both commit and review feedback
    And the title generator receives the final commit body

  @success @ac4
  Scenario: Handle approval with concerns
    Given I call execute-agent with valid agentFile
    When the approver returns "Approved (with concerns)"
    Then the concerns handler agent is loaded
    And the concerns handler is called with commit and approval text
    And the concerns handler produces a corrected commit
    And the corrected commit is used for subsequent steps

  @success @ac4
  Scenario: Handle clean approval without concerns
    Given I call execute-agent with valid agentFile
    When the approver returns "Approved"
    Then the concerns handler is not invoked
    And the original commit proceeds to title generation

  @success @ac5
  Scenario: Auto-fix commit message validation errors
    Given the generated commit has validation errors
    When validation is performed
    Then the validation errors are collected
    And the fixer agent is loaded
    And the fixer agent is called with errors
    And the fixer produces a corrected commit
    And the corrected commit is re-validated
    And the final commit passes validation

  @success @ac5
  Scenario: Skip auto-fix when commit passes validation
    Given the generated commit has no validation errors
    When validation is performed
    Then the fixer agent is not invoked
    And the commit proceeds without modification

  @error @ac5
  Scenario: Report validation errors that cannot be fixed
    Given the fixer agent fails to correct errors
    When validation is re-run
    Then the server returns an error
    And the error lists the remaining validation issues

  @success @ac6
  Scenario: Return complete commit message structure
    Given I call execute-agent successfully
    When the tool execution completes
    Then the tool result contains a content array
    And the content has type "text"
    And the text contains the full commit message
    And the commit has a top-level heading
    And the commit has module sections
    And the commit ends with a newline

  @success @ac6
  Scenario: Return multi-module commit message
    Given I have staged changes in 3 modules
    When I call execute-agent successfully
    Then the returned commit has 3 module sections
    And each module section lists its changed files
    And the file table matches the git context exactly

  @success @ac7
  Scenario: Send progress notifications for each stage
    Given I call execute-agent with valid agentFile
    When the pipeline executes
    Then a progress notification is sent for "Loading generator agent"
    And a progress notification is sent for "Gathering git context"
    And a progress notification is sent for "Reading documentation"
    And a progress notification is sent for "Generating initial commit message"
    And a progress notification is sent for "Validating file completeness"
    And a progress notification is sent for "Reviewing commit message"
    And a progress notification is sent for "Approving commit message"
    And a progress notification is sent for "Generating commit title"
    And a progress notification is sent for "Validating commit message"
    And a completion notification is sent

  @success @ac7
  Scenario: Progress notifications include timing information
    Given I call execute-agent with valid agentFile
    When progress notifications are sent
    Then each notification includes elapsed time
    And the time format is "[Xs]"

  @error @ac8
  Scenario: Handle Claude CLI execution failure
    Given the Claude CLI executable fails to run
    When I call execute-agent with valid agentFile
    Then the server returns an error
    And the error message starts with "ERROR:"
    And the error describes the Claude CLI failure

  @error @ac8
  Scenario: Handle agent file not found
    When I call execute-agent with agentFile "nonexistent.md"
    Then the server returns an error
    And the error indicates the agent file does not exist

  @error @ac8
  Scenario: Handle git command failure
    Given git status command fails
    When I call execute-agent with valid agentFile
    Then the server returns an error
    And the error describes the git command failure

  @success @ac9
  Scenario: Detect missing files in commit message
    Given the git context has 5 changed files
    When the generator produces a commit with only 3 files
    Then early validation detects 2 missing files
    And a feedback prompt is built listing missing files
    And the generator is called again with feedback
    And the regenerated commit includes all 5 files

  @success @ac9
  Scenario: Detect extra files in commit message
    Given the git context has 3 changed files
    When the generator produces a commit with 5 files
    Then early validation detects 2 extra files
    And a feedback prompt is built listing extra files
    And the generator is called again with feedback
    And the regenerated commit includes only the 3 actual files

  @success @ac9
  Scenario: Skip feedback loop when files are complete
    Given the git context has 5 changed files
    When the generator produces a commit with all 5 files
    Then early validation passes
    And the feedback loop is skipped
    And the commit proceeds to reviewer stage

  @success @ac10
  Scenario: Stitch multi-agent outputs with correct structure
    Given all agents complete successfully
    When the outputs are stitched together
    Then the title is used as the top-level heading
    And the base commit body follows the heading
    And the approver status line is appended
    And the file ends with exactly one newline
    And no extra blank lines are present

  @success @ac10
  Scenario: Apply auto-correction to stitched output
    Given the stitched output has formatting issues
    When auto-correction is applied
    Then markdown linting issues are fixed
    And blank lines are added before headings
    And trailing whitespace is removed
    And the final commit is properly formatted

  @success @ac11
  Scenario: Read required documentation files
    Given I call execute-agent with valid agentFile
    When the documentation reading stage begins
    Then semantic-commits.md is read
    And versioning.md is read
    And the documentation content is captured

  @success @ac11
  Scenario: Load module contracts from deployable units
    Given I have staged files in module "src-mcp-vscode"
    When the documentation reading stage begins
    Then the contract file "contracts/deployable-units/0.1.0/src-mcp-vscode.yml" is read
    And the contract content is included in documentation

  @success @ac11
  Scenario: Pass documentation to generator agent
    Given documentation files are read successfully
    When the generator agent is called
    Then the prompt includes semantic-commits.md content
    And the prompt includes versioning.md content
    And the prompt includes relevant module contracts

  @error @ac11
  Scenario: Handle missing documentation files
    Given semantic-commits.md does not exist
    When I call execute-agent with valid agentFile
    Then the server returns an error
    And the error indicates which documentation file is missing

  @success @ac3 @ac6
  Scenario: Full pipeline with single module
    Given I have staged changes in "src-mcp-vscode" module
    When I call execute-agent with valid agentFile
    Then the generator creates a commit with src-mcp-vscode section
    And the reviewer provides feedback
    And the approver approves the commit
    And the title generator creates a concise title
    And the final commit is returned with proper structure

  @success @ac3 @ac6
  Scenario: Full pipeline with multiple modules
    Given I have staged changes in modules "src-mcp-vscode", "docs", and "cli"
    When I call execute-agent with valid agentFile
    Then the generator creates a commit with 3 module sections
    And each module section lists its files
    And the reviewer validates all modules
    And the approver approves the multi-module commit
    And the title reflects the multi-module nature
    And the final commit is returned with all 3 modules
