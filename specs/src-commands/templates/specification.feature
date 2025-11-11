@commands
Feature: commands_templates

  As a developer
  I want to install and manage templates with value replacements
  So that I can generate project structures efficiently

  Rule: Template commands require valid inputs and handle errors gracefully

    @success @ac1
    Scenario: Install uses default repository when template not provided
      Given I have a values file "values.json" with:
        """
        {
          "ProjectName": "test"
        }
        """
      When I run command "templates install --values values.json --location ./output"
      Then the command should succeed
      And the file "./output/README.md" should exist
      And the file "./output/README.md" should contain "test"

    @error @ac1
    Scenario: Install fails with non-Git URL template
      Given I have a values file "values.json" with:
        """
        {
          "ProjectName": "test"
        }
        """
      When I run command "templates install --template ./local-path --values values.json --location ./output"
      Then the command should fail
      And the error output should contain "must be a Git repository URL"

    @error @ac1
    Scenario: Install fails without values flag
      When I run command "templates install --template https://github.com/user/repo --location ./output"
      Then the command should fail
      And the error output should contain "--values flag is required"

    @error @ac1
    Scenario: Install fails without location flag
      Given I have a values file "values.json" with:
        """
        {
          "ProjectName": "test"
        }
        """
      When I run command "templates install --template https://github.com/user/repo --values values.json"
      Then the command should fail
      And the error output should contain "--location flag is required"

    @success @ac1
    Scenario: List uses default repository when template not provided
      When I run command "templates list"
      Then the command should succeed
      And the command should attempt to clone from "https://github.com/ready-to-release/eac"
      And the output should contain "Template Placeholders in 'https://github.com/ready-to-release/eac':"

    @success @ac1
    Scenario: List scans local directory when path provided
      Given I have a template directory "test-templates/"
      And I have a template file "test-templates/README.md" with content:
        """
        # {{ .ProjectName }}
        """
      When I run command "templates list --template test-templates/"
      Then the command should succeed
      And the output should contain "{{ .ProjectName }}"
      And the output should contain "Total: 1 placeholders"

    @error @ac1
    Scenario: List fails with non-existent template directory
      When I run command "templates list --template non-existent/"
      Then the command should fail
      And the error output should contain "template directory does not exist"

  Rule: Template scanning discovers placeholders in files

    @success @ac2
    Scenario: List scans files with placeholders in content
      Given I have a template directory "test-templates/"
      And I have a template file "test-templates/config.yaml" with content:
        """
        name: {{ .ProjectName }}
        """
      When I run command "templates list --template test-templates/"
      Then the command should succeed
      And the output should contain "{{ .ProjectName }}"
      And the output should contain "Total: 1 placeholders"

    @success @ac2
    Scenario: List scans placeholders in filenames
      Given I have a template directory "test-templates/"
      And I have a template file "test-templates/README.md" with content:
        """
        # {{ .ProjectName }}
        """
      When I run command "templates list --template test-templates/"
      Then the command should succeed
      And the output should contain "{{ .ProjectName }}"
      And the output should contain "Total: 1 placeholders"
