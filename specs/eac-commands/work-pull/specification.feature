@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: eac-commands_work_pull

  As a developer working in a workspace
  I want to sync my branch with latest changes from main
  So that I can keep my work up to date and avoid merge conflicts

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "work pull"

  Rule: Pull rebases current branch onto latest main

    Scenario: Pull successfully syncs workspace with main
      Given I am in a workspace for "feature/sync-test"
      And I have no uncommitted changes
      When I run "work pull"
      Then the exit code is 0
      And I see "Fetching origin/main"

  Rule: Validation prevents invalid operations

    Scenario: Fail when uncommitted changes exist
      Given I am in a workspace for "feature/dirty"
      And I have uncommitted changes
      When I run "work pull"
      Then the exit code is 1
      And I see error "uncommitted changes detected"

    Scenario: Fail when in main workspace
      Given I am in the main workspace on branch "main"
      When I run "work pull"
      Then the exit code is 1
      And I see error "cannot rebase main onto itself"
