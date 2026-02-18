# Intent: Allow developers to create a cohesive commit message for GitHub PR squash merge by generating it from branch commits using AI
# Architecture: Affects eac-cli create squash-message command; reads branch commits ahead of base via git; invokes AI with .eac/ai/commit-message config; outputs conventional commit format; defaults base to main branch

@L2 @deps:go @deps:git @ov @env:isolated-test-project @control:ai-2
Feature: eac-cli_create-squash-message

  As a developer preparing a pull request
  I want to generate a squash commit message from branch commits
  So that I can create a cohesive commit message for GitHub PR squash merge

  Background:
    Given I am in a git repository with EAC configuration
    And AI configuration exists at ".eac/ai/commit-message"

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

  Rule: Base branch configuration

        Scenario: Default to main branch
      Given I am on a branch with commits ahead of "main"
      And the mock AI is configured to return a valid squash message
      When I run "create squash-message"
      Then the exit code is 0
      And the command compares against "main" branch
