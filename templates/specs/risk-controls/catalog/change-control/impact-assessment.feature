# @industry:PHARMA @industry:MEDDEV @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @sox
# @implementation:required @automation:partial

@risk-control:change-impact
Feature: Change Impact Assessment

  As a system administrator
  I want to implement Change Impact Assessment
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:change-impact-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:change-impact-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
