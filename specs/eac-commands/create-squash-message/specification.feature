@L2 @deps:go @deps:git @ov @env:isolated-test-project @control:ai-2
Feature: eac-commands_create-squash-message

  As a developer preparing a pull request
  I want to generate a squash commit message from branch commits
  So that I can create a cohesive commit message for GitHub PR squash merge

  Background:
    Given I am in a git repository with EAC configuration
    And AI configuration exists at ".r2r/eac/ai/commit-message"

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "create squash-message"

    Scenario: Command has proper description
      When I run the command "show help create squash-message"
      Then the exit code is 0
      And I should see "squash" or "commit" or "message"

  Rule: Requires commits ahead of base branch

    Scenario: Fail when no commits ahead of base
      Given I am on a branch with no commits ahead of "main"
      When I run "create squash-message --base=main"
      Then the exit code is 1
      And stdout contains "no commits ahead"

    Scenario: Generate message from branch commits
      Given I am on a branch with multiple commits ahead of "main"
      And the mock AI is configured to return a valid squash message
      When I run "create squash-message --base=main"
      Then the exit code is 0
      And stdout contains a conventional commit message
      And the message synthesizes all commits into cohesive narrative

  Rule: Message follows conventional commit format

    Scenario: Message has proper structure
      Given I am on a branch with multiple commits ahead of "main"
      And the mock AI is configured to return a valid squash message
      When I run "create squash-message"
      Then the exit code is 0
      And the message has a type prefix
      And the message has a scope
      And the message includes "Auditor-Summary" line
      And the message includes "Changes:" statistics line

  Rule: Base branch configuration

    Scenario: Default to main branch
      Given I am on a branch with commits ahead of "main"
      And the mock AI is configured to return a valid squash message
      When I run "create squash-message"
      Then the exit code is 0
      And the command compares against "main" branch

    Scenario: Custom base branch with --base flag
      Given I am on a branch with commits ahead of "develop"
      And the mock AI is configured to return a valid squash message
      When I run "create squash-message --base=develop"
      Then the exit code is 0
      And the command compares against "develop" branch

  Rule: Debug mode logs intermediate outputs to commands.log

    Scenario: Debug flag logs outputs to unified log
      Given I am on a branch with multiple commits ahead of "main"
      And the mock AI is configured to return a valid squash message
      When I run "create squash-message --debug"
      Then the exit code is 0
      And logs are written to "out/commands.log"
      And the log contains debug artifacts for "SQUASH-CONTEXT"
      And the log contains debug artifacts for "SQUASH-PROMPT"
      And the log contains debug artifacts for "SQUASH-AI-RESPONSE"
      And the log contains debug artifacts for "SQUASH-FINAL"
