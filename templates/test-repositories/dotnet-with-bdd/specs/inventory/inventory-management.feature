Feature: Inventory Management
  As an inventory manager
  I want to track stock levels
  So that I can prevent overselling

  Scenario: Add stock to inventory
    Given the inventory has 10 items of "Widget"
    When I add 5 items of "Widget"
    Then the inventory should have 15 items of "Widget"

  Scenario: Remove stock from inventory
    Given the inventory has 10 items of "Widget"
    When I remove 3 items of "Widget"
    Then the inventory should have 7 items of "Widget"

  Scenario: Cannot remove more than available
    Given the inventory has 5 items of "Widget"
    When I try to remove 10 items of "Widget"
    Then the operation should fail with "Insufficient stock"
