@deps:go @env:isolated-test-project @skip:wip
Feature: eac-commands_release-this

  As a developer
  I want the release process to validate release notes
  So that all releases have proper documentation

  Background:
    Given I am in a git repository
    And the repository has the EAC module system configured

  Rule: Release notes file must exist before release

    @L2 @ov
    Scenario: Release fails when RELEASE-NOTES.md is missing
      Given a module "test-module" exists
      And the module has commits since last release
      And "release/test-module/RELEASE-NOTES.md" does not exist
      When I run "release this test-module"
      Then the command should create "release/test-module/RELEASE-NOTES.md"
      And the command should fail
      And the output should contain "RELEASE-NOTES.md created"
      And the output should contain "Edit the file and fill in required sections"
      And the output should contain "Run 'release this test-module' again"

    @L2 @ov
    Scenario: Generated template has correct structure
      Given a module "test-module" exists
      And the module has commits since last release
      And "release/test-module/RELEASE-NOTES.md" does not exist
      And the calculated next version is "0.0.5"
      When I run "release this test-module"
      Then "release/test-module/RELEASE-NOTES.md" should exist
      And the file should contain "# Release release_notes"
      And the file should contain "## [0.0.5] - "
      And the file should contain "### Conclusion on Fitness for Intended Use"
      And the file should contain "### Impact on Business Process"
      And the file should contain "[TODO: Provide conclusion on fitness for intended use]"

  Rule: Release notes must contain version entry with content

    @L2 @ov
    Scenario: Release fails when version entry is missing
      Given a module "test-module" exists
      And the module has commits since last release
      And "release/test-module/RELEASE-NOTES.md" exists with version "0.0.1"
      And the calculated next version is "0.0.2"
      When I run "release this test-module"
      Then the command should fail
      And the output should contain "RELEASE-NOTES.md validation failed"
      And the output should contain "version [0.0.2] not found"
      And the output should contain "Please add version [0.0.2]"

    @L2 @ov
    Scenario: Release fails when version entry has no content
      Given a module "test-module" exists
      And the module has commits since last release
      And "release/test-module/RELEASE-NOTES.md" exists with empty version "0.0.2"
      And the calculated next version is "0.0.2"
      When I run "release this test-module"
      Then the command should fail
      And the output should contain "RELEASE-NOTES.md validation failed"
      And the output should contain "all sections are empty"

  Rule: Release succeeds with valid release notes

    @L2 @ov
    Scenario: Release succeeds when version entry exists with content
      Given a module "test-module" exists
      And the module has commits since last release
      And "release/test-module/RELEASE-NOTES.md" exists with version "0.0.2" and content
      And the calculated next version is "0.0.2"
      When I run "release this test-module"
      Then the command should succeed
      And "release/test-module/CHANGELOG.md" should be updated
      And the CHANGELOG should contain version "0.0.2"

    @L2 @ov
    Scenario: Release works for different modules independently
      Given module "module-a" has commits since last release
      And "release/module-a/RELEASE-NOTES.md" exists with version "1.0.0" and content
      And module "module-b" has commits since last release
      And "release/module-b/RELEASE-NOTES.md" exists with version "2.0.0" and content
      When I run "release this module-a"
      Then the command should succeed
      And "release/module-a/CHANGELOG.md" should be updated
      And "release/module-b/CHANGELOG.md" should not be modified

  Rule: Dry run skips release notes validation

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
      Given module "ext-eac" with releases
      And module "eac-commands" with releases
      When I check release notes locations
      Then "release/ext-eac/RELEASE-NOTES.md" should be for ext-eac
      And "release/eac-commands/RELEASE-NOTES.md" should be for eac-commands
      And there should be no cross-module contamination
