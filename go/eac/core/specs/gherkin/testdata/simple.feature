Feature: test-module_simple-calculator

  Rule: Basic arithmetic operations
    @Manual @ov
    Scenario: Add two numbers
      Given I have a calculator
      When I add 2 and 3
      Then the result should be 5

    @L1 @ov
    Scenario: Subtract two numbers
      Given I have a calculator
      When I subtract 3 from 5
      Then the result should be 2
