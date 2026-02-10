@L2 @deps:python @depm:mathlib
Feature: Geometry Calculations
  As a developer using the math library
  I want accurate geometry calculations
  So that I can compute areas correctly

  Rule: Circle area calculations

    Scenario Outline: Calculate circle area
      When I calculate the area of a circle with radius <radius>
      Then the area should be approximately <area>

      Examples:
        | radius | area    |
        | 1      | 3.14159 |
        | 5      | 78.5398 |
        | 0      | 0       |

    Scenario: Reject negative radius
      When I calculate the area of a circle with radius -1
      Then it should raise a ValueError

  Rule: Rectangle area calculations

    Scenario: Calculate rectangle area
      When I calculate the area of a rectangle with width 3 and height 4
      Then the area should be 12
