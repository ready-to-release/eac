@ov @env:isolated-test-project
Feature: src-commands_templates

  As a developer of the eac platform
  I want to install and manage templates with value replacements
  So that I can generate project structures efficiently

  # NOTE: All tests use temporary directories (os.MkdirTemp) for full isolation
  # No git-tracked files are modified during tests
  # Cleanup is automatic via sc.After() hook

  Rule: Template list command scans and displays placeholders

    @skip:broken
    Scenario: List uses default repository when template not provided
      When I run the command "templates list"
      Then the command should succeed
      And the command should attempt to clone from "https://github.com/ready-to-release/eac"
      And the output should contain "Template Placeholders in 'https://github.com/ready-to-release/eac':"

  Rule: Templates must preserve placeholders when no values file is provided

    # When no --input-json is provided, template files must be copied as-is
    # Placeholders like {{ .ProjectName }} must remain unchanged in output files
    # This allows users to apply templates without data and fill values later

    Scenario: Template placeholders are preserved when no JSON file provided
      Given a template file contains "# {{ .ProjectName }}"
      When I apply the template without providing --input-json
      Then the output file should contain "# {{ .ProjectName }}"
      And the output file should NOT contain "<no value>"
      And the output file should NOT contain empty string replacement

  Rule: Templates apply subcommand supports template-specific application with defaults

    @skip:broken
    Scenario: Apply docs template with all defaults
      When I run the command "templates apply docs"
      Then the command should succeed
      And the repository should be cloned from "https://github.com/ready-to-release/eac" at "main" branch
      And the templates should be read from subdirectory "templates/docs"
      And the destination should be ".docs/references/docs"
      And no value replacement should occur
      And template placeholders should remain unchanged

    @skip:broken
    Scenario: Apply docs template with custom local source
      Given I have a template directory "custom/templates/docs"
      And I have a template file "custom/templates/docs/README.md" with content:
        """
        # Documentation
        """
      When I run the command "templates apply docs --source ./custom/templates/docs"
      Then the command should succeed
      And the templates should be copied from local path "./custom/templates/docs"
      And the destination should be ".docs/references/docs"

    @skip:broken
    Scenario: Apply docs template with custom destination
      When I run the command "templates apply docs --destination ./custom/path"
      Then the command should succeed
      And the destination should be "./custom/path"

    Scenario: Apply docs template with value replacements
      Given I have a values file "values.json" with:
        """
        {
          "ProjectName": "MyProject",
          "CompanyName": "ACME Corp"
        }
        """
      When I run the command "templates apply docs --input-json values.json"
      Then the command should succeed
      And the rendered files should contain replaced values

    @skip:broken
    Scenario: Apply docs template with all custom parameters
      Given I have a template directory "local/templates"
      And I have a template file "local/templates/doc.md" with content:
        """
        Project: {{ .ProjectName }}
        """
      And I have a values file "values.json" with:
        """
        {
          "ProjectName": "MyProject"
        }
        """
      When I run the command "templates apply docs --source ./local/templates --destination ./output --input-json values.json"
      Then the command should succeed
      And the templates should be copied from local path "./local/templates"
      And the destination should be "./output"
      And the rendered files should contain replaced values

    @skip:broken
    Scenario: Apply unknown template should fail gracefully
      When I run the command "templates apply unknown-template"
      Then the command should fail
      And the error output should contain "unknown template: unknown-template"
      And the error output should contain "Available templates: docs"

  Rule: Templates install subcommand supports template-specific installation with defaults

    @skip:broken
    Scenario: Install specs template with all defaults
      When I run the command "templates install specs"
      Then the command should succeed
      And the repository should be cloned from "https://github.com/ready-to-release/eac" at "main" branch
      And the templates should be read from subdirectory "templates/specs"
      And the destination should be ".r2r/templates/specs"

    @skip:broken
    Scenario: Install specs template with custom local source
      Given I have a template directory "custom/templates/specs"
      And I have a template file "custom/templates/specs/spec.feature" with content:
        """
        Feature: Example
        """
      When I run the command "templates install specs --source ./custom/templates/specs"
      Then the command should succeed
      And the templates should be copied from local path "./custom/templates/specs"
      And the destination should be ".r2r/templates/specs"

    @skip:broken
    Scenario: Install specs template with custom destination
      When I run the command "templates install specs --destination ./custom/templates"
      Then the command should succeed
      And the destination should be "./custom/templates"

    @skip:broken
    Scenario: Install specs template with all custom parameters
      Given I have a template directory "local/templates"
      And I have a template file "local/templates/spec.feature" with content:
        """
        Feature: Test
        """
      When I run the command "templates install specs --source ./local/templates --destination ./output"
      Then the command should succeed
      And the templates should be copied from local path "./local/templates"
      And the destination should be "./output"

    @skip:broken
    Scenario: Install reports template with all defaults
      When I run the command "templates install reports"
      Then the command should succeed
      And the repository should be cloned from "https://github.com/ready-to-release/eac" at "main" branch
      And the templates should be read from subdirectory "templates/reports"
      And the destination should be ".r2r/templates/reports"

    @skip:broken
    Scenario: Install reports template with custom local source
      Given I have a template directory "custom/templates/reports"
      And I have a template file "custom/templates/reports/report.md" with content:
        """
        # Report
        """
      When I run the command "templates install reports --source ./custom/templates/reports"
      Then the command should succeed
      And the templates should be copied from local path "./custom/templates/reports"
      And the destination should be ".r2r/templates/reports"

    @skip:broken
    Scenario: Install reports template with custom destination
      When I run the command "templates install reports --destination ./custom/templates"
      Then the command should succeed
      And the destination should be "./custom/templates"

    @skip:broken
    Scenario: Install unknown template should fail gracefully
      When I run the command "templates install unknown-template"
      Then the command should fail
      And the error output should contain "unknown template: unknown-template"
      And the error output should contain "Available templates: reports, specs"

  Rule: Template commands are extensible for new template types

    Scenario: Adding new apply template requires minimal code changes
      Given the templates command system is implemented
      When a developer adds a new template type "architecture" for apply
      Then they should only need to create a new file "templates/apply/architecture.go"
      And the file should register the template with default configuration
      And the command "templates apply architecture" should automatically work

    Scenario: Adding new install template requires minimal code changes
      Given the templates command system is implemented
      When a developer adds a new template type "architecture" for install
      Then they should only need to create a new file "templates/install/architecture.go"
      And the file should register the template with default configuration
      And the command "templates install architecture" should automatically work

  Rule: Template system must prevent path traversal attacks

    # SECURITY: Malicious templates must not write files outside designated output directory

    @skip:broken
    Scenario: Reject template file with path traversal
      Given a template directory with file "../../etc/passwd.tmpl"
      When I render the template to output directory
      Then the command should fail
      And the error should contain "security: path traversal detected"
      And no files should be written outside the output directory

    @skip:broken
    Scenario: Reject rendered path that escapes after variable substitution
      Given a template directory with file "{{ .Path }}/file.txt"
      And template values with Path="../../../etc"
      When I render the template to output directory
      Then the command should fail
      And the error should contain "security"
      And no files should be written outside the output directory

  Rule: Security validations must be applied consistently across all template operations

    @skip:broken
    Scenario: Security validation errors are clear and actionable
      When a security violation is detected
      Then the error message must include "security:" prefix
      And the error message must describe what was detected
      And no partial files should be created
