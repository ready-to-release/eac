@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: eac-commands_work_remove

  As a developer who is done with a workspace
  I want to safely remove it and clean up associated branches
  So that I can keep my repository organized

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "work remove"

  Rule: Removes workspace and cleans up local branches

    Scenario: Remove current workspace
      Given I am in a workspace for "feature/test"
      When I run "work remove --force"
      Then the exit code is 0
      And I am switched to main branch
      And the local branch "feature/test" is deleted

    Scenario: Keep branch with --keep-branch flag
      Given I am in a workspace for "feature/keep-this"
      When I run "work remove --keep-branch --force"
      Then the exit code is 0
      And I am switched to main branch
      And the local branch "feature/keep-this" still exists

  Rule: Validation prevents invalid operations

    Scenario: Fail when uncommitted changes exist
      Given I am in a workspace for "feature/dirty"
      And I have uncommitted changes
      When I run "work remove"
      Then the exit code is 1
      And I see error "Uncommitted changes detected"

    Scenario: Fail when removing main workspace
      Given I am in the main workspace on branch "main"
      When I run "work remove"
      Then the exit code is 1
      And I see error "cannot remove main workspace"
