@deps:git @deps:go @L2 @ov @env:isolated-test-project
Feature: src-commands_commit-reset

  As a developer using the r2r CLI
  I want to soft reset the latest commit while preserving changes
  So that I can amend, restructure, or split my commit before pushing

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "commit"

    Scenario: Command shows help with --help flag
      When I run the command "commit reset --help"
      Then the exit code is 0
      And I should see "reset"
      And I should see "soft"

  Rule: Command must validate git environment before execution

    # Note: This scenario runs via subprocess - tests real git detection
    @skip:broken
    Scenario: Reset fails gracefully when no commits to reset
      Given I am in a git repository with only an initial commit
      When I run commit reset directly
      Then the exit code is 1

  Rule: Soft reset must preserve changes in staging area

    # Note: Uses direct execution to verify mock behavior
    Scenario: Reset latest commit successfully
      Given I am in a git repository with at least two commits
      And the latest commit contains file changes
      When I run commit reset directly
      Then the exit code is 0
      And the latest commit should be undone
      And the changes should remain staged
      And the working directory should be unchanged

    Scenario: Reset preserves staged changes from commit
      Given I am in a git repository with at least two commits
      And the latest commit added a new file "feature.txt"
      When I run commit reset directly
      Then the exit code is 0
      And "feature.txt" should be in the staging area
      And "feature.txt" should exist in the working directory

  Rule: Edge cases must be handled gracefully

    Scenario: Reset with uncommitted changes in working directory
      Given I am in a git repository with at least two commits
      And I have uncommitted changes in the working directory
      When I run commit reset directly
      Then the exit code is 0
      And the latest commit should be undone
      And the uncommitted changes should be preserved

    Scenario: Reset on detached HEAD state
      Given I am in a git repository in detached HEAD state
      And there are commits to reset
      When I run commit reset directly
      Then the exit code is 0
      And the reset should succeed
