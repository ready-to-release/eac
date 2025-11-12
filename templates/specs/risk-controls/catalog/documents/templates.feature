# @industry:PHARMA @industry:MEDDEV @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @fda-21cfr11 @iso13485
# @implementation:required @automation:partial

@risk-control:doc-templates
Feature: Document Templates

  As a system administrator
  I want to implement Document Templates
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:doc-templates-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:doc-templates-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
