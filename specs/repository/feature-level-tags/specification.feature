@L2 @verification @unisolated
Feature: repository_feature-level-tags

  As a repository maintainer
  I want to ensure Feature-level tags don't conflict with Scenario tags
  So that Gherkin tag inheritance doesn't cause tests to have multiple tags that violate validation rules

  Background:
    Given the repository root exists

  Rule: Features and Scenarios cannot have conflicting taxonomy tags

    Tag types with "exactly one" or "no duplicates" constraints must not conflict between
    Feature and Scenario levels. Currently validated:
    - L-level tags (@L0-@L4) - must have exactly one
    - Verification tags (@ov/@iv/@pv/@piv/@ppv) - validation checks for multiple

    @L2 @ov
    Scenario: No Feature has conflicting taxonomy tags with its Scenarios
      Given I discover all Gherkin feature files in the repository
      When I validate each feature file for conflicting L-level tags
      Then no feature should have an L-tag when its scenarios have different L-tags
      And no feature should have a verification tag when its scenarios have different verification tags
      And if any conflicts are found, I should see the file path, scenario name, and conflicting tags
