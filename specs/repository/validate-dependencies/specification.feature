@deps:go @control:si-7 @control:sa-10
Feature: repository_validate-dependencies

  As a repository maintainer
  I want to ensure all module dependencies are valid
  So that module dependency contracts are consistent and enforceable

  Background:
    Given the repository root exists

  Rule: All module dependencies must be valid

    @L1 @ov
    Scenario: All module dependencies in repository are valid
      When I run the command "validate dependencies"
      Then the exit code is 0
      And I should not see any dependency validation errors
