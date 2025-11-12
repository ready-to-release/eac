# ========================================
# Risk Control Template
# ========================================
#
# Controls define WHAT must be controlled (requirements).
# Implementations define HOW (solutions in src/features/).
#
# PROCESS:
# 1. Conduct risk assessment (MANDATORY FIRST STEP)
# 2. Select/tailor controls based on YOUR assessment findings
# 3. Create control feature files (this template)
# 4. Link to implementations via @risk-control:<name>-<id> tags
#
# USING THE CATALOG:
# - catalog/ contains 299 example controls (INSPIRATION ONLY)
# - DO NOT copy as-is - mappings are INDICATIVE ONLY
# - Conduct YOUR risk assessment with qualified personnel
# - Validate regulatory applicability for YOUR context
#
# STRUCTURE:
# - Location: specs/risk-controls/<control-name>.feature
# - Naming: kebab-case (e.g., auth-mfa.feature)
# - Tag: @risk-control:<name> (feature), @risk-control:<name>-<id> (scenarios)
# - Link implementations by tagging them with same @risk-control:<name>-<id>
#

@risk-control:<control-name>
Feature: [Control Name]

  # Source: Risk Assessment <ID>, Date: <YYYY-MM-DD>

  Rule: [Control requirement]

    @risk-control:<control-name>-01
    Scenario: [Specific requirement]
      Given [context]
      Then [requirement] MUST [behavior]

    @risk-control:<control-name>-02
    Scenario: [Another requirement]
      Given [context]
      Then [requirement] MUST [behavior]
