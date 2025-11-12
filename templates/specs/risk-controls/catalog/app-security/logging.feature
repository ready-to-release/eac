# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @fda-21cfr11 @owasp
# @implementation:required @automation:full

@risk-control:app-logging
Feature: Application Logging

  As a system administrator
  I want to implement Application Logging
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:app-logging-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:app-logging-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
