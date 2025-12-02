@L2 @deps:go @deps:git @ov @env:isolated-test-project
Feature: eac-commands_commit-message

  As a developer using the eac platform
  I want to generate commit messages using AI
  So that I can create consistent, well-structured commit messages

  Background:
    Given I am in a git repository with EAC configuration
    And AI configuration exists at ".r2r/eac/ai/commit"

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "commit message"

    Scenario: Command has proper description
      When I run the command "show help commit message"
      Then the exit code is 0
      And I should see "commit" or "message" or "AI"

  Rule: Requires staged changes

    @skip:wip
    Scenario: Fail when no staged changes
      Given no files are staged
      When I run "commit message"
      Then the exit code is 1
      And stderr contains "no staged changes"

    @skip:wip
    Scenario: Generate message for staged changes
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "commit message"
      Then the exit code is 0
      And stdout contains a conventional commit message
      And the message includes module-specific details

  Rule: Message follows conventional commit format

    @skip:wip
    Scenario: Message has proper structure
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "commit message"
      Then the exit code is 0
      And the message has a type prefix
      And the message has a scope
      And the message has a description

    @skip:wip
    Scenario: Message validated against contract
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "commit message"
      Then the exit code is 0
      And the message passes commit message contract validation

  Rule: Auto-commit option creates git commit

    @skip:wip
    Scenario: Create commit with --commit flag
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "commit message --commit"
      Then the exit code is 0
      And a git commit is created
      And the commit message matches the generated message

    @skip:wip
    Scenario: Shorthand -c for commit
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "commit message -c"
      Then the exit code is 0
      And a git commit is created

  Rule: Debug mode saves intermediate files

    @skip:wip
    Scenario: Debug flag saves outputs
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "commit message --debug"
      Then the exit code is 0
      And debug files are saved to "out/logs/commit/"
      And debug files include the AI prompt
      And debug files include the AI response

    @skip:wip
    Scenario: Shorthand -d for debug
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "commit message -d"
      Then the exit code is 0
      And debug files are saved to "out/logs/commit/"

  Rule: Handles module mappings

    @skip:wip
    Scenario: Include module context in message
      Given files are staged in module "eac-core"
      And the mock AI is configured to return a valid commit message
      When I run "commit message"
      Then the exit code is 0
      And the message references "eac-core"

    @skip:wip
    Scenario: Handle changes across multiple modules
      Given files are staged in modules "eac-core" and "r2r-cli"
      And the mock AI is configured to return a valid commit message
      When I run "commit message"
      Then the exit code is 0
      And the message includes sections for each module
