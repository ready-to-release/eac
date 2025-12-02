@L3 @iv
Feature: r2r-cli_release-command
  As a release engineer
  I want to create versioned git tags for modules
  So that I can mark stable release points in the repository

  Background:
    Given a git repository with r2r-cli module
    And the release command is available

  Rule: Module validation ensures only valid modules can be released

    @skip:wip
    Scenario: Release command rejects non-existent module
      Given the r2r-cli module exists in the repository
      When I run the release command with module name "invalid-module"
      Then the command fails with error "Module 'invalid-module' not found"
      And no git tag is created

  Rule: Semver validation ensures correct version format

    @skip:wip
    Scenario: Release command rejects invalid semver format
      Given the r2r-cli module exists in the repository
      When I run the release command with module "r2r-cli" and version "1.0"
      Then the command fails with error "Invalid semver format. Use x.y.z"
      And no git tag is created

  Rule: Git tag creation follows module/version naming convention

    @skip:wip
    Scenario: Release command creates and pushes git tag with correct format
      Given the r2r-cli module exists in the repository
      And no tag "r2r-cli/1.0.0" exists
      When I run the release command with module "r2r-cli" and version "1.0.0"
      Then a git tag "r2r-cli/1.0.0" is created
      And the tag is pushed to the remote repository
      And the command displays "Release r2r-cli/1.0.0 created successfully"