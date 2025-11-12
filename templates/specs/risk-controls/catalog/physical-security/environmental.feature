# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @iso13485
# @implementation:required @automation:full

@risk-control:physical-environmental
Feature: Environmental Controls

  As a system administrator
  I want to implement Environmental Controls
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:physical-environmental-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:physical-environmental-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
