# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @gdpr @hipaa @iso27001 @ccpa
# @implementation:required @automation:partial

@risk-control:privacy-pia
Feature: Privacy Impact Assessment

  As a system administrator
  I want to implement Privacy Impact Assessment
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @gdpr @hipaa @iso27001 @ccpa
# @implementation:required @automation:partial

    @risk-control:privacy-pia-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:medium @risk-type:security @control-type:preventive
# @gdpr @hipaa @iso27001 @ccpa
# @implementation:required @automation:partial

    @risk-control:privacy-pia-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
