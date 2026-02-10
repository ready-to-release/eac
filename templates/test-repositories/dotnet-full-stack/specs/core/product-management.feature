Feature: Product Management
  As a product manager
  I want to manage products
  So that I can maintain the catalog

  Scenario: Add a new product
    Given an empty product catalog
    When I add a product "Widget" priced at 9.99
    Then the catalog should contain 1 product
