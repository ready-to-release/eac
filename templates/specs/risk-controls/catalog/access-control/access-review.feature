# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:detective
# @iso27001 @hipaa @pci-dss @fda-21cfr11 @sox
# @implementation:required @automation:partial

@risk-control:access-review
Feature: Access Review

  As a system administrator
  I want to implement Access Review
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:access-review-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:access-review-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
