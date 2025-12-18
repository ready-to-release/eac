@L2 @deps:go @deps:git @ov @env:isolated-test-project @control:ai-2
Feature: eac-commands_create-commit-message

  As a developer using the eac platform
  I want to generate commit messages using AI
  So that I can create consistent, well-structured commit messages

  Background:
    Given I am in a git repository with EAC configuration
    And AI configuration exists at ".r2r/eac/ai/commit-message"

  Rule: Requires staged changes

    Scenario: Fail when no staged changes
      Given no files are staged
      When I run "create commit-message"
      Then the exit code is 1
      And stdout contains "No staged changes"

    Scenario: Generate message for staged changes
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message"
      Then the exit code is 0
      And stdout contains a conventional commit message
      And the message includes module-specific details

  Rule: Message follows conventional commit format

    Scenario: Message has proper structure
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message"
      Then the exit code is 0
      And the message has a type prefix
      And the message has a scope
      And the message has a description

    Scenario: Message validated against contract
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message"
      Then the exit code is 0
      And the message passes commit message contract validation

  Rule: Auto-commit option creates git commit

    Scenario: Create commit with --commit flag
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message --commit"
      Then the exit code is 0
      And a git commit is created
      And the commit message matches the generated message

    Scenario: Shorthand -c for commit
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message -c"
      Then the exit code is 0
      And a git commit is created

  Rule: Handles module mappings

    Scenario: Include module context in message
      Given files are staged in module "eac-core"
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message"
      Then the exit code is 0
      And the message references "eac-core"

    Scenario: Handle changes across multiple modules
      Given files are staged in modules "eac-core" and "r2r-cli"
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message"
      Then the exit code is 0
      And the message includes sections for each module
