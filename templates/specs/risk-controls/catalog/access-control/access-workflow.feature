# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:detective
# @iso27001 @fda-21cfr11 @sox
# @implementation:required @automation:full

@risk-control:access-workflow
Feature: Access Request Workflow

  As a system administrator
  I want to implement Access Request Workflow
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:access-workflow-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:access-workflow-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
