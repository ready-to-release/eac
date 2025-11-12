# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @pci-dss @fda-21cfr11 @nist-800-53
# @implementation:required @automation:partial

@risk-control:encrypt-keys
Feature: Encryption Key Management

  As a system administrator
  I want to implement Encryption Key Management
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:encrypt-keys-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:encrypt-keys-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
