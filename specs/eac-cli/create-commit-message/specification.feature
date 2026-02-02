@L2 @deps:go @deps:git @ov @env:isolated-test-project @control:ai-2
Feature: eac-cli_create-commit-message

  As a developer using the eac platform
  I want to generate commit messages using AI
  So that I can create consistent, well-structured commit messages

  Background:
    Given I am in a git repository with EAC configuration
    And AI configuration exists at ".eac/ai/commit-message"

  Rule: Requires staged changes

    Scenario: Fail when no staged changes
      Given no files are staged
      When I run "create commit-message"
      Then the exit code is 1
      And stdout contains "No staged changes"

  Rule: Message follows conventional commit format

        Scenario: Generate message for staged changes
      Given files are staged with changes
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message"
      Then the exit code is 0
      And stdout contains a conventional commit message

  Rule: Handles module mappings

        Scenario: Include module context in message
      Given files are staged in module "core"
      And the mock AI is configured to return a valid commit message
      When I run "create commit-message"
      Then the exit code is 0
      And the message references "core"
