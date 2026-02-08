@L1
Feature: repository_docs-drawio-cache

  As a repository maintainer
  I want to ensure all drawio images in docs/ are cached (optimized)
  So that CI builds are fast and PDF generation uses optimized images

  Background:
    Given the repository root exists

  Rule: All drawio images must have cached optimized PNG files

    @L1 @ov
    Scenario: All drawio images in docs/ have cached optimized PNGs
      Given I scan docs/ for all drawio.png images
      When I check each drawio image against the cache
      Then all drawio images should have cached optimized PNGs
      And if any images are missing from cache, I should see:
        """
        Run 'clie update docs' to update the drawio cache.
        """
