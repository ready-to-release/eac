# Intent: Enable changes to be reviewed before merging to main by creating a pull request with an AI-generated description
# Architecture: Affects eac-cli create-pr command; validates workspace state (no uncommitted changes, not on main branch); invokes AI to generate PR description; depends on git and GitHub integration

@L2 @deps:git @deps:go @ov @env:isolated-test-project @control:ai-2
Feature: eac-cli_create-pr

  As a developer who has completed work in a workspace
  I want to create a pull request with an AI-generated description
  So that my changes can be reviewed before merging to main

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
