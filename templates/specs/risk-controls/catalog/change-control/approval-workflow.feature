# @industry:PHARMA @industry:MEDDEV @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @sox
# @implementation:required @automation:full

@risk-control:change-approval
Feature: Change Approval Workflow

  As a system administrator
  I want to implement Change Approval Workflow
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:change-approval-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:change-approval-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
