# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @eu-ai-act @nist-ai-rmf @iso42001
# @implementation:required @automation:partial

@risk-control:ai-versioning
Feature: Model Versioning

  As a system administrator
  I want to implement Model Versioning
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @eu-ai-act @nist-ai-rmf @iso42001
# @implementation:required @automation:partial

    @risk-control:ai-versioning-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:high @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @eu-ai-act @nist-ai-rmf @iso42001
# @implementation:required @automation:partial

    @risk-control:ai-versioning-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
