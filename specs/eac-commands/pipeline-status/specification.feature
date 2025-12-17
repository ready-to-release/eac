@L2 @deps:go @deps:git @deps:gh-cli @ov @env:isolated-test-project
Feature: eac-commands_pipeline-status

  As a developer using the eac platform
  I want to see CI pipeline status for commits
  So that I can verify builds are passing before releases

  Background:
    Given I am in a git repository with GitHub remote

  Rule: Shows status for main branch HEAD by default

    @skip:wip
    Scenario: Show status for main branch
      Given GitHub Actions workflows exist for the repository
      When I run "pipeline status"
      Then the exit code is 0
      And I see workflow names and their status
      And I see the commit SHA being checked

    @skip:wip
    Scenario: Show passing workflows
      Given all workflows are passing on main
      When I run "pipeline status"
      Then the exit code is 0
      And I see status indicators showing success

    @skip:wip
    Scenario: Show failing workflows
      Given a workflow is failing on main
      When I run "pipeline status"
      Then the exit code is 0
      And I see status indicators showing failure
      And I see the failing workflow name

  Rule: Can check specific ref or commit

    @skip:wip
    Scenario: Show status for specific branch
      Given branch "develop" exists with workflows
      When I run "pipeline status --ref develop"
      Then the exit code is 0
      And I see status for the develop branch HEAD

    @skip:wip
    Scenario: Show status for specific commit
      Given a commit "abc123" has workflow runs
      When I run "pipeline status --commit abc123"
      Then the exit code is 0
      And I see status for commit "abc123"

  Rule: Error handling

    @skip:wip
    Scenario: Handle missing GitHub CLI
      Given gh CLI is not available
      When I run "pipeline status"
      Then the exit code is 1
      And stderr contains "gh" or "GitHub CLI"

    @skip:wip
    Scenario: Handle no workflows found
      Given no workflows have run for the commit
      When I run "pipeline status"
      Then the exit code is 0
      And I see "no workflow runs found" or similar message

    @skip:wip
    Scenario: Handle invalid ref
      When I run "pipeline status --ref nonexistent-branch"
      Then the exit code is 1
      And stderr contains "not found" or "invalid ref"
