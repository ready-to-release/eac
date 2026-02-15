@deps:go @env:isolated-test-project
Feature: eac-cli_release-this

  As a developer
  I want the release process to manage changelog updates
  So that releases are properly documented

  Background:
    Given I am in a git repository
    And the repository has the EAC module system configured

  Rule: Dry run skips validation

    @L2 @ov
    Scenario: Dry run does not validate release notes
      Given a module "test-module" exists
      And the module has commits since last release
      And "release/test-module/RELEASE-NOTES.md" does not exist
      When I run "release this test-module --dry-run"
      Then the command should succeed
      And "release/test-module/RELEASE-NOTES.md" should not exist

  Rule: Each module has independent release notes

    @L2 @ov
    Scenario: Each module has independent release notes
      Given module "eac-ext" with releases
      And module "eac" with releases
      When I check release notes locations
      Then "release/eac-ext/RELEASE-NOTES.md" should be for eac-ext
      And "release/eac/RELEASE-NOTES.md" should be for eac
      And there should be no cross-module contamination
