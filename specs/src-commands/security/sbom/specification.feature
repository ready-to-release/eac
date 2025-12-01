@ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_sbom

  As a security auditor
  I want to generate Software Bill of Materials (SBOM) for modules
  So that I can track dependencies for compliance and supply chain security

  Rule: SBOM scanner creates evidence file with integrity verification

    Scenario: Generate SBOM for single module with mocked tool
      When I run the command "security sbom src-core"
      Then the exit code is 0
      And evidence files should exist in directory "out/security/src-core/sbom/"
      And the latest evidence file should have JSON field "module" with value "src-core"
      And the latest evidence file should have JSON field "scanner" with value "sbom"
      And the latest evidence file should have JSON field "timestamp" matching RFC3339 format
      And the latest evidence file should have JSON field "sha256" with 64 character hex hash
      And the latest evidence file should have JSON field "findings" with non-empty data

    Scenario: SBOM with debug logging enabled
      When I run the command "security sbom src-core --debug"
      Then the exit code is 0
      And a log file should exist in directory "out/logs/security/"

    Scenario: SBOM with invalid module moniker
      When I run the command "security sbom nonexistent-module-xyz"
      Then the exit code is 1
      And I should see "not found" or "Error"

    Scenario: SBOM for all modules when no arguments provided
      When I run the command "security sbom"
      Then the exit code is 0
      And I should see "modules" or "Scanning"

    Scenario: SBOM help is accessible
      When I run the command "security sbom --help"
      Then the exit code is 0
      And I should see "SBOM" or "Software Bill of Materials" or "Usage"
