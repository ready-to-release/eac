@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: src-commands_work_merge

  As a developer who has completed work in a workspace
  I want to merge my changes back to main using squash merge
  So that my feature appears as a single, well-documented commit

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "work merge"

    Scenario: Command has proper description
      When I run the command "show help work merge"
      Then the exit code is 0
      And I should see "merge" or "squash" or "main"

  Rule: Default squash merge creates single commit with AI message

    @skip:wip
    Scenario: Merge workspace with squash (default)
      Given I am in a workspace for "feature/authentication"
      And I have 5 commits on my branch
      And I have no uncommitted changes
      And my branch is up to date with main
      When I run "r2r work merge"
      Then I am switched to branch "main"
      And main is updated from remote
      And commit generates a squash message from all 5 commits
      And a single squash commit is created on main
      And I see "Merged feature/authentication into main (squash)"
      And the workspace is removed
      And I am in the main workspace

    @skip:wip
    Scenario: Merge with --target specifies different branch
      Given I am in a workspace for "feature/new-feature"
      When I run "r2r work merge --target=develop"
      Then I am switched to branch "develop"
      And my branch is squash merged into develop
      And I see "Merged feature/new-feature into develop (squash)"

  Rule: Regular merge available with --no-squash flag

    @skip:wip
    Scenario: Merge with --no-squash preserves commits
      Given I am in a workspace for "feature/authentication"
      And I have 5 commits on my branch
      When I run "r2r work merge --no-squash"
      Then I am switched to branch "main"
      And all 5 commits are merged into main
      And I see "Merged feature/authentication into main (fast-forward)"
      And the workspace is removed

  Rule: Workspace cleanup is configurable

    @skip:wip
    Scenario: Keep workspace with --keep-worktree
      Given I am in a workspace for "feature/test"
      When I run "r2r work merge --keep-worktree"
      Then the merge completes successfully
      And the workspace is NOT removed
      And I see "Workspace preserved at"

    @skip:wip
    Scenario: Remove workspace after successful merge (default)
      Given I am in a workspace for "feature/test"
      When I run "r2r work merge"
      Then the workspace is removed after merge
      And I see "Removed workspace"

  Rule: Validation prevents invalid merges

    @skip:wip
    Scenario: Fail when uncommitted changes exist
      Given I am in a workspace
      And I have uncommitted changes
      When I run "r2r work merge"
      Then the exit code is 1
      And I see error "uncommitted changes"
      And I see suggestion "commit or stash"

    @skip:wip
    Scenario: Fail when not in workspace
      Given I am in the main workspace on branch "main"
      When I run "r2r work merge"
      Then the exit code is 1
      And I see error "cannot merge main into itself"

    @skip:wip
    Scenario: Fail when branch not up to date
      Given I am in a workspace
      And origin/main has new commits I don't have
      When I run "r2r work merge"
      Then the exit code is 1
      And I see error "branch not up to date"
      And I see suggestion "run 'work pull' first"

    @skip:wip
    Scenario: Fail when not in git repository
      Given I am not in a git repository
      When I run "r2r work merge"
      Then the exit code is 1
      And I see error "not in a git repository"

  Rule: Debug mode enables detailed logging

    @skip:wip
    Scenario: Merge with debug flag
      Given I am in a workspace
      When I run "r2r work merge --debug"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"
      And debug logs contain merge details

    @skip:wip
    Scenario: Merge with -d shorthand
      Given I am in a workspace
      When I run "r2r work merge -d"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"

  Rule: Provides clear progress feedback

    @skip:wip
    Scenario: Show merge progress
      Given I am in a workspace
      When I run "r2r work merge"
      Then I see "Switching to main..."
      And I see "Updating main from remote..."
      And I see "Generating commit message..."
      And I see "Creating squash commit..."
      And I see "Merge successful"

  Rule: Handle merge conflicts gracefully

    @skip:wip
    Scenario: Detect merge conflicts
      Given I am in a workspace
      And merging will cause conflicts
      When I run "r2r work merge"
      Then the exit code is 1
      And I see "Merge conflict detected"
      And I see resolution instructions
      And the workspace is NOT removed
      And I am still on my workspace branch
