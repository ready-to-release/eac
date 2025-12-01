@deps:go
Feature: repository_test-tags-contracted

  As a repository maintainer
  I want to ensure all test tags are defined in the tag contract
  So that tag filtering and test selection works correctly

  Background:
    Given the repository root exists

  Rule: All tests must use only contracted tags

    @L1 @ov
    Scenario: All test tags are defined in the tag contract
      When I run the command "validate test-tags"
      Then the exit code is 0
      And I should not see any undefined tag errors
