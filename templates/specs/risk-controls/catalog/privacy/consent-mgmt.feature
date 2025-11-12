# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @gdpr @hipaa @fda-21cfr11 @ccpa @lgpd
# @implementation:required @automation:partial

@risk-control:privacy-consent
Feature: Consent Management

  As a system administrator
  I want to implement Consent Management
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @gdpr @hipaa @fda-21cfr11 @ccpa @lgpd
# @implementation:required @automation:partial

    @risk-control:privacy-consent-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @gdpr @hipaa @fda-21cfr11 @ccpa @lgpd
# @implementation:required @automation:partial

    @risk-control:privacy-consent-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
