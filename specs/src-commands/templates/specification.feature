@ov
Feature: src-commands_templates

  As a developer of the eac platform
  I want to install and manage templates with value replacements
  So that I can generate project structures efficiently

  # NOTE: All tests use temporary directories (os.MkdirTemp) for full isolation
  # No git-tracked files are modified during tests
  # Cleanup is automatic via sc.After() hook

  Rule: Template list command scans and displays placeholders

    Scenario: List uses default repository when template not provided
      When I run the command "templates list"
      Then the command should succeed
      And the command should attempt to clone from "https://github.com/ready-to-release/eac"
      And the output should contain "Template Placeholders in 'https://github.com/ready-to-release/eac':"

  Rule: Templates apply subcommand supports template-specific application with defaults

    Scenario: Apply compliance template with all defaults
      When I run the command "templates apply compliance"
      Then the command should succeed
      And the templates should be cloned from "https://github.com/ready-to-release/eac" at "main" branch
      And the source path should be "templates/compliance"
      And the destination should be ".docs/references/compliance"
      And no value replacement should occur

    Scenario: Apply compliance template with custom source
      When I run the command "templates apply compliance --source https://github.com/custom/repo"
      Then the command should succeed
      And the templates should be cloned from "https://github.com/custom/repo" at "main" branch
      And the source path should be "templates/compliance"
      And the destination should be ".docs/references/compliance"

    Scenario: Apply compliance template with custom destination
      When I run the command "templates apply compliance --destination ./custom/path"
      Then the command should succeed
      And the destination should be "./custom/path"

    Scenario: Apply compliance template with value replacements
      Given I have a values file "values.json" with:
        """
        {
          "ProjectName": "MyProject",
          "CompanyName": "ACME Corp"
        }
        """
      When I run the command "templates apply compliance --input-json values.json"
      Then the command should succeed
      And the rendered files should contain replaced values

    Scenario: Apply compliance template with all custom parameters
      Given I have a values file "values.json" with:
        """
        {
          "ProjectName": "MyProject"
        }
        """
      When I run the command "templates apply compliance --source https://github.com/custom/repo --destination ./output --input-json values.json"
      Then the command should succeed
      And the templates should be cloned from "https://github.com/custom/repo"
      And the destination should be "./output"
      And the rendered files should contain replaced values

    Scenario: Apply unknown template should fail gracefully
      When I run the command "templates apply unknown-template"
      Then the command should fail
      And the error output should contain "unknown template: unknown-template"
      And the error output should contain "Available templates: compliance"

  Rule: Templates install subcommand supports template-specific installation with defaults

    Scenario: Install specs template with all defaults
      When I run the command "templates install specs"
      Then the command should succeed
      And the templates should be cloned from "https://github.com/ready-to-release/eac" at "main" branch
      And the source path should be "templates/specs"
      And the destination should be ".r2r/templates/specs"

    Scenario: Install specs template with custom source
      When I run the command "templates install specs --source https://github.com/custom/repo"
      Then the command should succeed
      And the templates should be cloned from "https://github.com/custom/repo" at "main" branch
      And the source path should be "templates/specs"

    Scenario: Install specs template with custom destination
      When I run the command "templates install specs --destination ./custom/templates"
      Then the command should succeed
      And the destination should be "./custom/templates"

    Scenario: Install specs template with all custom parameters
      When I run the command "templates install specs --source https://github.com/custom/repo --destination ./output"
      Then the command should succeed
      And the templates should be cloned from "https://github.com/custom/repo"
      And the destination should be "./output"

    Scenario: Install unknown template should fail gracefully
      When I run the command "templates install unknown-template"
      Then the command should fail
      And the error output should contain "unknown template: unknown-template"
      And the error output should contain "Available templates: specs"

  Rule: Template commands are extensible for new template types

    Scenario: Adding new template requires minimal code changes
      Given the templates command system is implemented
      When a developer adds a new template type "architecture"
      Then they should only need to create a new file "templates/apply/architecture.go"
      Or they should only need to create a new file "templates/install/architecture.go"
      And the file should register the template with default configuration
      And the command "templates apply architecture" should automatically work
      Or the command "templates install architecture" should automatically work
