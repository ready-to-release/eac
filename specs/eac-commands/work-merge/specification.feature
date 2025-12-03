@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: eac-commands_work_merge

  As a developer who has completed work in a workspace
  I want to merge my changes back to main using squash merge
  So that my feature appears as a single, well-documented commit

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "work merge"

  Rule: Validation prevents invalid merges

    Scenario: Fail when uncommitted changes exist
      Given I am in a workspace for "feature/dirty"
      And I have uncommitted changes
      When I run "work merge"
      Then the exit code is 1
      And I see error "uncommitted changes detected"

    Scenario: Fail when in main workspace
      Given I am in the main workspace on branch "main"
      When I run "work merge"
      Then the exit code is 1
      And I see error "cannot merge main into itself"
