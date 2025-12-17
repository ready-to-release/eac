@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: eac-commands_show-workspaces

  As a developer managing multiple workspaces
  I want to see all active workspaces and their status
  So that I can understand what work is in progress across branches

  Rule: List displays all worktrees in formatted table

    @skip:wip
    Scenario: List worktrees with default output
      Given I have multiple worktrees:
        | path                          | branch           | clean |
        | C:\source\cli                 | main             | true  |
        | C:\source\cli-feature-auth    | feature/auth     | true  |
        | C:\source\cli-bugfix-123      | bugfix/issue-123 | false |
      When I run "show workspaces"
      Then the exit code is 0
      And I see a table with headers "Path", "Branch", "Status"
      And I see "3 worktrees total"

    @skip:wip
    Scenario: List shows clean status
      Given I have a worktree at "../cli-feature-auth" with no uncommitted changes
      When I run "show workspaces"
      Then I see "feature/auth" with status "clean"

    @skip:wip
    Scenario: List shows dirty status
      Given I have a worktree at "../cli-feature-auth" with uncommitted changes
      When I run "show workspaces"
      Then I see "feature/auth" with status "dirty"

    @skip:wip
    Scenario: List single worktree
      Given I have only the main worktree
      When I run "show workspaces"
      Then the exit code is 0
      And I see "1 worktree total"

  Rule: Verbose mode shows additional details

    @skip:wip
    Scenario: List with verbose flag shows commit SHA
      Given I have multiple worktrees
      When I run "show workspaces --verbose"
      Then the exit code is 0
      And I see commit SHA for each worktree

    @skip:wip
    Scenario: List with -v shorthand
      Given I have multiple worktrees
      When I run "show workspaces -v"
      Then the exit code is 0
      And I see commit SHA for each worktree

  Rule: Debug mode enables detailed logging

    @skip:wip
    Scenario: Debug flag enables logging to out/logs/work/
      Given I am in a git repository
      When I run "show workspaces --debug"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"
      And debug logs contain worktree information

    @skip:wip
    Scenario: Debug with -d shorthand
      Given I am in a git repository
      When I run "show workspaces -d"
      Then the exit code is 0
      And debug logs are written to "out/logs/work/"

  Rule: Handles edge cases gracefully

    @skip:wip
    Scenario: No worktrees except main
      Given I am in a git repository with no additional worktrees
      When I run "show workspaces"
      Then the exit code is 0
      And I see the main worktree listed

    @skip:wip
    Scenario: Fail when not in git repository
      Given I am not in a git repository
      When I run "show workspaces"
      Then the exit code is 1
      And I see error "not in a git repository"
