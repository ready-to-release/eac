# Feature ID: src_mcp_vscode_tool_registration
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @critical @tools
Feature: Tool Registration and Discovery

  Background:
    Given the MCP server is initialized

  @success @ac1
  Scenario: List all available tools
    When I send a "tools/list" request
    Then the response contains a "tools" array
    And the tools array is not empty
    And the response follows JSON-RPC 2.0 format

  @success @ac2
  Scenario: Each tool has complete metadata
    When I send a "tools/list" request
    Then each tool in the response has a "name" field
    And each tool has a "description" field
    And each tool has an "inputSchema" field
    And each inputSchema has a "type" field equal to "object"
    And each inputSchema has a "properties" field

  @success @ac3
  Scenario: All registered tools are returned
    When I send a "tools/list" request
    Then the response contains exactly 3 tools
    And the tools array includes "vscode-action"
    And the tools array includes "execute-agent"
    And the tools array includes "quick-commit"

  @success @ac4
  Scenario: vscode-action tool schema defines required and optional params
    When I send a "tools/list" request
    And I find the "vscode-action" tool
    Then the tool inputSchema properties include "action"
    And the tool inputSchema properties include "message"
    And the tool inputSchema required array includes "action"
    And the tool inputSchema required array does not include "message"
    And "action" has description "Action to execute (e.g., 'git-commit', 'git-push', 'git-pull')"
    And "message" has description "Optional message for the action"

  @success @ac4
  Scenario: execute-agent tool schema defines required param
    When I send a "tools/list" request
    And I find the "execute-agent" tool
    Then the tool inputSchema properties include "agentFile"
    And the tool inputSchema required array includes "agentFile"
    And "agentFile" has description "Path to the agent file (for reference)"

  @success @ac4
  Scenario: quick-commit tool schema has no required params
    When I send a "tools/list" request
    And I find the "quick-commit" tool
    Then the tool inputSchema properties is an empty object
    And the tool inputSchema has no required array
