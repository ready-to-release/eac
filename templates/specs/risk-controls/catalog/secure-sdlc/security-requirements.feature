# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @owasp @nist-800-53 @nist-ssdf
# @implementation:required @automation:partial

@risk-control:sdlc-requirements
Feature: Security Requirements Definition

  As a system administrator
  I want to implement Security Requirements Definition
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @owasp @nist-800-53 @nist-ssdf
# @implementation:required @automation:partial

    @risk-control:sdlc-requirements-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @owasp @nist-800-53 @nist-ssdf
# @implementation:required @automation:partial

    @risk-control:sdlc-requirements-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
