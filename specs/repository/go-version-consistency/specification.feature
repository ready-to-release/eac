# Intent: Ensure CI and local environments use the same Go version to avoid module resolution differences
# Architecture: Affects repository Go version consistency checks; reads system-dependencies.yml, go.work, and GitHub Actions workflow files; validates version alignment across all configuration layers

@deps:go @control:cm-2
Feature: repository_go-version-consistency

  As a repository maintainer
  I want to ensure Go version is consistent across all configuration files
  So that CI and local environments use the same Go version and avoid module resolution differences

  Background:
    Given the repository root exists

  Rule: Go version must be consistent across configuration layers

    @L1 @ov @skip:broken
    Scenario: Go version matches between system-dependencies.yml and go.work
      Given I load the system dependencies configuration
      And I load the go.work file
      When I extract the Go version from system-dependencies.yml
      And I extract the Go version from go.work
      Then the Go versions should match exactly
      And if versions differ, I should see both versions and their locations

    @L1 @ov @skip:broken
    Scenario: GitHub Actions use the same Go version as system-dependencies.yml
      Given I load the system dependencies configuration
      And I discover all GitHub Action workflow files
      When I extract the Go version from system-dependencies.yml
      And I extract go-version defaults from all GitHub Actions
      Then all GitHub Action go-version defaults should match system-dependencies.yml
      And if versions differ, I should see the mismatched actions and versions

    @L1 @ov @skip:broken
    Scenario: All known GitHub Actions have correct Go version defaults
      Given I load the system dependencies configuration
      When I extract the Go version from system-dependencies.yml
      Then .github/actions/setup-commands/action.yaml should have matching go-version default
      And .github/actions/setup-module-deps/action.yaml should have matching go-version default
