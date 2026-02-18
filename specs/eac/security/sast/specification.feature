# Intent: Allow security auditors to identify and remediate risks by scanning modules for static application security issues
# Architecture: Affects eac-cli scan command with --scanner sast flag; invokes Docker-based SAST scanner; supports dry-run mode; depends on Docker availability and isolated test project

@L2 @ov @deps:go @deps:docker @env:isolated-test-project
Feature: security_sast

  As a security auditor
  I want to scan modules for security issues
  So that I can identify and remediate risks

  Rule: Scanner wrapper executes successfully with Docker mock

    Scenario: Run sast scan with mocked Docker  
      When I run the command "scan core --scanner sast --dry-run"
      Then the exit code is 0
