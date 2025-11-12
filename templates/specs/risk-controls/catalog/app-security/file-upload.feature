# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @owasp
# @implementation:required @automation:full

@risk-control:app-file-upload
Feature: File Upload Security

  As a system administrator
  I want to implement File Upload Security
  So that security and compliance requirements are met

  Rule: Control is implemented

    @risk-control:app-file-upload-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
    @risk-control:app-file-upload-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
