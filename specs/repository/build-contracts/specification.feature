@deps:go
Feature: repository_build-contracts

  As a repository maintainer
  I want to ensure the contracts module builds successfully
  So that module contract definitions are valid and can be used by the system

  Background:
    Given the repository root exists

  Rule: The contracts module must build without errors

    @L2 @ov
    Scenario: contracts module builds successfully
      When I run the command "build contracts"
      Then the exit code is 0
      And I should not see any build errors
