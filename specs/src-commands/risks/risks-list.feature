@skip:wip @deps:go @L2 @ov @env:isolated-test-project
Feature: src-commands_risks-list

  As a developer
  I want to list risk controls and their linkages
  So that I can understand the risk control landscape

  Background:
    Given I am in a git repository
    And risk controls exist in "specs/risk-controls/"
    And specifications exist in "specs/"

  @iv
  Rule: Command must be registered and accessible

    Scenario: Command is listed
      When I run "list commands"
      Then the exit code is 0
      And I should see "risks-list"

    Scenario: Command shows help
      When I run "risks list --help"
      Then the exit code is 0
      And stdout contains "List risk controls"
      And stdout contains "--filter"
      And stdout contains "-f"
      And stdout contains "--json"
      And stdout contains "-j"

  Rule: Command lists all risk controls by default

    Scenario: List all controls
      Given 5 risk controls exist
      And 3 controls are referenced by specs
      And 2 controls are orphaned
      When I run "risks list"
      Then the exit code is 0
      And the output shows 5 controls
      And each control shows its tags
      And each control shows its reference count
      And the summary shows "Total: 5, Linked: 3, Orphaned: 2"

  Rule: Command can filter controls by status

    Scenario: Show only orphaned controls (long flag)
      Given 5 risk controls exist
      And 2 controls are orphaned
      When I run "risks list --filter orphaned"
      Then the exit code is 0
      And the output shows 2 controls
      And all shown controls have status "Orphaned"

    Scenario: Show only orphaned controls (short flag)
      When I run "risks list -f orphaned"
      Then the output shows only orphaned controls

    Scenario: Show only linked controls (long flag)
      Given 5 risk controls exist
      And 3 controls are referenced
      When I run "risks list --filter linked"
      Then the exit code is 0
      And the output shows 3 controls
      And all shown controls have status "Linked"

    Scenario: Show only linked controls (short flag)
      When I run "risks list -f linked"
      Then the output shows only linked controls

    Scenario: Show specs missing links (long flag)
      Given specs exist that should reference risk controls
      And some specs are missing risk control tags
      When I run "risks list --filter missing-links"
      Then the exit code is 0
      And the output shows specs without risk control links
      And each spec shows suggested controls

    Scenario: Show missing links (short flag)
      When I run "risks list -f missing-links"
      Then the output shows specs missing risk control links

    Scenario: Invalid filter value
      When I run "risks list --filter invalid"
      Then the exit code is 1
      And stderr contains "invalid filter"
      And stderr contains "must be: all, orphaned, linked, or missing-links"

  Rule: Command supports JSON output

    Scenario: Output as JSON
      Given 3 risk controls exist
      When I run "risks list --json"
      Then the exit code is 0
      And the output is valid JSON
      And JSON contains "controls" array
      And JSON contains "summary" object

    Scenario: JSON output (short flag)
      When I run "risks list -j"
      Then the output is valid JSON

  Rule: Command shows detailed linkage information

    Scenario: Show which specs reference each control
      Given a control "specs/risk-controls/auth/mfa.feature"
      And the control has tag "@risk-control:auth-mfa-01"
      And "specs/auth/login.feature" references the tag at line 45
      And "specs/auth/session.feature" references the tag at line 23
      When I run "risks list"
      Then the output shows the control
      And the output shows "specs/auth/login.feature (line 45)"
      And the output shows "specs/auth/session.feature (line 23)"

  Rule: Command identifies orphaned controls

    Scenario: Flag orphaned controls
      Given a control exists with no references
      When I run "risks list"
      Then the control is marked as "⚠ Orphaned"
      And the summary counts it as orphaned

  Rule: Command performance with large repositories

    Scenario: Handle large number of controls efficiently
      Given 100 risk controls exist
      And 500 specifications exist
      When I run "risks list"
      Then the command completes in reasonable time
      And all controls are analyzed

  Rule: All flags have shorthand versions

    Scenario: Short flags work
      When I run "risks list -f orphaned -j"
      Then orphaned filter is applied
      And JSON output is generated
