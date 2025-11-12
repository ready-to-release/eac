# @industry:PHARMA @industry:MEDDEV
# @severity:medium @risk-type:security @control-type:preventive
# @fda-21cfr11 @iso13485
# @implementation:required @automation:partial

@risk-control:validation-docs
Feature: Validation Documentation

  As a system administrator
  I want to implement Validation Documentation
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:validation-docs-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:validation-docs-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
