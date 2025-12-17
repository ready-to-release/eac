@deps:go @env:isolated-test-project @skip:broken @control:ai-2 @control:si-10
Feature: eac-commands_create-spec

  As a developer of the eac platform
  I want AI-powered specification generation
  So that I can create Gherkin feature files from natural language descriptions

  Background:
    Given I am in a git repository

  Rule: Command validates input

    @L2 @ov
    Scenario: Command requires description argument
      When I run "create spec" without arguments
      Then the command exits with code 1
      And stderr contains "description is required"

    @L2 @ov
    Scenario: Long descriptions are truncated with warning
      Given a description longer than 1000 characters
      And the mock AI is configured to return a valid specification
      When I run the create spec command
      Then stderr contains "truncated"

  Rule: AI generates valid Gherkin specifications

    @L2 @ov
    Scenario: AI generates specification from description
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec 'Add user authentication with email and password'"
      Then the exit code is 0
      And stdout contains "Specification created"
      And a specification file is created

    @L2 @ov
    Scenario: AI output noise is filtered
      Given the AI provider returns output with initialization messages
      When the output is processed
      Then initialization noise should be removed
      And only valid Gherkin content should remain

    @L2 @ov
    Scenario: Generated specification must contain required Gherkin elements
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec 'Test feature generation'"
      Then the exit code is 0
      And it must contain a "Feature:" declaration
      And it must contain at least one "Rule:" declaration
      And it must contain at least one "Scenario:" declaration

  Rule: Output path is determined from feature name

    @L2 @ov
    Scenario: Feature name determines output path
      Given the AI generates a feature named "eac-commands_user-auth"
      When I run the command "create spec 'Add user authentication'"
      Then the exit code is 0
      And the file is saved at "specs/eac-commands/user-auth/specification.feature"

    @L2 @ov
    Scenario: Parent directories are created if needed
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec 'New module feature'"
      Then the exit code is 0
      And the parent directories are created if they don't exist

  Rule: Existing files are protected by default

    @L2 @ov
    Scenario: Command refuses to overwrite existing files
      Given a specification file exists at "specs/eac-commands/auth/specification.feature"
      And the mock AI generates a feature that would create the same path
      When I run the command "create spec 'Add authentication'"
      Then the exit code is 1
      And stderr contains "File already exists"
      And stderr contains "--force"

    @L2 @ov
    Scenario: Force flag allows overwriting existing files
      Given a specification file exists at "specs/eac-commands/auth/specification.feature"
      And the mock AI generates a feature that would create the same path
      When I run the command "create spec --force 'Add authentication'"
      Then the exit code is 0
      And the existing file is overwritten

  Rule: Custom output path can be specified

    @L2 @ov
    Scenario: Output flag overrides default path
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec -o 'custom/path/spec.feature' 'Test feature'"
      Then the exit code is 0
      And the file is saved at "custom/path/spec.feature"

  Rule: Debug mode saves intermediate outputs

    @L2 @ov
    Scenario: Debug flag saves intermediate files
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec --debug 'Test debug mode'"
      Then the exit code is 0
      And intermediate files are saved to "out" directory

    @L2 @ov
    Scenario: Debug mode saves full AI prompt
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec -d 'Test prompt capture'"
      Then the exit code is 0
      And "out/debug-full-prompt.md" contains the full AI prompt

  Rule: Custom prompts and templates can be used

    @L2 @ov
    Scenario: Custom prompt file overrides default
      Given a custom prompt file exists at "custom/prompt.md"
      And the mock AI is configured to return a valid specification
      When I run the command "create spec --prompt custom/prompt.md 'Test custom prompt'"
      Then the exit code is 0
      And the custom prompt is used

    @L2 @ov
    Scenario: Custom template file can be specified
      Given a custom template file exists at "custom/template.feature"
      And the mock AI is configured to return a valid specification
      When I run the command "create spec --template custom/template.feature 'Test custom template'"
      Then the exit code is 0
      And the custom template is used

  Rule: Contract-based validation ensures quality

    @L2 @ov
    Scenario: Generated specification is validated against contract
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec 'Test validation'"
      Then the exit code is 0
      And the content is validated

    @L2 @ov
    Scenario: Validation errors prevent file creation
      Given the AI provider fails to generate content
      When I run the command "create spec 'Test failure handling'"
      Then the exit code is 1
      And no specification file is created

  Rule: Security prevents path traversal

    @L2 @ov
    Scenario: Path traversal attempts are rejected
      Given the mock AI is configured to return a valid specification
      When I run the command "create spec -o '../../../etc/passwd' 'Test security'"
      Then the exit code is 1
      And stderr contains "security error"
