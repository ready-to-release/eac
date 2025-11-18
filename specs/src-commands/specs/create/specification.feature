@deps:claude @deps:go @ov
Feature: src-commands_specs-create

  As a developer of the eac platform
  I want AI-powered specification generation
  So that I can create Gherkin feature files from natural language descriptions

  Background:
    Given I am in a git repository
    And the AI provider is configured

  Rule: Command must be registered and accessible

    @L2 @iv
    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "specs-create"

    @L2 @iv
    Scenario: Command has proper description
      When I run the command "describe commands specs-create"
      Then the exit code is 0
      And I should see "specs" or "specification" or "create"

  Rule: Command validates input and configuration

    @L2 @ov
    Scenario: Command requires description argument
      When I run "specs create" without arguments
      Then the command exits with code 1
      And stderr contains "Error: description is required"
      And stderr contains "Usage: specs create <description>"

    @L2 @ov
    Scenario: Command validates AI configuration exists
      Given no .r2r/agent-config.yml file exists
      When I run "specs create Add user authentication"
      Then the command exits with code 1
      And stderr contains "agent-config.yml not found"
      And stderr contains "Please run: r2r agent init --ai <provider>"

    @L2 @ov
    Scenario: Command handles malformed AI configuration
      Given a malformed .r2r/agent-config.yml file
      When I run "specs create Add user authentication"
      Then the command exits with code 1
      And stderr contains "failed to parse agent-config.yml"

  Rule: Template resolution follows fallback hierarchy

    @L2 @ov
    Scenario: Use local template when available
      Given a template exists at ".r2r/templates/specs/specification.feature"
      When I run "specs create Add user login"
      Then the local template should be loaded
      And the template from "templates/specs/specification.feature" should not be used

    @L2 @ov
    Scenario: Fallback to built-in template when local not found
      Given no template exists at ".r2r/templates/specs/specification.feature"
      When I run "specs create Add user login"
      Then the template at "templates/specs/specification.feature" should be loaded

    @L2 @ov
    Scenario: Error when no template exists
      Given no template exists at ".r2r/templates/specs/specification.feature"
      And no template exists at "templates/specs/specification.feature"
      When I run "specs create Add user login"
      Then the command exits with code 1
      And stderr contains "template not found"

  Rule: AI agent generates specification from description

    @L2 @ov
    Scenario: Generate specification from simple description
      Given a valid specification template
      When I run "specs create Add user authentication with email and password"
      Then the AI agent is invoked with the description
      And the agent receives the template content
      And the agent generates a valid Gherkin feature file

    @L2 @ov
    Scenario: Generate specification from complex description
      Given a valid specification template
      When I run "specs create Implement multi-factor authentication supporting TOTP and SMS with rate limiting"
      Then the AI agent is invoked with the description
      And the generated specification includes multiple Rules
      And the generated specification includes multiple Scenarios

    @L2 @ov
    Scenario: AI agent output is cleaned of noise
      Given the AI provider returns output with initialization messages
      When the output is processed
      Then initialization noise should be removed
      And only valid Gherkin content should remain

  Rule: Output file naming and location

    @L2 @ov
    Scenario: Output file is created in specs directory
      Given I run "specs create Add user authentication"
      When the generation succeeds
      Then a file is created at "specs/<module>/specification.feature"
      And the file contains valid Gherkin syntax

    @L2 @ov
    Scenario: Module name is extracted from feature
      Given the AI generates a feature named "src-commands_user-authentication"
      When the output file is created
      Then the file is saved at "specs/src-commands/user-authentication/specification.feature"
      And the parent directories are created if they don't exist

    @L2 @ov
    Scenario: Handle feature name without module prefix
      Given the AI generates a feature named "user-authentication"
      When the output file is created
      Then the file is saved at "specs/user-authentication/specification.feature"

    @L2 @ov
    Scenario: Prevent overwriting existing specification
      Given a file exists at "specs/src-commands/auth/specification.feature"
      And the AI generates a feature that would create the same path
      When the output file is created
      Then the command exits with code 1
      And stderr contains "specification already exists"
      And the existing file is not modified

    @L2 @ov
    Scenario: Allow overwriting with --force flag
      Given a file exists at "specs/src-commands/auth/specification.feature"
      When I run "specs create Add authentication --force"
      Then the existing file is overwritten
      And the command exits with code 0
      And stdout contains "overwritten"

  Rule: Output validation and feedback

    @L2 @ov
    Scenario: Validate generated Gherkin syntax
      Given the AI generates specification content
      When the content is validated
      Then it must contain a "Feature:" declaration
      And it must contain at least one "Rule:" declaration
      And it must contain at least one "Scenario:" declaration

    @L2 @ov
    Scenario: Show success message with file path
      When I run "specs create Add user authentication"
      And the generation succeeds
      Then stdout contains "✅ Specification created"
      And stdout contains the file path
      And the command exits with code 0

    @L2 @ov
    Scenario: Show helpful error for AI failures
      Given the AI provider fails to generate content
      When the command is run
      Then the command exits with code 1
      And stderr contains "AI execution failed"
      And stderr contains actionable recovery instructions

  Rule: Debug mode provides visibility

    @L2 @ov
    Scenario: Debug flag saves intermediate outputs
      When I run "specs create Add authentication --debug"
      Then intermediate files are saved to "out/" directory
      And "out/debug-template.feature" contains the loaded template
      And "out/debug-prompt.md" contains the full AI prompt
      And "out/debug-raw-output.feature" contains raw AI output
      And "out/debug-cleaned-output.feature" contains cleaned output
      And debug messages are shown on stderr

    @L2 @ov
    Scenario: Debug mode shows AI execution details
      When I run "specs create Add authentication --debug"
      Then stderr contains "🔍 DEBUG: Loading template"
      And stderr contains "🔍 DEBUG: Invoking AI provider"
      And stderr contains "🔍 DEBUG: Processing output"
      And stderr contains execution timing information

  Rule: Edge cases must be handled gracefully

    @L2 @ov
    Scenario: Handle empty description
      When I run "specs create ''"
      Then the command exits with code 1
      And stderr contains "description cannot be empty"

    @L2 @ov
    Scenario: Handle very long description
      Given a description longer than 1000 characters
      When I run the specs create command
      Then the description is truncated with warning
      And the truncated description is used for generation

    @L2 @ov
    Scenario: Handle special characters in description
      When I run "specs create Add API endpoint /v1/users/{id} with rate-limiting"
      Then the description is properly escaped
      And the AI receives the full unmodified description

    @L2 @ov
    Scenario: Handle repository root not found
      Given I am not in a git repository
      When I run "specs create Add feature"
      Then the command exits with code 1
      And stderr contains "failed to find repository root"

    @L2 @ov
    Scenario: Handle file system permissions error
      Given the specs directory is not writable
      When I run "specs create Add feature"
      Then the command exits with code 1
      And stderr contains "failed to write specification"
      And stderr contains permission error details

    @L2 @ov
    Scenario: Handle AI provider timeout
      Given the AI provider takes longer than the timeout
      When the command is run
      Then the command exits with code 1
      And stderr contains "timeout"
      And stderr suggests retry or configuration adjustment
