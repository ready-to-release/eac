# @industry:PHARMA @industry:MEDDEV @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @sox @fda-21cfr11
# @implementation:required @automation:full

@risk-control:audit-changes
Feature: Change Audit Logging

  As a system administrator
  I want to implement Change Audit Logging
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:audit-changes-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:audit-changes-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
