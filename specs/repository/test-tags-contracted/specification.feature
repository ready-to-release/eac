# Intent: Ensure tag filtering and test selection works correctly by requiring all test tags to be defined in the tag contract
# Architecture: Affects repository tag validation; invokes validate test-tags command; depends on tag contract definition and all test files across the repository

@deps:go @control:sa-3
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
