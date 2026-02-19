# Intent: Allow developers to keep their work up to date and avoid merge conflicts by syncing their branch with the latest changes from main
# Architecture: Affects eac-cli work pull command; fetches origin/main and rebases current branch; validates no uncommitted changes exist and not on main branch; depends on git worktree context and isolated test project

@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: eac-cli_work_pull

  As a developer working in a workspace
  I want to sync my branch with latest changes from main
  So that I can keep my work up to date and avoid merge conflicts

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
