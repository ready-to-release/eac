# Intent: Allow features to appear as a single, well-documented commit by merging changes back to main using squash merge
# Architecture: Affects eac-cli work merge command; validates workspace state (no uncommitted changes, not on main branch); performs squash merge into main; depends on git worktree context and isolated test project

@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: eac-cli_work_merge

  As a developer who has completed work in a workspace
  I want to merge my changes back to main using squash merge
  So that my feature appears as a single, well-documented commit

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
