# Intent: Ensure commits have consistent, high-quality messages by committing changes with AI-generated messages from within a workspace
# Architecture: Affects eac-cli work commit command; validates staged changes exist; invokes AI commit message generation; depends on git workspace context and EAC AI configuration

@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-cli_work-commit

  As a developer working in a workspace
  I want to commit changes with AI-generated messages
  So that my commits have consistent, high-quality messages

  Rule: Validation prevents invalid commits

    Scenario: Fail when no staged changes
      Given I am in the main workspace on branch "main"
      When I run "work commit"
      Then the exit code is 1
      And I see error "No staged changes"
