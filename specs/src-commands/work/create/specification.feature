@deps:go @ov
Feature: src-commands_work_create

  As a developer using parallel Claude sessions
  I want to create isolated workspaces for different features
  So that I can work on multiple branches simultaneously with separate Claude Code sessions

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "work create"

    Scenario: Command has proper description
      When I run the command "describe commands work create"
      Then the exit code is 0
      And I should see "workspace" or "worktree" or "parallel"

  Rule: Workspace creation follows standard naming and location

    Scenario: Create workspace with default settings
      Given I am in a git repository on branch "main"
      And branch "feature/authentication" does not exist
      When I run "r2r work create feature/authentication"
      Then the exit code is 0
      And a git worktree is created at "../<repo-name>-feature-authentication"
      And I see output containing "Created worktree at:"
      And I see output containing "Branch: feature/authentication"
      And I see output containing "Start Claude:"

    Scenario: Create workspace with custom path
      Given I am in a git repository
      When I run "r2r work create feature/auth --path=../custom-workspace"
      Then the exit code is 0
      And a git worktree is created at "../custom-workspace"
      And branch "feature/auth" is checked out in the workspace

    Scenario: Create workspace from specific base branch
      Given I am in a git repository
      And I have a branch "develop"
      When I run "r2r work create feature/new --from=develop"
      Then the exit code is 0
      And the workspace is created based on branch "develop"
      And branch "feature/new" is checked out

  Rule: Validation prevents invalid workspace creation

    Scenario: Fail when branch already exists
      Given I am in a git repository
      And branch "feature/existing" exists
      When I run "r2r work create feature/existing"
      Then the exit code is 1
      And I see error "branch 'feature/existing' already exists"

    Scenario: Fail when not in git repository
      Given I am not in a git repository
      When I run "r2r work create feature/test"
      Then the exit code is 1
      And I see error "not in a git repository"

    Scenario: Fail when base branch does not exist
      Given I am in a git repository
      And branch "nonexistent" does not exist
      When I run "r2r work create feature/test --from=nonexistent"
      Then the exit code is 1
      And I see error "base branch 'nonexistent' does not exist"

    Scenario: Fail when no branch name provided
      Given I am in a git repository
      When I run "r2r work create"
      Then the exit code is 1
      And I see error "branch name is required"

  Rule: Workspace path generation is consistent and predictable

    Scenario: Path sanitizes branch slashes
      Given I am in repository "my-repo"
      When I generate workspace path for "feature/auth"
      Then the path is "../my-repo-feature-auth"

    Scenario: Path uses repository name from root
      Given I am in repository "eac-cli"
      When I generate workspace path for "bugfix/issue-123"
      Then the path is "../eac-cli-bugfix-issue-123"
