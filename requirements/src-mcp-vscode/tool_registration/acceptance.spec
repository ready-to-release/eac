# Tool Registration and Discovery

> **Feature ID**: src_mcp_vscode_tool_registration
> **BDD Scenarios**: [behavior.feature](./behavior.feature)
> **Module**: src-mcp-vscode
> **Tags**: mcp, vscode, critical, tools

## User Story

* As a VSCode extension
* I want to discover available tools via JSON-RPC
* So that I know what actions I can execute through the MCP server

## Acceptance Criteria

* Responds to `tools/list` request with array of tools
* Each tool includes name, description, and inputSchema
* Returns all three registered tools (vscode-action, execute-agent, quick-commit)
* Input schemas include required and optional parameters

## Acceptance Tests

### AC1: Responds to tools/list request with array of tools
**Validated by**: behavior.feature -> @ac1 scenarios

Tags: critical, tools

* Send JSON-RPC tools/list request
* Receive response with tools array
* Verify tools array is not empty
* Verify response follows JSON-RPC 2.0 format

### AC2: Each tool includes name, description, and inputSchema
**Validated by**: behavior.feature -> @ac2 scenarios

Tags: tools

* Send JSON-RPC tools/list request
* Parse tools array from response
* For each tool verify "name" field exists
* For each tool verify "description" field exists
* For each tool verify "inputSchema" field exists
* Verify inputSchema has "type" field
* Verify inputSchema has "properties" field

### AC3: Returns all three registered tools
**Validated by**: behavior.feature -> @ac3 scenarios

Tags: critical, tools

* Send JSON-RPC tools/list request
* Parse tools array
* Verify tool count is 3
* Verify "vscode-action" tool is present
* Verify "execute-agent" tool is present
* Verify "quick-commit" tool is present

### AC4: Input schemas include required and optional parameters
**Validated by**: behavior.feature -> @ac4 scenarios

Tags: tools, validation

* Send JSON-RPC tools/list request
* Find "vscode-action" tool in response
* Verify inputSchema properties includes "action"
* Verify inputSchema properties includes "message"
* Verify required array includes "action"
* Verify "message" is not in required array
