@L3 @skip:wip
Feature: github_release-workflow

As a developer
I want to automatically build and release multi-platform Go binaries when tagging releases
So that users can download pre-built binaries for their platform

Background:
  Given the repository is initialized with git
  And the Go toolchain is available
  And the src-cli module is built

Rule: Workflow triggers on release tags matching pattern

  @skip:wip @L3 @iv @deps:git @depm:src-cli
  Scenario: Workflow triggers when tag matches src-cli pattern
    Given a git repository with src-cli source code
    When a git tag matching pattern "src-cli/*" is created
    And the tag is pushed to origin
    Then the GitHub Actions workflow is triggered
    And the workflow job completes successfully

Rule: Go binary is built for multiple platforms

  @skip:wip @L3 @iv @deps:go @depm:src-cli
  Scenario: Binaries compiled for all target platforms
    Given the workflow environment is configured
    And the Go source code is available
    When the build process runs for all platforms
    Then binaries are created for linux/amd64, darwin/amd64, darwin/arm64, and windows/amd64
    And all binaries are executable

Rule: GitHub release is created with built artifacts

  @skip:wip @L3 @ov @depm:src-cli
  Scenario: GitHub release created with all platform binaries attached
    Given the workflow has built binaries for all platforms
    And the version "1.2.3" is extracted from the tag
    When the workflow creates a GitHub release
    Then a release is created with tag "src-cli/1.2.3"
    And all platform binaries are attached to the release
    And the release includes installation instructions