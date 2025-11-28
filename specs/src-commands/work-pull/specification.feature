@deps:git @deps:go @ov @env:isolated-test-project
Feature: src-commands_work_pull

  As a developer working in a workspace
  I want to sync my branch with latest changes from main
  So that I can keep my work up to date and avoid merge conflicts

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "work pull"

    @skip:wip
    Scenario: Command has proper description
      When I run the command "describe commands work pull"
      Then the exit code is 0
      And I should see "sync" or "rebase" or "main"

  Rule: Pull rebases current branch onto latest main

    @skip:wip
    Scenario: Pull successfully rebases onto main
      Given I am in a workspace for "feature/authentication"
      And I have 3 commits on my branch
      And I have no uncommitted changes
      And origin/main has 2 new commits
      When I run "r2r work pull"
      Then the exit code is 0
      And origin/main is fetched
      And my branch is rebased onto origin/main
      And I see "Rebased feature/authentication onto origin/main"
      And I see "Your commits: 3"
      And I see "New commits from main: 2"

    @skip:wip
    Scenario: Pull with custom target branch
      Given I am in a workspace for "feature/new-feature"
      When I run "r2r work pull --target=develop"
      Then my branch is rebased onto "origin/develop"
      And I see "Rebased feature/new-feature onto origin/develop"

    @skip:wip
    Scenario: Pull when already up to date
      Given I am in a workspace
      And main has no new commits
      When I run "r2r work pull"
      Then the exit code is 0
      And I see "Already up to date"

  Rule: Handles uncommitted changes safely

    @skip:wip
    Scenario: Fail when uncommitted changes without autostash
      Given I am in a workspace
      And I have uncommitted changes
      When I run "r2r work pull"
      Then the exit code is 1
      And I see error "uncommitted changes"
      And I see suggestion "use --autostash"

    @skip:wip
    Scenario: Autostash handles uncommitted changes
      Given I am in a workspace
      And I have uncommitted changes
      When I run "r2r work pull --autostash"
      Then the exit code is 0
      And my changes are stashed before rebase
      And my branch is rebased
      And my stashed changes are reapplied
      And I see "Stashed uncommitted changes"
      And I see "Reapplied stashed changes"

  Rule: Detects and reports conflicts clearly

    @skip:wip
    Scenario: Detect rebase conflicts
      Given I am in a workspace
      And rebasing will cause conflicts in "src/auth/handler.go"
      When I run "r2r work pull"
      Then the exit code is 1
      And I see "Rebase conflict detected"
      And I see conflicting files:
        | src/auth/handler.go |
      And I see resolution instructions

    @skip:wip
    Scenario: Conflict resolution instructions shown
      Given I am in a workspace with rebase conflicts
      When the rebase fails
      Then I see "Resolve conflicts then:"
      And I see "git add <files>"
      And I see "git rebase --continue"
      And I see "Or abort: git rebase --abort"

  Rule: Debug mode enables detailed logging

    @skip:wip
    Scenario: Debug flag enables logging to out/logs/work/
      Given I am in a workspace
      When I run "r2r work pull --debug"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"
      And debug logs contain rebase details

    @skip:wip
    Scenario: Debug with -d shorthand
      Given I am in a workspace
      When I run "r2r work pull -d"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"

  Rule: Validation prevents invalid operations

    @skip:wip
    Scenario: Fail when not in git repository
      Given I am not in a git repository
      When I run "r2r work pull"
      Then the exit code is 1
      And I see error "not in a git repository"

    @skip:wip
    Scenario: Fail when in main workspace
      Given I am in the main workspace on branch "main"
      When I run "r2r work pull"
      Then the exit code is 1
      And I see error "cannot rebase main onto itself"

  Rule: Provides clear progress feedback

    @skip:wip
    Scenario: Show rebase progress
      Given I am in a workspace
      When I run "r2r work pull"
      Then I see "Fetching origin/main..."
      And I see "Rebasing feature/authentication onto origin/main"
      And I see "Rebase successful"
