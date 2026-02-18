# Intent: Allow developers to work on multiple branches simultaneously with separate Claude Code sessions by creating isolated workspaces for different features
# Architecture: Affects eac-cli work create command; creates git worktrees with standard naming convention; validates branch name is provided; depends on git worktree functionality and isolated test project

@L2 @deps:git @deps:go @ov @env:isolated-test-project
Feature: eac-cli_work-create

  As a developer using parallel Claude sessions
  I want to create isolated workspaces for different features
  So that I can work on multiple branches simultaneously with separate Claude Code sessions

  Rule: Workspace creation follows standard naming and location

    Scenario: Create workspace successfully
      Given I am in the main workspace on branch "main"
      When I run "work create feature/new-workspace"
      Then the exit code is 0
      And I see "Created worktree at"

  Rule: Validation prevents invalid workspace creation

    Scenario: Fail when no branch name provided
      Given I am in the main workspace on branch "main"
      When I run "work create"
      Then the exit code is 1
      And I see error "branch name is required"
