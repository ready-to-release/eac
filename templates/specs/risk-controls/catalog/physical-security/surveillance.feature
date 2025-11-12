# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @iso27001 @gdpr
# @implementation:required @automation:full

@risk-control:physical-surveillance
Feature: Surveillance Systems

  As a system administrator
  I want to implement Surveillance Systems
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:physical-surveillance-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:physical-surveillance-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
