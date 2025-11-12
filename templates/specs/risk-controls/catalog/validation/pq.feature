# @industry:PHARMA @industry:MEDDEV
# @severity:high @risk-type:security @control-type:preventive
# @fda-21cfr11 @gamp5 @iso13485
# @implementation:required @automation:partial

@risk-control:validation-pq
Feature: Performance Qualification

  As a system administrator
  I want to implement Performance Qualification
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:validation-pq-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:validation-pq-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
