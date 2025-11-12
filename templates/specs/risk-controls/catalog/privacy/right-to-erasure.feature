# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @gdpr @hipaa @ccpa @lgpd
# @implementation:required @automation:partial

@risk-control:privacy-erasure
Feature: Right to Erasure

  As a system administrator
  I want to implement Right to Erasure
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @gdpr @hipaa @ccpa @lgpd
# @implementation:required @automation:partial

    @risk-control:privacy-erasure-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @gdpr @hipaa @ccpa @lgpd
# @implementation:required @automation:partial

    @risk-control:privacy-erasure-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
