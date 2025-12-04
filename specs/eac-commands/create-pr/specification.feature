@L2 @deps:git @deps:go @ov @env:isolated-test-project @control:ai-2
Feature: eac-commands_create-pr

  As a developer who has completed work in a workspace
  I want to create a pull request with an AI-generated description
  So that my changes can be reviewed before merging to main

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "create pr"

  Rule: Validation prevents invalid PRs

    Scenario: Fail when uncommitted changes exist
      Given I am in a workspace for "feature/dirty"
      And I have uncommitted changes
      When I run "create pr"
      Then the exit code is 1
      And I see error "uncommitted changes detected"

    Scenario: Fail when in main workspace
      Given I am in the main workspace on branch "main"
      When I run "create pr"
      Then the exit code is 1
      And I see error "cannot create PR from main"
