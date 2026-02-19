# Intent: Allow security auditors to identify and remediate risks by scanning modules for exposed secrets and credentials
# Architecture: Affects eac-cli scan command with --scanner secrets flag; invokes Docker-based secrets scanner; supports dry-run mode; depends on Docker availability and isolated test project

@L2 @ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_secrets

  As a security auditor
  I want to scan modules for security issues
  So that I can identify and remediate risks

  Rule: Scanner wrapper executes successfully with Docker mock

    Scenario: Run secrets scan with mocked Docker  
      When I run the command "scan core --scanner secrets --dry-run"
      Then the exit code is 0
