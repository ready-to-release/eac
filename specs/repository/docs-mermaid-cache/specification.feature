@L1
Feature: repository_docs-mermaid-cache

  As a repository maintainer
  I want to ensure all mermaid diagrams in docs/ are cached
  So that CI builds are fast and diagrams render consistently

  Background:
    Given the repository root exists

  Rule: All mermaid diagrams must have cached SVG files

    @L1 @ov
    Scenario: All mermaid diagrams in docs/ have cached SVGs
      Given I scan docs/ for all mermaid code blocks
      When I check each mermaid block against the cache
      Then all mermaid diagrams should have cached SVGs
      And if any diagrams are missing from cache, I should see:
        """
        Run 'clie update docs' to update the mermaid cache.
        """
