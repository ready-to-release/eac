# Feature ID: src_mcp_vscode_quick_commit_generation
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @commit-gen
Feature: Quick Commit Generation

  Background:
    Given the MCP server is initialized
    And the workspace has a valid git repository
    And there are staged changes in the repository

  @success @ac1
  Scenario: Accept quick-commit call with no parameters
    When I send a "tools/call" request with tool name "quick-commit"
    And the arguments are empty
    Then the request is accepted
    And the tool execution begins

  @success @ac1
  Scenario: Quick-commit tool schema has no required parameters
    When I send a "tools/list" request
    And I find the "quick-commit" tool
    Then the tool inputSchema properties is empty or minimal
    And the tool inputSchema has no required array

  @success @ac2
  Scenario: Gather git context for quick commit
    Given I call quick-commit tool
    When the git context gathering stage begins
    Then git status is collected
    And git diff --staged is collected
    And file changes are parsed into FileChange objects
    And each FileChange has module attribution

  @success @ac2
  Scenario: Handle single module changes
    Given I have staged files in one module
    When I call quick-commit tool
    Then the git context includes files from that module
    And the module is correctly identified

  @success @ac2
  Scenario: Handle multi-module changes
    Given I have staged files in 2 different modules
    When I call quick-commit tool
    Then the git context includes files from both modules
    And each file is correctly attributed

  @error @ac2
  Scenario: Fail when no staged changes exist
    Given there are no staged changes
    When I call quick-commit tool
    Then the server returns an error
    And the error indicates no staged changes found

  @success @ac3
  Scenario: Generate commit with single agent only
    Given I call quick-commit tool
    When the generation process executes
    Then the generator agent is loaded
    And the generator is called with git context and documentation
    And the generator produces a commit message
    And the reviewer agent is NOT loaded
    And the approver agent is NOT loaded
    And the title generator is NOT called

  @success @ac3
  Scenario: Generator receives full context
    Given I call quick-commit tool
    When the generator agent is called
    Then the prompt includes git status
    And the prompt includes git diff
    And the prompt includes file changes with modules
    And the prompt includes documentation

  @success @ac4
  Scenario: Return complete quick commit message
    Given I call quick-commit successfully
    When the tool execution completes
    Then the tool result contains a content array
    And the content has type "text"
    And the text contains the full commit message
    And the commit has a top-level heading
    And the commit has semantic format

  @success @ac4
  Scenario: Return commit with single module
    Given I have staged changes in "cli" module
    When I call quick-commit successfully
    Then the returned commit has a cli module section
    And the section lists all changed files
    And the format follows semantic commit guidelines

  @success @ac4
  Scenario: Return commit with multiple modules
    Given I have staged changes in "cli" and "docs" modules
    When I call quick-commit successfully
    Then the returned commit has 2 module sections
    And each module section lists its files
    And the format is consistent across sections

  @success @ac5
  Scenario: Validate commit has top-level heading
    Given I call quick-commit successfully
    When the commit is generated
    Then the first line is a markdown heading
    And the heading uses single # character
    And the heading contains the commit subject

  @success @ac5
  Scenario: Validate semantic format in commit
    Given I call quick-commit successfully
    When the commit is generated
    Then module sections follow format "## <module>: <type>: <description>"
    And the type is one of: feat, fix, docs, refactor, test, chore
    And the description is concise and clear

  @success @ac5
  Scenario: Validate file table completeness
    Given I have 4 staged files
    When I call quick-commit successfully
    Then the commit message includes all 4 files
    And no extra files are listed
    And each file is in the correct module section

  @error @ac5
  Scenario: Detect validation errors in quick commit
    Given the generator produces an invalid commit structure
    When validation is performed
    Then validation errors are detected
    And the errors are reported in the tool result

  @success @ac6
  Scenario: Quick-commit sends fewer progress notifications
    Given I call quick-commit tool
    When the process executes
    Then progress notification is sent for "Loading generator"
    And progress notification is sent for "Gathering git context"
    And progress notification is sent for "Generating commit"
    And NO notification is sent for "Reviewing commit"
    And NO notification is sent for "Approving commit"
    And completion notification is sent

  @success @ac6
  Scenario: Quick-commit completes faster than full pipeline
    Given I have the same staged changes for both tools
    When I measure execution time of quick-commit
    And I measure execution time of execute-agent
    Then quick-commit completes in less time
    And quick-commit makes fewer agent calls

  @success @ac3 @ac4
  Scenario: Full quick-commit flow with single module
    Given I have staged changes in "src-mcp-vscode" module
    When I call quick-commit tool
    Then git context is gathered
    And generator produces commit with src-mcp-vscode section
    And the commit is returned without review/approval
    And the commit is properly formatted

  @success @ac3 @ac4
  Scenario: Full quick-commit flow with multiple modules
    Given I have staged changes in "cli", "docs", and "automation-cli" modules
    When I call quick-commit tool
    Then git context is gathered with all modules
    And generator produces commit with 3 module sections
    And each module section has correct files
    And the commit is returned without review/approval
