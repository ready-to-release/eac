# Feature ID: src_mcp_vscode_module_detection
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode
# Part: 3 of 5 - Infrastructure Module Detection

@mcp @vscode @modules @infrastructure
Feature: Infrastructure Module Detection

  Background:
    Given the module detection system is initialized

  @success @ac3
  Scenario: Detect postgres container module
    When I determine module for "containers/postgres/Dockerfile"
    Then the detected module is "postgres"

  @success @ac3
  Scenario: Detect redis container module
    When I determine module for "containers/redis/docker-compose.yml"
    Then the detected module is "redis"

  @success @ac3
  Scenario: Detect nested container configuration
    When I determine module for "containers/postgres/config/init.sql"
    Then the detected module is "postgres"

  @success @ac4
  Scenario: Detect module from contract filename
    When I determine module for "contracts/deployable-units/0.1.0/src-mcp-vscode.yml"
    Then the detected module is "src-mcp-vscode"

  @success @ac4
  Scenario: Detect CLI module from contract
    When I determine module for "contracts/deployable-units/0.1.0/cli.yml"
    Then the detected module is "cli"

  @success @ac4
  Scenario: Detect automation module from contract
    When I determine module for "contracts/deployable-units/0.1.0/automation-cli-deploy.yml"
    Then the detected module is "automation-cli-deploy"

  @success @ac4
  Scenario: Handle contracts with hyphens in name
    When I determine module for "contracts/deployable-units/0.1.0/src-mcp-slack.yml"
    Then the detected module is "src-mcp-slack"
