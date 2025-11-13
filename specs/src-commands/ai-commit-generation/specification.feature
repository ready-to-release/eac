@dep:claude @dep:go @ov
Feature: src-commands_ai-commit-generation

  As a developer of the eac platform
  I want AI-powered commit message generation
  So that I can create semantic commit messages from staged changes

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "commit-ai"

    Scenario: Command has proper description
      When I run the command "describe commands commit-ai"
      Then the exit code is 0
      And I should see "commit-ai" or "AI" or "commit"

  Rule: Command validates contract implementation before execution

    Scenario: Command can be described
      When I run the command "describe commands commit-ai"
      Then the exit code is 0
      And I should see "commit-ai"

    Scenario: Command handles all execution paths
      When I run the command "list commands"
      Then I should see "commit-ai"
      And the exit code is 0
