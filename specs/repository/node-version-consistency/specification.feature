@deps:npm @control:cm-2
Feature: repository_node-version-consistency

  As a repository maintainer
  I want to ensure Node.js version is consistent across all configuration files
  So that CI and local environments use the same Node.js version and avoid npm compatibility issues

  Background:
    Given the repository root exists

  Rule: Node.js version must be consistent across configuration layers

    @L1 @ov
    Scenario: GitHub Actions use the same Node.js version as system-dependencies.yml
      Given I load the system dependencies configuration
      And I discover all GitHub Action workflow files
      When I extract the Node.js version from system-dependencies.yml
      And I extract node-version defaults from all GitHub Actions
      Then all GitHub Action node-version defaults should match system-dependencies.yml
      And if versions differ, I should see the mismatched actions and versions

    @L1 @ov
    Scenario: Known GitHub Actions have correct Node.js version defaults
      Given I load the system dependencies configuration
      When I extract the Node.js version from system-dependencies.yml
      Then .github/actions/setup-module-deps/action.yaml should have matching node-version default
