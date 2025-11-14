@ov
Feature: src-cli_verify-configuration

  As a developer
  I want to verify my CLI is configured correctly
  So that I can ensure the tool is ready to use

  Background:
    Given I create a test folder "test-verify-cli"
    And I create a ".git" folder in the test folder
    And I build the CLI with "build module src-cli"
    And the build succeeds
    And I change directory to the test folder

  Rule: Verify command detects missing configuration

    Scenario: Reports error when config file does not exist
      Given no config file exists in the test folder
      When I run the built CLI with "verify"
      Then the exit code is 1
      And I should see "config" or "not found" or "Error"

  Rule: Verify command validates configuration file

    @skip:wip # WIP: This test expects default extension to exist upstream. Currently no extensions configured causes warning and exit 1. Test will pass when default extension is available.
    Scenario: Successfully verifies with valid config file
      Given I create a test config file ".r2r.yaml" with valid settings
      When I run the built CLI with "verify"
      Then the exit code is 0
      And I should see "✓" or "success" or "Verifying"

    Scenario: Reports error with invalid config file
      Given I create a test config file ".r2r.yaml" with invalid settings
      When I run the built CLI with "verify"
      Then the exit code is 1
      And I should see "Error" or "invalid" or "failed"

  Rule: Built CLI executable must be functional

    Scenario: Built CLI shows version
      When I run the built CLI with "--version"
      Then the exit code is 0
      And I should see version number

    Scenario: Built CLI shows help
      When I run the built CLI with "--help"
      Then the exit code is 0
      And I should see "verify" or "Usage"
