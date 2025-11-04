@src-commands @git @ai @commit-message @critical
Feature: src-commands-ai-commit-generation

  As a developer
  I want AI-powered commit message generation
  So that I can create semantic commit messages that follow project contracts

  Background:
    Given I have staged changes in my git repository
    And module contracts are defined
    And commit message contract exists at version 0.1.0
    And I am in the src/commands directory

  Rule: Contract implementation must be verified before generation

    @success @ac1
    Scenario: Contract implementation verified before generation
      When I run "go run . commit-ai"
      Then VerifyContractImplementation is called first
      And contract path "../../contracts/commit-message/0.1.0/structure.yml" is checked
      And if verification passes, generation continues

    @error @ac1
    Scenario: Contract implementation failure prevents generation
      Given commit message contract is not properly implemented
      When I run "go run . commit-ai"
      Then I should see "❌ Contract implementation verification failed:" on stderr
      And contract error codes are listed
      And the exit code is 1
      And agent is not invoked

  Rule: Staged files must be collected with module mappings

    @success @ac2
    Scenario: Staged files collected with module mappings
      When I run "go run . commit-ai"
      Then GetFilesModulesReport is called with stagedOnly=true
      And staged files are built into markdown table
      And table has columns "File | Modules"

    @success @ac2
    Scenario: No staged changes shows message and exits
      Given I have no staged changes
      When I run "go run . commit-ai"
      Then I should see "No staged changes."
      And the exit code is 0
      And agent is not invoked

    @error @ac2
    Scenario: Module report error returns exit code 1
      Given GetFilesModulesReport fails
      When I run "go run . commit-ai"
      Then I should see "Error getting module mappings:" on stderr
      And the exit code is 1

  Rule: Git diff must be included in agent context

    @success @ac3
    Scenario: Git diff included in context
      When I run "go run . commit-ai"
      Then command executes "git diff --staged"
      And diff output is included in agent context
      And diff is wrapped in markdown code fence

    @error @ac3
    Scenario: Git diff error returns exit code 1
      Given git diff command fails
      When I run "go run . commit-ai"
      Then I should see "Error getting git diff:" on stderr
      And the exit code is 1

  Rule: Claude agent must be invoked with full context

    @success @ac4
    Scenario: Claude agent invoked with full context
      When I run "go run . commit-ai"
      Then agent file "../../.claude/agents/commit-message-generator.md" is read
      And context includes staged files table
      And context includes git diff
      And context includes instructions for agent
      And agent is invoked via Claude CLI

    @success @ac4
    Scenario: Agent runs with progress indicator
      When I run "go run . commit-ai"
      Then I should see "🤖 Analyzing changes and generating commit message..."
      And progress indicator wraps agent invocation

    @success @ac4
    Scenario: Agent runs in isolated session
      When I run "go run . commit-ai"
      Then Claude CLI is called without --continue flag
      And Claude CLI is called without --resume flag
      And session is isolated from previous conversations

    @success @ac4
    Scenario: Agent model extracted from frontmatter
      Given agent file has "model: sonnet" in frontmatter
      When I run "go run . commit-ai"
      Then Claude CLI is called with --model sonnet
      And fallback-model haiku is specified

    @success @ac4
    Scenario: API key removed to use subscription
      When I run "go run . commit-ai"
      Then ANTHROPIC_API_KEY is removed from environment
      And Claude CLI uses subscription auth

    @success @ac4
    Scenario: Agent context includes instructions
      When I run "go run . commit-ai"
      Then context includes "INSTRUCTIONS:" section
      And instructions explain how to use staged files table
      And instructions explain how to extract code snippets
      And instructions mention focusing on significant changes (5-15 lines per module)

    @error @ac4
    Scenario: Agent invocation failure returns exit code 1
      Given Claude CLI fails to execute
      When I run "go run . commit-ai"
      Then I should see "❌ Error running commit-message-generator:" on stderr
      And the exit code is 1

  Rule: Agent output must be cleaned automatically

    @success @ac5
    Scenario: Agent output cleaned automatically
      Given agent returns output with meta-commentary
      When I run "go run . commit-ai"
      Then AutoCleanup removes conversational wrappers
      And emojis are stripped
      And markdown fences are removed
      And pure commit message is extracted

    @success @ac5
    Scenario: Cleanup removes forbidden patterns
      Given agent output starts with "Here is the commit message:"
      When I run "go run . commit-ai"
      Then "Here is" prefix is removed
      And only commit message content remains

  Rule: Generated message must be verified against contract

    @success @ac6
    Scenario: Generated message verified against contract
      When I run "go run . commit-ai"
      Then VerifyCommitMessageContract is called on cleaned output
      And validation errors are collected
      And errors categorized by severity (error/warning)

    @success @ac6
    Scenario: Validation is silent during generation
      When I run "go run . commit-ai"
      Then validation runs after output is printed
      And validation does not interrupt message display

  Rule: File table placeholder must be injected into output

    @success @ac7
    Scenario: File table placeholder injected
      Given agent output contains "\filetable-placeholder"
      When I run "go run . commit-ai"
      Then placeholder is replaced with actual staged files table
      And replacement happens before cleanup

  Rule: Output must be wrapped with delimiters for parsing

    @success @ac8
    Scenario: Output wrapped with delimiters for parsing
      When I run "go run . commit-ai"
      Then output starts with ">>>>>>OUTPUT START<<<<<<"
      And commit message follows
      And output ends with "\n---\n"
      And delimiters allow VSCode extension to parse message

  Rule: Validation errors must be displayed after output

    @success @ac9
    Scenario: Validation errors displayed after output
      Given generated message has contract violations
      When I run "go run . commit-ai"
      Then commit message is printed first
      And then validation errors are shown
      And errors prefixed with ❌
      And error count is displayed

    @success @ac9
    Scenario: Validation warnings displayed separately
      Given generated message has warnings but no errors
      When I run "go run . commit-ai"
      Then warnings are shown with ⚠️ icon
      And warning count is displayed
      And warnings don't prevent success

    @success @ac9
    Scenario: No validation issues shows blank line only
      Given generated message passes all validations
      When I run "go run . commit-ai"
      Then only a blank line is printed after output
      And no error or warning messages shown

  Rule: Appropriate exit codes must be returned

    @success @ac10
    Scenario: Successful generation with no errors returns 0
      Given generated message passes contract validation
      When I run "go run . commit-ai"
      Then the exit code is 0

    @error @ac10
    Scenario: Contract violations return exit code 1
      Given generated message has validation errors
      When I run "go run . commit-ai"
      Then the exit code is 1
      And errors are displayed
