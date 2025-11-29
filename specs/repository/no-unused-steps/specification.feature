Feature: repository_no-unused-steps

  As a repository maintainer
  I want to ensure no step definitions are left unused
  So that the test codebase remains clean and maintainable

  Background:
    Given the repository root exists

  Rule: All step definitions must be used by at least one feature file

    @L1 @ov
    Scenario: No unused step definitions in repository
      When I run the command "specs unused-steps"
      Then the command should succeed
      And the output should contain "No unused step definitions found"
