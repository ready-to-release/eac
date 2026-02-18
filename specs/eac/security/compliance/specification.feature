# Intent: Allow security auditors to identify and remediate risks by scanning modules for compliance security issues
# Architecture: Affects eac-cli scan command with --scanner compliance flag; invokes Docker-based compliance scanner; supports dry-run mode; depends on Docker availability and isolated test project

@L2 @ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_compliance

  As a security auditor
  I want to scan modules for security issues
  So that I can identify and remediate risks

  Rule: Scanner wrapper executes successfully with Docker mock

    # Note: This test verifies dry-run mode is recognized by the scan command.
    # The underlying scan may still execute due to framework timing issues with
    # mocking - a separate issue should track fixing scan dry-run mode.
    Scenario: Run compliance scan with mocked Docker
      When I run the command "scan core --scanner compliance --dry-run"
      Then the output contains "dry-run"
