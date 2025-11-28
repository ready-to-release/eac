@L2 @ov @deps:go @deps:git @env:isolated-test-project
Feature: src-commands_templates

  As a developer of the eac platform
  I want to install and manage templates with value replacements
  So that I can generate project structures efficiently

  # NOTE: All tests use temporary directories for full isolation
  # Git cloning is mocked to use local directories
  # Cleanup is automatic via sc.After() hook

  Rule: Template list command scans and displays placeholders

    Scenario: List scans local template directory
      Given I have a template directory "templates"
      And I have a template file "templates/README.md" with content:
        """
        # {{ .ProjectName }}
        """
      When I run the templates command "templates list --template templates"
      Then the command should succeed
      And the output should contain "ProjectName"

    Scenario: List with debug flag saves scan results
      Given I have a template directory "templates"
      And I have a template file "templates/test.md" with content:
        """
        # {{ .Name }}
        """
      When I run the templates command "templates list --template templates --debug"
      Then the command should succeed
      And debug files should exist in "out/logs/templates/list"

  Rule: Templates apply subcommand renders templates with value replacements

    Scenario: Apply docs template with local source
      Given I have a template directory "templates"
      And I have a template file "templates/README.md" with content:
        """
        # Documentation
        """
      When I run the templates command "templates apply docs --source templates --destination output"
      Then the command should succeed
      And the file "output/README.md" should exist
      And the file "output/README.md" should contain "# Documentation"

    Scenario: Apply docs template preserves placeholders when no values provided
      Given I have a template directory "templates"
      And I have a template file "templates/README.md" with content:
        """
        # {{ .ProjectName }}
        """
      When I run the templates command "templates apply docs --source templates --destination output"
      Then the command should succeed
      And the file "output/README.md" should contain "{{ .ProjectName }}"

    Scenario: Apply docs template with value replacements
      Given I have a template directory "templates"
      And I have a template file "templates/README.md" with content:
        """
        # {{ .ProjectName }}
        Company: {{ .CompanyName }}
        """
      And I have a values file "values.json" with:
        """
        {
          "ProjectName": "MyProject",
          "CompanyName": "ACME Corp"
        }
        """
      When I run the templates command "templates apply docs --source templates --destination output --input-json values.json"
      Then the command should succeed
      And the file "output/README.md" should contain "# MyProject"
      And the file "output/README.md" should contain "Company: ACME Corp"

    Scenario: Apply docs template with debug flag
      Given I have a template directory "templates"
      And I have a template file "templates/README.md" with content:
        """
        # Test
        """
      When I run the templates command "templates apply docs --source templates --destination output --debug"
      Then the command should succeed
      And debug files should exist in "out/logs/templates/apply"

  Rule: Templates install subcommand copies templates without value replacement

    Scenario: Install reports template with local source
      Given I have a template directory "templates"
      And I have a template file "templates/report.md" with content:
        """
        # Report Template
        """
      When I run the templates command "templates install reports --source templates --destination output"
      Then the command should succeed
      And the file "output/report.md" should exist
      And the file "output/report.md" should contain "# Report Template"

    Scenario: Install preserves template placeholders
      Given I have a template directory "templates"
      And I have a template file "templates/report.md" with content:
        """
        # {{ .ReportName }}
        """
      When I run the templates command "templates install reports --source templates --destination output"
      Then the command should succeed
      And the file "output/report.md" should contain "{{ .ReportName }}"

    Scenario: Install reports template with debug flag
      Given I have a template directory "templates"
      And I have a template file "templates/report.md" with content:
        """
        # Report
        """
      When I run the templates command "templates install reports --source templates --destination output --debug"
      Then the command should succeed
      And debug files should exist in "out/logs/templates/install"
