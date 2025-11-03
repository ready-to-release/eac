# Claude MCP VSCode Extension

A simple VSCode extension that adds a button to the Git source control view. This button allows you to interact with the Claude MCP server.

## Features

- Adds a button to the Git source control view (SCM)
- Calls the local Claude MCP vscode server
- Supports Git actions (commit, push, pull)
- Extensible for custom actions

## Installation

1. Install dependencies:

   ```bash
   npm install
   ```

2. Compile the extension:

   ```bash
   npm run compile
   ```

3. Press F5 in VSCode to open a new window with the extension loaded

## Usage

1. Open the Source Control view in VSCode (Ctrl+Shift+G)
2. Look for the robot icon button in the toolbar
3. Click the button to select an action
4. The extension will call the MCP server and display the result

## Requirements

- VSCode 1.80.0 or higher
- Go installed (for running the MCP server)
- The MCP server must be located at `src/mcp/vscode/` in your workspace

## Development

To watch for changes and recompile automatically:

```bash
npm run watch
```

## Extension Settings

This extension doesn't require any specific settings.

## Known Issues

None at this time.

## Release Notes

### 0.1.0

Initial release with basic MCP server integration.
