@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-commands_risk-assess

  As a security engineer
  I want to assess modules against their risk profiles
  So that I can track control satisfaction status with evidence

  Background:
    Given I am in a git repository
    And module "billing" exists with a profile at "specs/risk-controls/billing.profile.json"

  Rule: Command must be registered and accessible

    @L3 @iv
    Scenario: Command is listed in available commands
      When I run the command "show help"
      Then the exit code is 0
      And I should see "risk assess"

    @L3 @iv
    Scenario: Command shows help
      When I run the command "show help risk assess"
      Then the exit code is 0
      And stdout contains "assessment-results"
      And stdout contains "--max-evidence-age"
      And stdout contains "--force-tests"
      And stdout contains "--force-security"
      And stdout contains "--skip-auto-run"

  Rule: Command requires module argument

    @L2 @ov
    Scenario: Missing module shows error
      When I run "risk assess"
      Then the exit code is 1
      And stderr contains "module name required"

    @L2 @ov
    Scenario: Non-existent module shows error
      When I run "risk assess nonexistent-module"
      Then the exit code is 1
      And stderr contains "not found" or "no profile"

  Rule: Command creates OSCAL assessment-results

    @L2 @ov
    Scenario: Create assessment-results for module
      Given module "billing" has test results with @control tags
      And module "billing" has security scan results
      When I run "risk assess billing"
      Then the exit code is 0
      And a file exists at "out/risk/billing/assessment-results.json"
      And the assessment-results contains valid OSCAL 1.1.2 JSON
      And the assessment-results references the profile

    @L2 @ov
    Scenario: Assessment includes findings for each control
      Given the profile has controls "ac-2" and "ia-2"
      And tests exist with "@control(ac-2)" tag that pass
      And no tests exist for "ia-2"
      When I run "risk assess billing"
      Then the assessment-results has finding for "ac-2" with status "satisfied"
      And the assessment-results has finding for "ia-2" with status "not-satisfied"

  Rule: Command collects test evidence

    @L2 @ov
    Scenario: Extract @control tags from test results
      Given cucumber results exist at "out/test/*/billing/*.cucumber.json"
      And tests have "@control(ac-2)" tags
      When I run "risk assess billing"
      Then the assessment-results includes test evidence
      And control "ac-2" is linked to test results

    @L2 @ov
    Scenario: Calculate test coverage per control
      Given tests with "@control(ac-2)" have 3 passing and 1 failing scenarios
      When I run "risk assess billing"
      Then the finding for "ac-2" includes coverage metrics
      And the coverage shows "3/4 scenarios passing"

  Rule: Command collects security evidence

    @L2 @ov
    Scenario: Include security scan results
      Given vulnerability scan results exist at "out/security/billing/vuln/*.json"
      When I run "risk assess billing"
      Then the assessment-results includes security observations
      And vulnerabilities are linked to relevant controls

    @L2 @ov
    Scenario: Multiple security scan types supported
      Given SBOM results exist at "out/security/billing/sbom/*.json"
      And SAST results exist at "out/security/billing/sast/*.json"
      When I run "risk assess billing"
      Then the assessment-results includes SBOM observations
      And the assessment-results includes SAST observations

  Rule: Auto-run tests and scans when evidence is stale

    @L2 @ov
    Scenario: Auto-run tests when evidence older than 24h
      Given test results are 25 hours old
      When I run "risk assess billing"
      Then tests are automatically executed
      And fresh evidence is collected

    @L2 @ov
    Scenario: Auto-run security scans when evidence older than 24h
      Given security scan results are 25 hours old
      When I run "risk assess billing"
      Then security scans are automatically executed
      And fresh evidence is collected

    @L2 @ov
    Scenario: Custom evidence age threshold
      Given test results are 2 hours old
      When I run "risk assess billing --max-evidence-age 1h"
      Then tests are automatically executed

    @L2 @ov
    Scenario: Skip auto-run with flag
      Given test results are 25 hours old
      When I run "risk assess billing --skip-auto-run"
      Then tests are not executed
      And existing evidence is used

  Rule: Force flags override evidence age check

    @L2 @ov
    Scenario: Force tests regardless of age
      Given test results are 1 hour old
      When I run "risk assess billing --force-tests"
      Then tests are executed despite fresh evidence

    @L2 @ov
    Scenario: Force security scans regardless of age
      Given security scan results are 1 hour old
      When I run "risk assess billing --force-security"
      Then security scans are executed despite fresh evidence

  Rule: Assessment-results updates existing file

    @L2 @ov
    Scenario: Update existing assessment-results
      Given assessment-results exists at "out/risk/billing/assessment-results.json"
      When I run "risk assess billing"
      Then the assessment-results is updated
      And the UUID is preserved
      And the last-modified timestamp is updated

    @L2 @ov
    Scenario: Preserve historical observations
      Given assessment-results has existing observations
      When I run "risk assess billing"
      Then new observations are added
      And existing observations are preserved

  Rule: Debug mode provides detailed output

    @L2 @ov
    Scenario: Debug flag shows evidence collection details
      When I run "risk assess billing --debug"
      Then stdout shows test evidence discovery
      And stdout shows security evidence discovery
      And stdout shows control mapping details

  Rule: Error handling is robust

    @L2 @ov
    Scenario: No profile found for module
      Given module "new-module" has no profile
      When I run "risk assess new-module"
      Then the exit code is 1
      And stderr contains "no profile found"

    @L2 @ov
    Scenario: Test execution fails
      Given test execution will fail
      And --skip-auto-run is not set
      When I run "risk assess billing"
      Then the exit code is 1
      And stderr contains "test execution failed"

    @L2 @ov
    Scenario: Missing evidence with skip-auto-run
      Given no test results exist
      And no security scan results exist
      When I run "risk assess billing --skip-auto-run"
      Then the exit code is 0
      And all controls are marked "not-satisfied"
      And a warning is displayed about missing evidence
