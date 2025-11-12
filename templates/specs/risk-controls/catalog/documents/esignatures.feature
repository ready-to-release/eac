# @industry:PHARMA @industry:MEDDEV @industry:FINANCE
# @severity:critical @risk-type:security @control-type:preventive
# @fda-21cfr11 @iso13485 @esign-act @eidas
# @implementation:required @automation:full

@risk-control:doc-esignatures
Feature: Electronic Signatures

  As a system administrator
  I want to implement Electronic Signatures
  So that security and compliance requirements are met

  Rule: Control is implemented

# @industry:PHARMA @industry:MEDDEV @industry:FINANCE
# @severity:critical @risk-type:security @control-type:preventive
# @fda-21cfr11 @iso13485 @esign-act @eidas
# @implementation:required @automation:full

    @risk-control:doc-esignatures-01
    Scenario: Control is active
      Given the control is configured
      When the system operates
      Then the control requirements are enforced
      
# @industry:PHARMA @industry:MEDDEV @industry:FINANCE
# @severity:critical @risk-type:security @control-type:preventive
# @fda-21cfr11 @iso13485 @esign-act @eidas
# @implementation:required @automation:full

    @risk-control:doc-esignatures-02
    Scenario: Control prevents violations
      Given a user attempts to violate the control
      When the action is detected
      Then it is blocked and logged
