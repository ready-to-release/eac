# VSCode MCP Server

A Model Context Protocol (MCP) server for VSCode integration.

## Features

- Execute VSCode actions from Claude
- Git integration (commit, push, pull)
- Extensible action system

## Tools

### vscode-action

Execute a VSCode action.

**Arguments:**
- `action` (required): Action to execute (e.g., 'git-commit', 'git-push', 'git-pull')
- `message` (optional): Optional message for the action

## Usage

The server is designed to be called from a VSCode extension that provides a button interface.

## Development

Run the server:

```bash
./run.sh
```

Or using Go directly:

```bash
go run .
```
