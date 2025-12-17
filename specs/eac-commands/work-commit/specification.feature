@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_work-commit

  As a developer working in a workspace
  I want to commit changes with AI-generated messages
  So that my commits have consistent, high-quality messages

  Rule: Validation prevents invalid commits

    Scenario: Fail when no staged changes
      Given I am in the main workspace on branch "main"
      When I run "work commit"
      Then the exit code is 1
      And I see error "No staged changes"
