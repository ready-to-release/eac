@skip:todo
@deps:go @deps:ai @ov
Feature: src-commands_work_pr

  As a developer who has completed work in a workspace
  I want to create a pull request with an AI-generated description
  So that my changes can be reviewed before merging to main

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "work pr"

    Scenario: Command has proper description
      When I run the command "describe commands work pr"
      Then the exit code is 0
      And I should see "pull request" or "PR" or "review"

  Rule: Creates PR with AI-generated description

    Scenario: Create PR for current workspace
      Given I am in a workspace for "feature/authentication"
      And I have 3 commits on my branch
      And I have pushed my branch to origin
      When I run "r2r work pr"
      Then commit analyzes all commits to generate PR description
      And a pull request is created from "feature/authentication" to "main"
      And the PR title is generated from commit messages
      And the PR body includes AI-generated summary
      And I see the PR URL
      And the exit code is 0

    Scenario: Create PR with custom target branch
      Given I am in a workspace for "feature/new-feature"
      When I run "r2r work pr --target=develop"
      Then a PR is created targeting "develop" branch
      And I see "Created pull request targeting develop"

    Scenario: Create PR with custom title
      Given I am in a workspace
      When I run "r2r work pr --title 'Add authentication feature'"
      Then the PR is created with title "Add authentication feature"
      And the description is still AI-generated

  Rule: Handles unpushed changes automatically

    Scenario: Push branch before creating PR
      Given I am in a workspace
      And my branch has unpushed commits
      When I run "r2r work pr"
      Then the branch is pushed to origin
      And the PR is created
      And I see "Pushed branch to origin"

    Scenario: Create and push new branch
      Given I am in a workspace
      And my branch doesn't exist on remote
      When I run "r2r work pr"
      Then the branch is pushed with upstream tracking
      And the PR is created

  Rule: Validation prevents invalid PRs

    Scenario: Fail when uncommitted changes exist
      Given I am in a workspace
      And I have uncommitted changes
      When I run "r2r work pr"
      Then the exit code is 1
      And I see error "uncommitted changes"
      And I see suggestion "commit your changes first"

    Scenario: Fail when not in workspace
      Given I am in the main workspace on branch "main"
      When I run "r2r work pr"
      Then the exit code is 1
      And I see error "cannot create PR from main"

    Scenario: Fail when no commits ahead of target
      Given I am in a workspace
      And my branch has no commits ahead of main
      When I run "r2r work pr"
      Then the exit code is 1
      And I see error "no commits to create PR"

    Scenario: Fail when not in git repository
      Given I am not in a git repository
      When I run "r2r work pr"
      Then the exit code is 1
      And I see error "not in a git repository"

    Scenario: Fail when gh CLI not available
      Given I am in a workspace
      And gh CLI is not installed
      When I run "r2r work pr"
      Then the exit code is 1
      And I see error "gh CLI not found"
      And I see suggestion "install GitHub CLI"

  Rule: Provides clear progress feedback

    Scenario: Show PR creation progress
      Given I am in a workspace
      When I run "r2r work pr"
      Then I see "Checking branch status..."
      And I see "Pushing to origin..."
      And I see "Generating PR description..."
      And I see "Creating pull request..."
      And I see "✓ Pull request created"

  Rule: Debug mode shows AI details

    Scenario: Create PR with debug flag
      Given I am in a workspace
      When I run "r2r work pr --debug"
      Then commit runs in debug mode
      And debug files are created in "out/" directory
      And I see detailed AI reasoning
