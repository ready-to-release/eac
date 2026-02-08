@L1
Feature: repository_docs-structurizr-cache

  As a repository maintainer
  I want to ensure all Structurizr diagrams are built
  So that CI builds are fast and architecture diagrams render consistently

  Background:
    Given the repository root exists

  Rule: All Structurizr views must have built SVG files

    @L1 @ov
    Scenario: All workspace.dsl views have built SVGs
      Given I scan specs/ for all workspace.dsl files
      When I check each Structurizr view against the cache
      Then all Structurizr views should have cached SVGs
      And if any views are missing from cache, I should see:
        """
        Run 'clie eac build --module docs' to build the Structurizr diagrams.
        """
