@deps:go @verification @unisolated
Feature: repository_go-modules-tidy

  As a repository maintainer
  I want to ensure all Go modules have tidy dependencies
  So that builds are reproducible and dependencies are consistent

  Background:
    Given the repository root exists

  Rule: All Go modules must be tidy

    @L2 @ov @skip:broken
    Scenario: All Go modules in repository are tidy
      Given I discover all Go modules in the repository using module contracts
      When I run "go mod tidy -diff" in each Go module directory
      Then all modules should have exit code 0
      And no module should have any diff output
      And if any module is not tidy, I should see the module path and diff
