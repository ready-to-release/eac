@L2 @ov @deps:go @env:isolated-test-project
Feature: eac-cli_templates

  As a developer of the eac platform
  I want to install project templates
  So that I can bootstrap project structure efficiently

  # NOTE: All tests use temporary directories for full isolation
  # Templates are always sourced from local repository directories
  # Cleanup is automatic via sc.After() hook

  Rule: Templates install subcommand copies templates without value replacement

    Scenario: Install docs template to default location
      When I run the templates command "templates install docs"
      Then the command should succeed
      And the file "docs/reference/reference/implementation-plan.md" should exist
      And the file "docs/reference/reference/intended-use.md" should exist

    Scenario: Install docs template to custom destination
      When I run the templates command "templates install docs --destination custom-docs"
      Then the command should succeed
      And the file "custom-docs/reference/implementation-plan.md" should exist
      And the file "custom-docs/reference/intended-use.md" should exist

    Scenario: Install AI templates
      When I run the templates command "templates install ai"
      Then the command should succeed
      And the file ".r2r/eac/templates/ai/commit-message/module.md" should exist
      And the file ".r2r/eac/templates/ai/design/design.md" should exist

    Scenario: Install specs templates
      When I run the templates command "templates install specs"
      Then the command should succeed
      And the file "specs/.risk-controls/risk-catalog/controls.catalog.json" should exist

    Scenario: Install with debug flag creates debug logs
      When I run the templates command "templates install reports --debug"
      Then the command should succeed
      And debug files should exist in "out/logs/templates/install"
