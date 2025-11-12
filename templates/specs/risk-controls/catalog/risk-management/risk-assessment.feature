# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @iso13485 @sox @iso14971 @iso31000
# @implementation:required @automation:partial

@risk-control:risk-assessment
Feature: Risk Assessment

  As a system administrator
  I want to implement Risk Assessment
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @iso13485 @sox @iso14971 @iso31000
# @implementation:required @automation:partial

    @risk-control:risk-assessment-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @fda-21cfr11 @iso13485 @sox @iso14971 @iso31000
# @implementation:required @automation:partial

    @risk-control:risk-assessment-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
