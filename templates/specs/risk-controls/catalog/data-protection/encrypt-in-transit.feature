# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @hipaa @pci-dss @fda-21cfr11 @gdpr
# @implementation:required @automation:full

@risk-control:encrypt-transit
Feature: Data Encryption in Transit

  As a system administrator
  I want to implement Data Encryption in Transit
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:encrypt-transit-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:encrypt-transit-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
