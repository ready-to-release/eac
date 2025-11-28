@skip:wip @deps:go @L2 @ov @env:isolated-test-project
Feature: src-commands_risks-create

  As a developer
  I want to create risk control specifications from assessment reports
  So that I can implement controls for identified risks

  Background:
    Given I am in a git repository
    And a risk assessment exists at ".docs/reference/assessment.md"

  @iv
  Rule: Command must be registered and accessible

    Scenario: Command is listed
      When I run "list commands"
      Then the exit code is 0
      And I should see "risks-create"

    Scenario: Command shows help
      When I run "risks create --help"
      Then the exit code is 0
      And stdout contains "Create risk control specifications"
      And stdout contains "--force"
      And stdout contains "-f"
      And stdout contains "--output"
      And stdout contains "-o"

  Rule: Command creates controls from assessment file

    Scenario: Create controls from single assessment
      Given the assessment contains 3 identified risks
      When I run "risks create .docs/reference/assessment.md"
      Then the exit code is 0
      And 3 feature files are created under "specs/risk-controls/"
      And each file is valid Gherkin
      And each file has proper @risk-control tags

    Scenario: Controls grouped by domain
      Given the assessment contains risks in domains "authentication" and "api-security"
      When I run "risks create .docs/reference/assessment.md"
      Then a file exists at "specs/risk-controls/authentication/*.feature"
      And a file exists at "specs/risk-controls/api-security/*.feature"

  Rule: Command parses assessment structure

    Scenario: Extract risk metadata
      Given the assessment has risk "RISK-001" with severity "high"
      When I run "risks create .docs/reference/assessment.md"
      Then the generated control includes severity metadata
      And the control filename includes the risk ID

    Scenario: Map risks to files and specs
      Given risk "RISK-001" affects "src/auth/handler.go"
      And risk "RISK-001" relates to "specs/auth/authentication.feature"
      When I run "risks create .docs/reference/assessment.md"
      Then the generated control documents these relationships

  Rule: Command uses specs create internally

    Scenario: Calls specs create for each risk
      Given the assessment contains 2 risks
      When I run "risks create .docs/reference/assessment.md --debug"
      Then specs create is called 2 times
      And each call includes risk description
      And debug logs show specs create invocations

  Rule: Force flag overwrites existing controls

    Scenario: Skip existing controls by default
      Given a control exists at "specs/risk-controls/auth/mfa.feature"
      When I run "risks create assessment.md"
      Then the existing control is not overwritten
      And stdout contains "Skipped: already exists"

    Scenario: Overwrite with force flag (long form)
      Given a control exists at "specs/risk-controls/auth/mfa.feature"
      When I run "risks create assessment.md --force"
      Then the existing control is overwritten
      And stdout contains "Overwritten"

    Scenario: Overwrite with force flag (short form)
      Given a control exists at "specs/risk-controls/auth/mfa.feature"
      When I run "risks create assessment.md -f"
      Then the existing control is overwritten

  Rule: Force mode prevents orphaned tags

    Scenario: Detect orphaned tags when overwriting
      Given a control exists at "specs/risk-controls/auth/mfa.feature"
      And the control has tag "@risk-control:auth-mfa-01"
      And other specs reference "@risk-control:auth-mfa-01"
      When I run "risks create assessment.md --force"
      Then the command analyzes tag usage across all specs
      And warns if tags would become orphaned
      And preserves referenced tags in new control

    Scenario: Prevent overwrite if orphaned tags detected
      Given a control exists at "specs/risk-controls/auth/mfa.feature"
      And the control has tags "@risk-control:auth-mfa-01" and "@risk-control:auth-mfa-02"
      And other specs reference these tags
      And the new control does not include these tags
      When I run "risks create assessment.md --force"
      Then the exit code is 1
      And stderr contains "would create orphaned tags"
      And stderr lists the orphaned tags
      And stderr suggests adding tags to new control

    Scenario: Safe overwrite when tags preserved
      Given a control exists with tag "@risk-control:auth-mfa-01"
      And other specs reference "@risk-control:auth-mfa-01"
      And the new control includes "@risk-control:auth-mfa-01"
      When I run "risks create assessment.md --force"
      Then the exit code is 0
      And the control is overwritten safely
      And no orphaned tags are created

    Scenario: Allow force overwrite with explicit confirmation
      Given a control would create orphaned tags
      When I run "risks create assessment.md --force --allow-orphans"
      Then the control is overwritten
      And a warning is displayed about orphaned tags

  Rule: Custom output directory supported

    Scenario: Use default output directory
      When I run "risks create assessment.md"
      Then controls are created under "specs/risk-controls/"

    Scenario: Use custom output directory (long form)
      When I run "risks create assessment.md --output specs/custom/"
      Then controls are created under "specs/custom/"

    Scenario: Use custom output directory (short form)
      When I run "risks create assessment.md -o specs/custom/"
      Then controls are created under "specs/custom/"

  Rule: All flags have shorthand versions

    Scenario: Short flags work
      When I run "risks create assessment.md -f -o specs/custom/ -p prompt.md -D"
      Then all flags are processed correctly
      And debug mode is enabled
      And custom prompt is used
      And output goes to "specs/custom/"
      And force mode is enabled

  Rule: Process multiple assessment files

    Scenario: Create from folder of assessments
      Given a folder ".docs/assessments/" with 3 assessment files
      When I run "risks create .docs/assessments/"
      Then all assessments are processed
      And controls are created for all identified risks
      And a summary report is displayed

  Rule: Error handling

    Scenario: Assessment file not found
      When I run "risks create nonexistent.md"
      Then the exit code is 1
      And stderr contains "assessment file not found"

    Scenario: Invalid assessment format
      Given an invalid assessment file "bad.md"
      When I run "risks create bad.md"
      Then the exit code is 1
      And stderr contains "failed to parse assessment"

    Scenario: specs create fails
      Given the assessment contains invalid risk data
      When I run "risks create assessment.md"
      Then the exit code is 1
      And stderr contains "failed to create control"
