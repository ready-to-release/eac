@skip:wip @deps:go @deps:git @L2 @ov @env:isolated-test-project
Feature: eac-commands_create-risk-assessment

  As a developer
  I want to generate risk assessments from code changes
  So that I can identify potential compliance and security risks

  Background:
    Given I am in a git repository
    And specifications exist in "specs/"

  @iv
  Rule: Command must be registered and accessible

    Scenario: Command is listed
      When I run "show help"
      Then the exit code is 0
      And I should see "create risk-assessment"

    Scenario: Command shows help
      When I run "show help create risk-assessment"
      Then the exit code is 0
      And stdout contains "Generate risk assessment"
      And stdout contains "--scope"
      And stdout contains "-s"
      And stdout contains "--destination"
      And stdout contains "-d"

  Rule: Command analyzes staged changes by default

    Scenario: Analyze staged files with default settings
      Given I have staged "src/auth/handler.go"
      When I run "create risk-assessment"
      Then the exit code is 0
      And a file exists matching ".docs/reference/risk-assessment-*.md"
      And the report contains "Scope: staged"
      And the report lists "src/auth/handler.go"

    Scenario: Report filename includes timestamp
      When I run "create risk-assessment"
      Then the output file matches pattern "risk-assessment-\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}.md"

  Rule: Command reads specs directory from contracts

    Scenario: Specs directory determined from contracts
      Given repository contracts define specs directory as "specs"
      When I run "create risk-assessment"
      Then specifications are loaded from "specs/"
      And no specs directory argument is required

  Rule: All flags have shorthand versions

    Scenario: Use short flag for scope
      When I run "create risk-assessment -s changed"
      Then the report contains "Scope: changed"

    Scenario: Use short flag for destination
      When I run "create risk-assessment -d .docs/my-report.md"
      Then the output is written to ".docs/my-report.md"

    Scenario: Use short flag for prompt
      Given a custom prompt at "custom.md"
      When I run "create risk-assessment -p custom.md"
      Then the risk assessment custom prompt is used

    Scenario: Use short flag for debug
      When I run "create risk-assessment -D"
      Then intermediate outputs are saved to "out/logs/risks/"

  Rule: Command supports different scopes

    Scenario: Analyze changed files
      Given I have modified "src/auth/handler.go"
      When I run "create risk-assessment --scope changed"
      Then the exit code is 0
      And the report contains "Scope: changed"

    Scenario: Analyze all files
      When I run "create risk-assessment --scope all"
      Then the exit code is 0
      And the report contains "Scope: all"

  Rule: Report generation uses AI and contracts

    Scenario: AI generates structured report
      Given I have staged changes
      When I run "create risk-assessment --debug"
      Then the AI is invoked with file changes
      And the AI receives specification context
      And the AI generates a markdown report
      And the report follows the contract structure
      And intermediate outputs are saved to "out/logs/risks/"

    Scenario: Report includes risk IDs
      When I run "create risk-assessment"
      Then risks are assigned IDs like "RISK-001", "RISK-002"

    Scenario: Report suggests creating controls
      When I run "create risk-assessment"
      Then the report includes section "Risk Controls Needed"
      And the report mentions "create risk-controls <this-file>"

  Rule: Error handling and validation

    Scenario: No git repository
      Given I am not in a git repository
      When I run "create risk-assessment"
      Then the exit code is 1
      And stderr contains "not a git repository"

    Scenario: Invalid scope
      When I run "create risk-assessment --scope invalid"
      Then the exit code is 1
      And stderr contains "invalid scope"
      And stderr contains "must be: staged, changed, or all"

    Scenario: Path traversal prevention
      When I run "create risk-assessment --destination ../../etc/passwd.md"
      Then the exit code is 1
      And stderr contains "security error"
