@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_create-risk

  As a security engineer
  I want to create OSCAL profiles from risk assessment documents
  So that I can define security controls for my modules

  Background:
    Given I am in a git repository
    And AI provider is configured

  Rule: Command must be registered and accessible

    @L3 @iv
    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "create risk"

    @L3 @iv
    Scenario: Command shows help
      When I run the command "show help create risk"
      Then the exit code is 0
      And stdout contains "OSCAL profile"
      And stdout contains "--module"
      And stdout contains "--catalog"
      And stdout contains "--output"
      And stdout contains "--force"
      And stdout contains "--debug"

  Rule: Command requires assessment file argument

    @L2 @ov
    Scenario: Missing assessment file shows error
      When I run "create risk"
      Then the exit code is 1
      And stderr contains "assessment file path required"

    @L2 @ov
    Scenario: Non-existent file shows error
      When I run "create risk nonexistent.md"
      Then the exit code is 1
      And stderr contains "assessment file not found"

  Rule: Command generates OSCAL profile from assessment

    @L2 @ov
    Scenario: Create profile from assessment document
      Given a risk assessment file at ".docs/assessments/security-assessment.md"
      And the assessment identifies risks requiring controls "ac-2" and "ia-2"
      When I run "create risk .docs/assessments/security-assessment.md --module billing"
      Then the exit code is 0
      And a file exists at "specs/risk-controls/billing.profile.json"
      And the profile contains valid OSCAL 1.1.2 JSON
      And the profile imports NIST 800-53 catalog
      And the profile includes controls "ac-2" and "ia-2"

    @L2 @ov
    Scenario: Default module name is 'common'
      Given a risk assessment file at "assessment.md"
      When I run "create risk assessment.md"
      Then the exit code is 0
      And a file exists at "specs/risk-controls/common.profile.json"

  Rule: Command validates generated profile

    @L2 @ov
    Scenario: Profile is validated after generation
      Given a risk assessment file at "assessment.md"
      When I run "create risk assessment.md --module test"
      Then the exit code is 0
      And stdout contains "Profile created"
      And the generated profile passes OSCAL validation

    @L2 @ov
    Scenario: AI retry on validation failure
      Given a risk assessment that causes invalid output
      When I run "create risk assessment.md --max-retries 2"
      Then the AI is called up to 2 times
      And validation errors are fed back to AI

  Rule: Force flag controls overwrite behavior

    @L2 @ov
    Scenario: Existing profile not overwritten by default
      Given a profile exists at "specs/risk-controls/billing.profile.json"
      When I run "create risk assessment.md --module billing"
      Then the exit code is 1
      And stderr contains "already exists"
      And the existing profile is not modified

    @L2 @ov
    Scenario: Force flag overwrites existing profile
      Given a profile exists at "specs/risk-controls/billing.profile.json"
      When I run "create risk assessment.md --module billing --force"
      Then the exit code is 0
      And the profile is overwritten

    @L2 @ov
    Scenario: Short force flag works
      Given a profile exists at "specs/risk-controls/billing.profile.json"
      When I run "create risk assessment.md --module billing -f"
      Then the exit code is 0
      And the profile is overwritten

  Rule: Custom catalog URL supported

    @L2 @ov
    Scenario: Use default NIST 800-53 catalog
      Given a risk assessment file at "assessment.md"
      When I run "create risk assessment.md"
      Then the generated profile references NIST 800-53 Rev 5 catalog

    @L2 @ov
    Scenario: Use custom catalog URL
      Given a risk assessment file at "assessment.md"
      When I run "create risk assessment.md --catalog https://example.com/catalog.json"
      Then the generated profile references "https://example.com/catalog.json"

  Rule: Custom output path supported

    @L2 @ov
    Scenario: Use default output path
      When I run "create risk assessment.md --module billing"
      Then the profile is created at "specs/risk-controls/billing.profile.json"

    @L2 @ov
    Scenario: Use custom output path
      When I run "create risk assessment.md --output custom/path/my.profile.json"
      Then the profile is created at "custom/path/my.profile.json"

  Rule: Debug mode saves intermediate outputs

    @L2 @ov
    Scenario: Debug flag saves prompts and responses
      Given a risk assessment file at "assessment.md"
      When I run "create risk assessment.md --debug"
      Then the exit code is 0
      And debug output is saved to "out/logs/risk/"
      And the debug output includes the AI prompt
      And the debug output includes the AI response

  Rule: AI prompts can be customized

    @L2 @ov
    Scenario: Load prompt from .r2r/eac/ai/risk-create/profile.md
      Given a custom prompt exists at ".r2r/eac/ai/risk-create/profile.md"
      When I run "create risk assessment.md"
      Then the custom prompt is used for AI generation

  Rule: Error handling is robust

    @L2 @ov
    Scenario: AI provider not configured
      Given AI provider is not configured
      When I run "create risk assessment.md"
      Then the exit code is 1
      And stderr contains "AI" or "provider"

    @L2 @ov
    Scenario: AI call fails
      Given AI provider returns an error
      When I run "create risk assessment.md"
      Then the exit code is 1
      And stderr contains "AI" or "failed"
