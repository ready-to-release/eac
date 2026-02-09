@deps:go @L2 @ov
Feature: eac-cli_books

  As a documentation maintainer
  I want to build books that aggregate static and generated content
  So that I can produce comprehensive documentation with up-to-date reference material

  # ============================================================================
  # Configuration Commands - Run against real repository
  # ============================================================================

  Rule: Books configuration must be discoverable

    Scenario: Show configured books
      When I run the command "show books"
      Then the exit code is 0
      And I should see "reference" in the output
      And I should see a table with book information

  Rule: Books configuration must be validated

    Scenario: Validate valid books.yml
      When I run the command "validate books"
      Then the exit code is 0
      And I should see "valid" or "success"
