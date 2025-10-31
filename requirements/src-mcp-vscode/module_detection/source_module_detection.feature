# Feature ID: src_mcp_vscode_module_detection
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode
# Part: 2 of 5 - Source Module Detection

@mcp @vscode @modules @critical @source
Feature: Source Module Detection

  Background:
    Given the module detection system is initialized

  @success @ac2
  Scenario: Detect src-mcp-vscode module
    When I determine module for "src/mcp/vscode/main.go"
    Then the detected module is "src-mcp-vscode"

  @success @ac2
  Scenario: Detect src-mcp-slack module
    When I determine module for "src/mcp/slack/handler.go"
    Then the detected module is "src-mcp-slack"

  @success @ac2
  Scenario: Detect nested files in mcp service
    When I determine module for "src/mcp/vscode/internal/git/context.go"
    Then the detected module is "src-mcp-vscode"

  @success @ac2
  Scenario: Detect test files in mcp service
    When I determine module for "src/mcp/vscode/main_test.go"
    Then the detected module is "src-mcp-vscode"

  @success @ac9
  Scenario: Consistent detection for multiple src files
    When I determine module for "src/mcp/vscode/main.go"
    And I determine module for "src/mcp/vscode/handler.go"
    And I determine module for "src/mcp/vscode/types.go"
    Then all detected modules are "src-mcp-vscode"
