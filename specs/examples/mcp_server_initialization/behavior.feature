# Feature ID: src_mcp_vscode_mcp_server_initialization
# Acceptance Spec: acceptance.spec
# Module: src-mcp-vscode

@mcp @vscode @critical @initialization
Feature: MCP Server Initialization

  Background:
    Given the MCP server executable is available
    And the server is not running

  @success @ac1
  Scenario: Initialize server with valid JSON-RPC request
    Given I start the MCP server process
    When I send an initialize request via stdin
    Then the server responds with protocol version "2024-11-05"
    And the response has JSONRPC field "2.0"
    And the response ID matches the request ID

  @success @ac2
  Scenario: Server returns correct server information
    Given I start the MCP server process
    When I send an initialize request
    Then the response contains serverInfo
    And serverInfo name is "mcp-server-vscode"
    And serverInfo version is "0.1.0"

  @success @ac3
  Scenario: Server declares tool capabilities
    Given I start the MCP server process
    When I send an initialize request
    Then the response contains capabilities
    And capabilities includes "tools" capability
    And tools capability is a valid object

  @success @ac4
  Scenario: Server follows JSON-RPC 2.0 specification
    Given I start the MCP server process
    When I send a valid JSON-RPC initialize request with ID 1
    Then the response has "jsonrpc" field equal to "2.0"
    And the response has "id" field equal to 1
    And the response has "result" field
    And the response does not have "error" field

  @error @ac4
  Scenario: Server handles malformed JSON gracefully
    Given I start the MCP server process
    When I send invalid JSON via stdin
    Then the server responds with a parse error
    And the error code is -32700
    And the error message is "Parse error"

  @error @ac4
  Scenario: Server handles empty input
    Given I start the MCP server process
    When I send an empty line via stdin
    Then the server continues listening
    And no response is sent
