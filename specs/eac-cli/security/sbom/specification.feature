@L2 @ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_sbom

  As a security auditor
  I want to scan modules for security issues
  So that I can identify and remediate risks

  Rule: Scanner wrapper executes successfully with Docker mock

    Scenario: Run sbom scan with mocked Docker  
      When I run the command "scan core --scanner sbom --dry-run"
      Then the exit code is 0
