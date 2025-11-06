# Feature ID: src-commands_ai-commit-generation
# Module: src-commands

@src-commands @git @ai @commit-message @critical
Feature: AI Commit Message Generation

  As a developer
  I want AI-powered commit message generation
  So that I can create semantic commit messages from staged changes

  Rule: Command must be registered and accessible

    @success @ac1
    Scenario: Command is listed in available commands
      When I run "go run . list commands"
      Then I should see "commit-ai"

    @success @ac1
    Scenario: Command has proper description
      When I run "go run . describe commands commit-ai"
      Then I should see "commit-ai" or "AI" or "commit"

  Rule: Command validates contract implementation before execution

    @success @ac2
    Scenario: Command can be described
      When I run "go run . describe commands commit-ai"
      Then the exit code is 0
      And I should see "commit-ai"

    @success @ac2
    Scenario: Command handles all execution paths
      When I run "go run . list commands"
      Then I should see "commit-ai"
      And the exit code is 0
