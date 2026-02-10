@L2 @deps:python @depm:report-cli
Feature: Report Generation
  As a user of the report CLI
  I want to generate formatted reports
  So that I can present data clearly

  Rule: Basic report formatting

    Scenario: Generate a summary report
      Given a dataset with values 10, 20, 30
      When I generate a summary report
      Then the report should contain the mean value 20.0
      And the report should contain the median value 20

    Scenario: Handle empty dataset
      Given an empty dataset
      When I generate a summary report
      Then the report should show an error message
