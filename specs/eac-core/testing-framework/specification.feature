@L1 @ov @depm:eac-core @skip:broken
Feature: eac-core_testing-framework

  As a developer
  I want the testing framework to work correctly
  So that test discovery, inference, and validation work as expected

  Background:
    Given the testing framework meta tests exist in testdata

  Rule: Meta tests for testing infrastructure must pass in isolation

    Scenario: Test discovery logic works correctly
      When I run the discovery meta tests in isolation
      Then all discovery tests should pass

    Scenario: Test inference logic works correctly
      When I run the inference meta tests in isolation
      Then all inference tests should pass

    Scenario: Test suite selection logic works correctly
      When I run the suite meta tests in isolation
      Then all suite tests should pass

    Scenario: Test validation logic works correctly
      When I run the validation meta tests in isolation
      Then all validation tests should pass

    Scenario: Test reporting logic works correctly
      When I run the reports meta tests in isolation
      Then all reports tests should pass

  Rule: All meta tests must pass together

    Scenario: All testing framework meta tests pass
      When I run all testing framework meta tests in isolation
      Then the complete meta test suite should pass
      And there should be no test failures
