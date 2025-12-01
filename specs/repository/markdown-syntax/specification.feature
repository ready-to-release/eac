@L1
Feature: repository_markdown-syntax

  As a repository maintainer
  I want to ensure all Markdown files have valid syntax
  So that documentation renders correctly and is readable

  Background:
    Given the repository root exists

  Rule: All Markdown files must have valid syntax

    @L1 @ov
    Scenario: All Markdown files in repository have valid syntax
      Given I discover all Markdown files in the repository
      When I validate each Markdown file for syntax errors
      Then all files should have valid Markdown syntax
      And no files should have broken links
      And no files should have malformed headers
      And if any file has errors, I should see the file path and error details
