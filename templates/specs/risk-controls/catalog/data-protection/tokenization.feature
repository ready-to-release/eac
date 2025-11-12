# @industry:FINANCE @industry:HEALTH @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @pci-dss @hipaa @gdpr
# @implementation:required @automation:full

@risk-control:data-tokenization
Feature: Tokenization and Masking

  As a system administrator
  I want to implement Tokenization and Masking
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:data-tokenization-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:data-tokenization-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
