@L2 @deps:docker @ov @env:isolated-test-project @control:ai-2
Feature: eac-commands_create-design

  As a developer of the eac platform
  I want AI-powered architecture design generation
  So that I can create Structurizr DSL workspace files from source code analysis

  Background:
    Given I am in a git repository
    And a repository with AI configs at "contracts/ai/design/0.1.0"

  Rule: Module must exist in module contracts

    @L2 @ov
    Scenario: Create design for existing module
      Given a module contract exists for "test-module" at "go/test/module"
      And source code exists at "go/test/module"
      And docker service is available
      And the mock AI is configured to return a valid workspace
      When I run "create design test-module"
      Then the exit code should be 0
      And the file "specs/test-module/.design/workspace.dsl" should exist
      And the output should contain "Architecture design created"

    @L2 @ov
    Scenario: Fail when module does not exist in contracts
      When I run "create design nonexistent-module"
      Then the exit code should be 1
      And the output should contain "module not found"

    @L2 @ov
    Scenario: Fail when source code does not exist
      Given a module contract exists for "test-module" at "go/test/module"
      When I run "create design test-module"
      Then the exit code should be 1
      And the output should contain "source code not found"

  Rule: Docker validation can be skipped

    @L2 @ov
    Scenario: Skip validation with flag
      Given a module contract exists for "test-module" at "go/test/module"
      And source code exists at "go/test/module"
      And the mock AI is configured to return a valid workspace
      When I run "create design test-module --skip-validation"
      Then the exit code should be 0
      And the output should contain "Skipping Docker validation"

  Rule: Existing files require force flag

    @L2 @ov
    Scenario: Fail when workspace exists without force
      Given a module contract exists for "test-module" at "go/test/module"
      And source code exists at "go/test/module"
      And a workspace file exists at "specs/test-module/.design/workspace.dsl"
      When I run "create design test-module"
      Then the exit code should be 1
      And the output should contain "workspace already exists"
      And the output should contain "Use --force to overwrite"

    @L2 @ov
    Scenario: Overwrite with force flag
      Given a module contract exists for "test-module" at "go/test/module"
      And source code exists at "go/test/module"
      And a workspace file exists at "specs/test-module/.design/workspace.dsl"
      And docker service is available
      And the mock AI is configured to return a valid workspace
      When I run "create design test-module --force"
      Then the exit code should be 0
      And the output should contain "Architecture design created"

