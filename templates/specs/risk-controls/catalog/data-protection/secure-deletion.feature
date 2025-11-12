# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @gdpr @hipaa @fda-21cfr11
# @implementation:required @automation:partial

@risk-control:data-deletion
Feature: Secure Data Deletion

  As a system administrator
  I want to implement Secure Data Deletion
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:data-deletion-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:data-deletion-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
