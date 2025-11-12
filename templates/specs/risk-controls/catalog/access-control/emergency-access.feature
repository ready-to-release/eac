# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:detective
# @iso27001 @fda-21cfr11 @hipaa
# @implementation:required @automation:partial

@risk-control:access-emergency
Feature: Emergency Access

  As a system administrator
  I want to implement Emergency Access
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:access-emergency-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:access-emergency-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
