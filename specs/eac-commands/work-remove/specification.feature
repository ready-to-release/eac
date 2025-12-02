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

    Scenario: Command has proper description
      When I run the command "show help work remove"
      Then the exit code is 0
      And I should see "remove" or "delete" or "cleanup"

  Rule: Removes workspace and cleans up branches

    @skip:wip
    Scenario: Remove current workspace
      Given I am in a workspace for "feature/test"
      And I have no uncommitted changes
      When I run "r2r work remove"
      Then I am switched to main branch
      And the workspace is removed
      And the local branch "feature/test" is deleted
      And I see "✓ Removed workspace"
      And I am in the main workspace
      And the exit code is 0

    @skip:wip
    Scenario: Remove specific workspace by branch name
      Given I am in the main workspace
      And a workspace exists for "feature/old-feature"
      When I run "r2r work remove feature/old-feature"
      Then the workspace for "feature/old-feature" is removed
      And the branch "feature/old-feature" is deleted
      And I see "✓ Removed workspace for feature/old-feature"

    @skip:wip
    Scenario: Keep branch with --keep-branch flag
      Given I am in a workspace for "feature/keep-this"
      When I run "r2r work remove --keep-branch"
      Then the workspace is removed
      And the branch "feature/keep-this" is NOT deleted
      And I see "Branch feature/keep-this preserved"

  Rule: Remote branch cleanup is optional

    @skip:wip
    Scenario: Delete remote branch with --delete-remote
      Given I am in a workspace for "feature/pushed"
      And the branch exists on remote
      When I run "r2r work remove --delete-remote"
      Then the workspace is removed
      And the local branch is deleted
      And the remote branch is deleted
      And I see "Deleted remote branch"

    @skip:wip
    Scenario: Keep remote branch by default
      Given I am in a workspace for "feature/pushed"
      And the branch exists on remote
      When I run "r2r work remove"
      Then the remote branch is NOT deleted
      And I see "Remote branch preserved"

  Rule: Safely handles uncommitted changes

    @skip:wip
    Scenario: Fail when uncommitted changes exist
      Given I am in a workspace
      And I have uncommitted changes
      When I run "r2r work remove"
      Then the exit code is 1
      And I see error "uncommitted changes"
      And I see suggestion "commit, stash, or use --force"
      And the workspace is NOT removed

    @skip:wip
    Scenario: Force remove with uncommitted changes
      Given I am in a workspace
      And I have uncommitted changes
      When I run "r2r work remove --force"
      Then I see warning "Uncommitted changes will be lost"
      And the workspace is removed
      And uncommitted changes are discarded

  Rule: Debug mode enables detailed logging

    @skip:wip
    Scenario: Debug flag enables logging to out/logs/work/
      Given I am in a workspace
      When I run "r2r work remove --debug"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"
      And debug logs contain workspace removal details

    @skip:wip
    Scenario: Debug with -d shorthand
      Given I am in a workspace
      When I run "r2r work remove -d"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"

  Rule: Validation prevents invalid operations

    @skip:wip
    Scenario: Fail when not in git repository
      Given I am not in a git repository
      When I run "r2r work remove"
      Then the exit code is 1
      And I see error "not in a git repository"

    @skip:wip
    Scenario: Fail when removing main workspace
      Given I am in the main workspace on branch "main"
      When I run "r2r work remove"
      Then the exit code is 1
      And I see error "cannot remove main workspace"

    @skip:wip
    Scenario: Fail when workspace doesn't exist
      Given I am in the main workspace
      When I run "r2r work remove feature/nonexistent"
      Then the exit code is 1
      And I see error "workspace not found"

  Rule: Provides clear progress feedback

    @skip:wip
    Scenario: Show removal progress
      Given I am in a workspace
      When I run "r2r work remove"
      Then I see "Checking workspace status..."
      And I see "Switching to main..."
      And I see "Removing workspace..."
      And I see "Deleting branch..."
      And I see "✓ Cleanup complete"
