# Feature ID: src_mcp_vscode_error_handling
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @error @critical
Feature: Error Handling

  Background:
    Given the MCP server is initialized

  @error @ac1
  Scenario: Handle git status failure in non-git directory
    Given I am in a directory that is not a git repository
    When I call execute-agent to generate a commit
    Then a JSON-RPC error is returned
    And the error message indicates "not a git repository"
    And the server does not crash

  @error @ac1
  Scenario: Handle git diff failure
    Given git diff command fails with exit code 128
    When I attempt to gather git context
    Then an error is returned
    And the error message describes the git failure
    And the error includes the exit code

  @error @ac1
  Scenario: Handle git command timeout
    Given git status hangs indefinitely
    When I attempt to gather git context with timeout
    Then the operation times out gracefully
    And an error message indicates timeout
    And the server remains responsive

  @error @ac1
  Scenario: Handle corrupted git repository
    Given the .git directory is corrupted
    When I attempt git operations
    Then git commands fail
    And error messages describe repository corruption
    And suggested action is to check repository integrity

  @error @ac2
  Scenario: Handle Claude CLI not found
    Given the Claude CLI executable is not in PATH
    When I call execute-agent
    Then an error is returned
    And the error starts with "ERROR:"
    And the error indicates Claude CLI not found
    And suggested action is to install or configure Claude CLI

  @error @ac2
  Scenario: Handle Claude CLI execution failure
    Given the Claude CLI fails with exit code 1
    When I call an agent
    Then an error is returned
    And the error includes Claude CLI output
    And the error describes what command failed

  @error @ac2
  Scenario: Handle agent file not found
    Given I call execute-agent with agentFile "nonexistent.md"
    When the agent file is loaded
    Then an error is returned
    And the error indicates the file does not exist
    And the error includes the attempted file path

  @error @ac2
  Scenario: Handle agent produces invalid output
    Given the agent returns malformed commit message
    When the output is parsed
    Then a validation error is returned
    And the error describes what is malformed
    And the original agent output is included for debugging

  @error @ac3
  Scenario: Handle documentation file not found
    Given semantic-commits.md does not exist
    When I attempt to read documentation
    Then an error is returned
    And the error indicates which file is missing
    And the error includes expected file path
    And suggested action is to create or restore the file

  @error @ac3
  Scenario: Handle contract file not found
    Given a module has no contract file in contracts/deployable-units/
    When I attempt to read module contract
    Then an error is returned
    And the error indicates missing contract
    And the error includes module name

  @error @ac3
  Scenario: Handle permission denied reading file
    Given I attempt to read a file without read permission
    When the file read is attempted
    Then an error is returned
    And the error indicates permission denied
    And the error includes file path

  @error @ac3
  Scenario: Handle directory not found
    Given workspace root directory does not exist
    When I attempt operations in that directory
    Then an error is returned
    And the error describes directory not found
    And the error includes directory path

  @error @ac4
  Scenario: Handle malformed JSON in stdin
    Given I send invalid JSON via stdin: "{broken"
    When the server parses the input
    Then a JSON-RPC parse error is returned
    And the error code is -32700
    And the error message is "Parse error"
    And the server continues listening

  @error @ac4
  Scenario: Handle empty input
    Given I send an empty line via stdin
    When the server processes the input
    Then the server ignores the empty line
    And no error is returned
    And the server continues listening

  @error @ac4
  Scenario: Handle JSON with invalid method
    Given I send valid JSON with method "invalid_method"
    When the server processes the request
    Then a JSON-RPC method not found error is returned
    And the error code is -32601
    And the error message indicates method not found

  @error @ac4
  Scenario: Handle JSON with missing required field
    Given I send JSON missing the "jsonrpc" field
    When the server processes the request
    Then a JSON-RPC invalid request error is returned
    And the error code is -32600

  @error @ac5
  Scenario: Handle commit missing top-level heading
    Given the generator produces a commit without top-level heading
    When validation is performed
    Then validation errors are collected
    And the error indicates MD041 violation
    And the error message is "First line in a file should be a top-level heading"

  @error @ac5
  Scenario: Handle commit with invalid semantic format
    Given the generator produces "## cli: add command" (missing type)
    When validation is performed
    Then a semantic format error is returned
    And the error explains the format should be "<module>: <type>: <description>"
    And the error includes the line number

  @error @ac5
  Scenario: Handle commit with missing files
    Given the commit lists 3 files but git has 5 changed files
    When validation is performed
    Then a file completeness error is returned
    And the error lists the 2 missing files by path
    And the error explains files must be 1-1 with git changes

  @error @ac5
  Scenario: Format multiple validation errors clearly
    Given the commit has 5 validation errors
    When errors are formatted for user
    Then each error is on a separate line
    And each error includes line number
    And each error includes error code
    And each error includes descriptive message
    And the format is easy to read

  @error @ac6
  Scenario: Return JSON-RPC error response structure
    Given an error occurs during request handling
    When the error response is built
    Then the response has "jsonrpc" field equal to "2.0"
    And the response has "id" matching the request
    And the response has "error" object
    And the response has no "result" field

  @error @ac6
  Scenario: Error object contains required fields
    Given an error response is returned
    When I parse the error object
    Then the error has "code" field (integer)
    And the error has "message" field (string)
    And the error optionally has "data" field

  @error @ac6
  Scenario: Use standard JSON-RPC error codes
    Given various error conditions occur
    Then parse errors use code -32700
    And invalid request uses code -32600
    And method not found uses code -32601
    And invalid params uses code -32602
    And internal errors use code -32603
    And application errors use codes >= -32000

  @error @ac7
  Scenario: Error includes code for categorization
    Given a git command failure occurs
    When the error is returned
    Then the error code indicates the error category
    And clients can programmatically handle by code

  @error @ac7
  Scenario: Error includes message for display
    Given a validation failure occurs
    When the error is returned
    Then the error message is human-readable
    And the message can be shown to users
    And the message clearly describes the problem

  @error @ac7
  Scenario: Error includes data with additional details
    Given a file completeness error occurs
    When the error is returned
    Then the error data includes missing file list
    And the error data includes extra file list
    And the error data provides context for resolution

  @error @ac8
  Scenario: Continue after git command error
    Given git status fails
    When the error is handled
    And a new request is sent
    Then the server processes the new request
    And the server is still responsive
    And no crash occurred

  @error @ac8
  Scenario: Continue after agent execution error
    Given Claude CLI fails for one request
    When the error is returned
    And another request is sent
    Then the server processes the second request
    And the error did not affect subsequent operations

  @error @ac8
  Scenario: Continue after JSON parse error
    Given I send malformed JSON
    When the parse error is returned
    And I send valid JSON
    Then the valid request is processed successfully
    And the server recovered from parse error

  @error @ac8
  Scenario: Graceful degradation under multiple errors
    Given multiple errors occur in sequence
    When each error is handled
    Then each error returns appropriate error response
    And the server never crashes
    And the server remains fully functional

  @error @ac8
  Scenario: No panics under any condition
    Given I trigger edge cases: nil pointers, empty arrays, invalid indices
    When these conditions are encountered
    Then no Go panics occur
    And all errors are handled gracefully
    And error messages are returned

  @error @ac9
  Scenario: Log error to stderr
    Given an error occurs during commit generation
    When the error is handled
    Then an error message is written to stderr
    And the log includes timestamp
    And the log includes error details

  @error @ac9
  Scenario: Log includes stack trace for internal errors
    Given an unexpected internal error occurs
    When the error is logged
    Then the log includes a stack trace
    And the stack trace helps identify the error source

  @error @ac9
  Scenario: Logs are separate from JSON-RPC output
    Given errors are logged to stderr
    And JSON-RPC responses are written to stdout
    When errors occur
    Then stderr logs do not interfere with stdout
    And JSON-RPC output remains valid
    And logs are available for debugging

  @error @ac10
  Scenario: Git error message suggests action
    Given I attempt git operations in non-git directory
    When the error message is returned
    Then the message is "Not a git repository. Initialize with 'git init' or navigate to a git repository."
    And the message explains the problem
    And the message suggests corrective action

  @error @ac10
  Scenario: Missing file error suggests action
    Given semantic-commits.md is missing
    When the error message is returned
    Then the message indicates the file path
    And the message suggests creating or restoring the file
    And the message is actionable

  @error @ac10
  Scenario: No staged changes error suggests action
    Given there are no staged changes
    When I attempt commit generation
    Then the error message is "No staged changes found. Stage files with 'git add' before generating commit."
    And the message explains how to stage files

  @error @ac10
  Scenario: Validation error suggests fix
    Given a commit has subject line too long
    When the validation error is returned
    Then the error indicates the length limit (72 chars)
    And the error shows the actual length
    And the error suggests shortening the subject

  @error @ac10
  Scenario: Agent failure suggests troubleshooting
    Given the Claude CLI execution fails
    When the error is returned
    Then the error suggests checking Claude CLI installation
    And the error suggests verifying API keys if applicable
    And the error provides troubleshooting steps

  @error @ac1 @ac2 @ac3 @ac4 @ac5 @ac6 @ac7 @ac8 @ac9 @ac10
  Scenario: Comprehensive error handling across all subsystems
    Given I test error conditions for:
      | Subsystem         | Error Condition                       | Expected Behavior                          |
      | Git               | Not a git repo                        | Clear error, suggested action, no crash    |
      | Agent             | Claude CLI fails                      | ERROR message with details, actionable     |
      | Filesystem        | File not found                        | Path included, suggestion to create        |
      | JSON-RPC          | Malformed JSON                        | Parse error -32700, continue listening     |
      | Validation        | Missing files                         | List missing files, suggest adding         |
      | Network (future)  | Timeout                               | Timeout error, retry suggestion            |
    When each error condition is triggered
    Then appropriate error response is returned
    And error follows JSON-RPC spec
    And error message is actionable
    And server does not crash
    And error is logged for debugging
    And subsequent requests work normally
