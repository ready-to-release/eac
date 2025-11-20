@ov
Feature: src-commands_design-command

  As a developer of the eac platform
  I want to manage architecture documentation using Structurizr
  So that I can create, validate, and view system design through interactive C4 diagrams

  Rule: Command generates architecture documentation from source code using AI

    @ov @L2 @deps:ai @deps:docker
    Scenario: Generate architecture design from source code
      Given docker service is available
      And source code exists at "src/commands"
      And ANTHROPIC_API_KEY environment variable is set
      When I run the command "design create commands --force"
      Then the command should validate Docker availability first
      And AI should analyze source code at "src/commands"
      And workspace.dsl should be generated with quick validation
      And workspace.dsl should be validated with Docker
      And output should be saved to "specs/commands/design/workspace.dsl"
      And I should see success message with file path
      And the exit code is 0

    @ov @L2 @deps:ai @deps:docker
    Scenario: Generate design with custom output path
      Given docker service is available
      And source code exists at "src/commands"
      And ANTHROPIC_API_KEY environment variable is set
      When I run the command "design create commands --output custom/path/workspace.dsl --force"
      Then workspace.dsl should be saved to "custom/path/workspace.dsl"
      And I should see success message with custom path
      And the exit code is 0

    @ov @L2 @deps:ai
    Scenario: Generate design with debug mode
      Given docker service is available
      And source code exists at "src/commands"
      And ANTHROPIC_API_KEY environment variable is set
      When I run the command "design create commands --force --debug"
      Then debug files should be saved to "out/" directory
      And debug files should include "design-debug-full-prompt.md"
      And debug files should include "attempt-1-raw.txt"
      And debug files should include "attempt-1-cleaned.txt"
      And the exit code is 0

    @ov
    Scenario: Prevent overwrite without force flag
      Given source code exists at "src/commands"
      And workspace.dsl already exists at "specs/commands/design/workspace.dsl"
      When I run the command "design create commands"
      Then I should see error message "workspace already exists"
      And I should see instruction "Use --force to overwrite"
      And the exit code is 1

    @ov
    Scenario: Fail when source code not found
      Given source code does not exist at "src/invalid-module"
      When I run the command "design create invalid-module"
      Then I should see error message "source code not found for module 'invalid-module'"
      And I should see expected path in error message
      And I should see usage example
      And the exit code is 1

    @ov @L2 @deps:docker
    Scenario: Fail fast when Docker not available
      Given docker service is not available
      And source code exists at "src/commands"
      When I run the command "design create commands --force"
      Then Docker availability should be checked before AI generation
      And I should see error message "Docker is not running"
      And I should see instruction to start Docker
      And the exit code is 1
      And AI generation should not have occurred

    @ov @L2 @deps:ai @deps:docker
    Scenario: Quick validation catches syntax errors before Docker validation
      Given docker service is available
      And source code exists at "src/commands"
      And ANTHROPIC_API_KEY environment variable is set
      And AI generates output with syntax errors
      When I run the command "design create commands --force"
      Then quick validation should detect errors in under 100ms
      And Docker validation should be skipped for first attempt
      And I should see validation error details
      And AI should retry with error feedback
      And the exit code is 1 if all attempts fail

  Rule: Command starts Structurizr container and displays documentation

    @ov @L2 @deps:docker
    Scenario: Start Structurizr for a module
      Given docker service is available
      And module "src-cli" has workspace.dsl file
      When I run the command "design serve src-cli --no-browser"
      Then Structurizr container should start successfully
      And I should see success message with URL
      And documentation should be accessible at the URL

    @ov
    Scenario: List available modules
      When I run the command "design list"
      Then the exit code is 0
      And I should see a list of available modules
      And "src-cli" module should be in the list

  Rule: Command validates workspace files using Structurizr CLI

    @ov @L2 @deps:docker
    Scenario: Validate one module
      Given docker service is available
      And module "src-cli" has workspace.dsl file
      When I run the command "design validate src-cli"
      Then the workspace should be validated using Structurizr CLI
      And validation results should be displayed in console
      And validation results should be written to JSON file
      And I should see validation summary with errors and warnings

    @ov @L2 @deps:docker
    Scenario: Validate all modules
      Given docker service is available
      And multiple modules have workspace.dsl files
      When I run the command "design validate --all"
      Then all workspaces should be validated using Structurizr CLI
      And validation results for each module should be displayed in console
      And aggregated validation results should be written to JSON file
      And I should see overall summary with total errors and warnings
