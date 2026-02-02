@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-cli_books_validation_edge_cases

  As a documentation maintainer
  I want validation to catch configuration errors
  So that I can fix problems before building

  Rule: Validation must detect configuration errors

    Scenario: Show books shows default books when none configured
      Given books.yml does not exist
      When I run the command "show books"
      Then the exit code is 0
      And I should see "Name" or "docs"

    Scenario: Validate detects missing module
      Given books.yml references module "nonexistent"
      When I run the command "validate books"
      Then the exit code is 1
      And I should see "Unknown command" or "not found"

    Scenario: Validate detects invalid command
      Given books.yml has inline source with command "nonexistent-command"
      When I run the command "validate books"
      Then the exit code is 1
      And I should see "Unknown inline command" or "unknown command"
