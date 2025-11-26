@skip:todo @deps:ai @deps:go @env:isolated-test-project @ov
Feature: src-commands_specs-create

  As a developer of the eac platform
  I want AI-powered specification generation
  So that I can create Gherkin feature files from natural language descriptions

  Background:
    Given I am in a git repository

  @iv
  Rule: Command must be registered and accessible

    @iv
    Scenario: Command is listed in available commands
      When I run the command "list commands"
      Then the exit code is 0
      And I should see "specs create"

    @iv
    Scenario: Command has proper description
      When I run the command "describe commands specs create"
      Then the exit code is 0
      And I should see "specs" or "specification" or "create"

  Rule: Command validates input

    @ov
    Scenario: Command requires description argument
      When I run "specs create" without arguments
      Then the command exits with code 1
      And stderr contains "description is required"

    @ov
    Scenario: Long descriptions are truncated with warning
      Given a description longer than 1000 characters
      And the mock AI is configured to return a valid specification
      When I run the specs create command
      Then stderr contains "truncated"

  Rule: AI generates valid Gherkin specifications

    @ov
    Scenario: AI generates specification from description
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create 'Add user authentication with email and password'"
      Then the exit code is 0
      And stdout contains "Specification created"
      And a specification file is created

    @ov
    Scenario: AI output noise is filtered
      Given the AI provider returns output with initialization messages
      When the output is processed
      Then initialization noise should be removed
      And only valid Gherkin content should remain

    @ov
    Scenario: Generated specification must contain required Gherkin elements
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create 'Test feature generation'"
      Then the exit code is 0
      And it must contain a "Feature:" declaration
      And it must contain at least one "Rule:" declaration
      And it must contain at least one "Scenario:" declaration

  Rule: Output path is determined from feature name

    @ov
    Scenario: Feature name determines output path
      Given the AI generates a feature named "src-commands_user-auth"
      When I run the command "specs create 'Add user authentication'"
      Then the exit code is 0
      And the file is saved at "specs/src-commands/user-auth/specification.feature"

    @ov
    Scenario: Parent directories are created if needed
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create 'New module feature'"
      Then the exit code is 0
      And the parent directories are created if they don't exist

  Rule: Existing files are protected by default

    @ov
    Scenario: Command refuses to overwrite existing files
      Given a specification file exists at "specs/src-commands/auth/specification.feature"
      And the mock AI generates a feature that would create the same path
      When I run the command "specs create 'Add authentication'"
      Then the exit code is 1
      And stderr contains "File already exists"
      And stderr contains "--force"

    @ov
    Scenario: Force flag allows overwriting existing files
      Given a specification file exists at "specs/src-commands/auth/specification.feature"
      And the mock AI generates a feature that would create the same path
      When I run the command "specs create --force 'Add authentication'"
      Then the exit code is 0
      And the existing file is overwritten

  Rule: Custom output path can be specified

    @ov
    Scenario: Output flag overrides default path
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create -o 'custom/path/spec.feature' 'Test feature'"
      Then the exit code is 0
      And the file is saved at "custom/path/spec.feature"

    @ov
    Scenario: Module flag sets target module
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create -m src-core 'Add validation helper'"
      Then the exit code is 0
      And the AI receives module context "src-core"

  Rule: Debug mode saves intermediate outputs

    @ov
    Scenario: Debug flag saves intermediate files
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create --debug 'Test debug mode'"
      Then the exit code is 0
      And intermediate files are saved to "out" directory

    @ov
    Scenario: Debug mode saves full AI prompt
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create -d 'Test prompt capture'"
      Then the exit code is 0
      And "out/debug-full-prompt.md" contains the full AI prompt

  Rule: Custom prompts and templates can be used

    @ov
    Scenario: Custom prompt file overrides default
      Given a custom prompt file exists at "custom/prompt.md"
      And the mock AI is configured to return a valid specification
      When I run the command "specs create --prompt custom/prompt.md 'Test custom prompt'"
      Then the exit code is 0
      And the custom prompt is used

    @ov
    Scenario: Custom template file can be specified
      Given a custom template file exists at "custom/template.feature"
      And the mock AI is configured to return a valid specification
      When I run the command "specs create --template custom/template.feature 'Test custom template'"
      Then the exit code is 0
      And the custom template is used

  Rule: Contract-based validation ensures quality

    @ov
    Scenario: Generated specification is validated against contract
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create 'Test validation'"
      Then the exit code is 0
      And the content is validated

    @ov
    Scenario: Validation errors prevent file creation
      Given the AI provider fails to generate content
      When I run the command "specs create 'Test failure handling'"
      Then the exit code is 1
      And no specification file is created

  Rule: Security prevents path traversal

    @ov
    Scenario: Path traversal attempts are rejected
      Given the mock AI is configured to return a valid specification
      When I run the command "specs create -o '../../../etc/passwd' 'Test security'"
      Then the exit code is 1
      And stderr contains "security error"
