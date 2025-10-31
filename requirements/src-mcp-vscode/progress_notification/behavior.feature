# Feature ID: src_mcp_vscode_progress_notification
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @progress
Feature: Progress Notification

  Background:
    Given the MCP server is initialized
    And I am monitoring stdout for notifications

  @success @ac1
  Scenario: Send progress notification with $/progress method
    Given I start a commit generation operation
    When a progress notification is sent
    Then the JSON message has "method" field equal to "$/progress"
    And the message follows JSON-RPC 2.0 format
    And the message has no "id" field

  @success @ac1
  Scenario: Multiple progress notifications use $/progress
    Given I execute full pipeline
    When progress notifications are sent
    Then each notification has "method" equal to "$/progress"
    And all notifications are properly formatted

  @success @ac2
  Scenario: Include stage identifier in notification
    Given I send a progress notification for "init" stage
    When I parse the notification params
    Then the params object includes "stage" field
    And the stage value is "init"

  @success @ac2
  Scenario: Different stages have different identifiers
    Given I execute a pipeline with 5 stages
    When progress notifications are sent
    Then stage 1 has identifier "init"
    And stage 2 has identifier "git"
    And stage 3 has identifier "docs"
    And stage 4 has identifier "gen-claude"
    And stage 5 has identifier "complete"
    And each stage has a unique identifier

  @success @ac3
  Scenario: Include descriptive message for git context stage
    Given I send a progress notification for "git" stage
    When I parse the notification params
    Then the params object includes "message" field
    And the message is "Gathering git context..."

  @success @ac3
  Scenario: Include descriptive message for generator stage
    Given I send a progress notification for "gen-claude" stage
    When I parse the notification params
    Then the message is "Generating initial commit message..."

  @success @ac3
  Scenario: All messages are human-readable
    Given I execute full pipeline
    When progress notifications are sent
    Then each notification has a human-readable message
    And each message describes the current activity
    And messages use active voice with present participle

  @success @ac4
  Scenario: Include elapsed time in notification
    Given I start progress tracking at time T0
    And I wait 2 seconds
    When I send a progress notification at time T0+2
    Then the message includes elapsed time "[2s]"
    And the time format matches pattern \[\d+s\]

  @success @ac4
  Scenario: Elapsed time increases with each notification
    Given I start progress tracking
    When I send notification 1 at 1 second
    And I send notification 2 at 3 seconds
    And I send notification 3 at 5 seconds
    Then notification 1 shows "[1s]"
    And notification 2 shows "[3s]"
    And notification 3 shows "[5s]"

  @success @ac4
  Scenario: Time format is consistent
    Given I execute operations of varying duration
    When progress notifications include timing
    Then all times match format "[Xs]"
    And X is an integer representing seconds

  @success @ac5
  Scenario: Send notification for loading generator agent
    Given I start semantic commit generation
    When the generator agent is being loaded
    Then a progress notification is sent
    And the message is "Loading generator agent..."

  @success @ac5
  Scenario: Send notification for gathering git context
    Given I start commit generation
    When git context gathering begins
    Then a progress notification is sent
    And the message is "Gathering git context..."

  @success @ac5
  Scenario: Send notification for reading documentation
    Given I start commit generation
    When documentation files are being read
    Then a progress notification is sent
    And the message is "Reading documentation..."

  @success @ac5
  Scenario: Send notification for generating commit
    Given I start commit generation
    When the generator agent is called
    Then a progress notification is sent
    And the message is "Generating initial commit message..."

  @success @ac5
  Scenario: Send notification for validating file completeness
    Given I start commit generation
    When file completeness validation runs
    Then a progress notification is sent
    And the message is "Validating file completeness..."

  @success @ac5
  Scenario: Send notification for reviewing commit
    Given I start commit generation
    When the reviewer agent is called
    Then a progress notification is sent
    And the message is "Reviewing commit message..."

  @success @ac5
  Scenario: Send notification for approving commit
    Given I start commit generation
    When the approver agent is called
    Then a progress notification is sent
    And the message is "Approving commit message..."

  @success @ac5
  Scenario: Send notification for generating title
    Given I start commit generation
    When the title generator is called
    Then a progress notification is sent
    And the message is "Generating commit title..."

  @success @ac5
  Scenario: Send notification for final validation
    Given I start commit generation
    When final validation is performed
    Then a progress notification is sent
    And the message is "Validating commit message..."

  @success @ac5
  Scenario: Count notifications in full pipeline
    Given I execute full semantic commit generation
    When I count progress notifications
    Then at least 9 notifications are sent
    And each pipeline stage has a notification

  @success @ac6
  Scenario: Send completion notification
    Given I start commit generation
    When the operation completes successfully
    Then a completion notification is sent
    And the message indicates completion
    And the message includes total elapsed time

  @success @ac6
  Scenario: Completion notification is distinguishable
    Given I execute commit generation
    When I receive all notifications
    Then the final notification is clearly a completion
    And it is distinct from progress notifications
    And it indicates success

  @success @ac6
  Scenario: Completion includes final timing
    Given I execute commit generation taking 8 seconds
    When the completion notification is sent
    Then the message includes "[8s]"
    And the timing represents total duration

  @success @ac7
  Scenario: Notification has jsonrpc field
    Given I send a progress notification
    When I parse the JSON
    Then the "jsonrpc" field equals "2.0"

  @success @ac7
  Scenario: Notification has method field
    Given I send a progress notification
    When I parse the JSON
    Then the "method" field equals "$/progress"

  @success @ac7
  Scenario: Notification has params object
    Given I send a progress notification
    When I parse the JSON
    Then the "params" field is an object
    And the params object is not empty

  @success @ac7
  Scenario: Notification has no id field
    Given I send a progress notification
    When I parse the JSON
    Then the message has no "id" field
    And the message is a notification, not a request

  @success @ac7
  Scenario: Notification is valid JSON
    Given I send a progress notification
    When I parse the output
    Then the JSON parsing succeeds
    And the structure is valid JSON-RPC 2.0

  @success @ac8
  Scenario: Notification does not block execution
    Given I send a progress notification
    When the notification is written to stdout
    Then execution continues immediately
    And no response is awaited
    And the next operation starts without delay

  @success @ac8
  Scenario: Multiple notifications fire rapidly
    Given I have a pipeline with 10 stages
    When progress notifications are sent
    Then all notifications fire without waiting
    And the pipeline completes quickly
    And no notification blocks the next

  @success @ac8
  Scenario: Notification failure does not crash pipeline
    Given stdout write fails for a notification
    When the pipeline continues
    Then the error is handled gracefully
    And the next operation proceeds
    And the pipeline completes successfully

  @success @ac1 @ac2 @ac3 @ac4 @ac5 @ac6 @ac7
  Scenario: Full notification sequence for semantic commit
    Given I execute full semantic commit generation
    When I capture all progress notifications
    Then I receive notifications in order:
      | Stage      | Message                                 | Has Timing |
      | init       | Loading generator agent...              | Yes        |
      | git        | Gathering git context...                | Yes        |
      | docs       | Reading documentation...                | Yes        |
      | gen-claude | Generating initial commit message...    | Yes        |
      | gen-validate | Validating file completeness...       | Yes        |
      | review     | Reviewing commit message...             | Yes        |
      | approve    | Approving commit message...             | Yes        |
      | title      | Generating commit title...              | Yes        |
      | validate   | Validating commit message...            | Yes        |
      | complete   | Commit generation complete              | Yes        |
    And each notification follows JSON-RPC 2.0 format
    And each notification has increasing elapsed time
    And all notifications use "$/progress" method

  @success @ac5 @ac8
  Scenario: Quick-commit sends fewer notifications
    Given I execute quick-commit
    When I count progress notifications
    Then fewer than 7 notifications are sent
    And no review or approval notifications are sent
    And the sequence completes faster

  @success @ac4 @ac8
  Scenario: Notifications provide accurate time tracking
    Given I execute a commit generation
    When I track elapsed time externally
    And I track elapsed time via notifications
    Then the notification times match actual elapsed time
    And the timing is accurate within 1 second
    And users can trust the timing information
