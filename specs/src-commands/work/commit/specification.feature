@deps:go @deps:claude @ov
Feature: src-commands_work_commit

  As a developer working in a workspace
  I want to commit changes with AI-generated messages
  So that my commits have consistent, high-quality messages

  Rule: Command must be registered and accessible

    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "work commit"

    Scenario: Command has proper description
      When I run the command "describe commands work commit"
      Then the exit code is 0
      And I should see "commit" or "AI" or "message"

  Rule: Commit uses commit-ai for message generation

    Scenario: Commit staged changes with AI message
      Given I am in a workspace
      And I have staged changes in "src/auth/handler.go"
      When I run "r2r work commit"
      Then commit-ai generates a commit message
      And the commit is created with the AI-generated message
      And the exit code is 0

    Scenario: Commit with --all stages changes first
      Given I am in a workspace
      And I have unstaged changes
      When I run "r2r work commit --all"
      Then all changes are staged
      And commit-ai generates a message
      And the commit is created

    Scenario: Commit with custom message skips AI
      Given I am in a workspace
      And I have staged changes
      When I run "r2r work commit --message 'fix: resolve bug'"
      Then the commit uses message "fix: resolve bug"
      And commit-ai is not called
      And the exit code is 0

    Scenario: Commit with -m shorthand
      Given I am in a workspace
      And I have staged changes
      When I run "r2r work commit -m 'fix: bug'"
      Then the commit uses message "fix: bug"

  Rule: Debug mode passes through to commit-ai

    Scenario: Commit with debug flag
      Given I am in a workspace
      And I have staged changes
      When I run "r2r work commit --debug"
      Then commit-ai runs in debug mode
      And debug files are created in "out/" directory

    Scenario: Commit with -d shorthand
      Given I am in a workspace
      And I have staged changes
      When I run "r2r work commit -d"
      Then commit-ai runs in debug mode

  Rule: Validation prevents invalid commits

    Scenario: Fail when no staged changes
      Given I am in a workspace
      And I have no staged or unstaged changes
      When I run "r2r work commit"
      Then the exit code is 1
      And I see error "No staged changes"

    Scenario: Fail when not in git repository
      Given I am not in a git repository
      When I run "r2r work commit"
      Then the exit code is 1
      And I see error "not in a git repository"
