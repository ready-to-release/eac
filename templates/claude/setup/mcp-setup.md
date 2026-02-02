# MCP Server Setup Guide

## What is MCP?

Model Context Protocol (MCP) allows Claude to interact with external tools and data sources.

## EAC MCP Server Setup

### Production Configuration (Default)

**File**: `.mcp.json`

```json
{
  "mcpServers": {
    "commands": {
      "command": "eac-mcp-commands"
    }
  }
}
```

**Uses**: `r2r eac <command>` (Docker-based)
**Startup**: ~2s
**Best for**: CI/CD, production, consistent environments

### Development Configuration (Fast Iteration)

**File**: `.mcp.json`

```json
{
  "mcpServers": {
    "commands": {
      "command": "go",
      "args": ["run", "./go/mcp/commands/main.go"],
      "env": {
        "EAC_USE_DIRECT_BINARY": "true"
      }
    }
  }
}
```

**Uses**: Direct binary execution
**Startup**: ~100ms
**Best for**: Local development, fast feedback

## Environment Variable

**`EAC_USE_DIRECT_BINARY`**:
- `true`: Direct binary (development)
- Not set or `false`: r2r eac (production, default)

## Verification

```bash
# Test production mode
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | r2r eac mcp commands

# Test development mode
EAC_USE_DIRECT_BINARY=true go run ./go/mcp/commands/main.go
```

## Troubleshooting

**Issue**: r2r not found
- **Solution**: Use development mode

**Issue**: Slow commands
- **Solution**: Use development mode (no Docker overhead)

**Issue**: Binary not found
- **Solution**: Build binary (`go build -o out/tools/eac ./go/cli/eac`)

## Available MCP Commands

Once configured, Claude can use these commands via MCP:

### Discovery Commands
- `get-modules` - List all modules
- `get-dependencies <module>` - Show module dependencies
- `show-modules` - Display module details
- `get-files-by-module <module>` - List module files

### Build & Test Commands
- `build <module>` - Build a module
- `test <module>` - Run module tests
- `get-tests` - List all tests
- `show-test-results` - View test outcomes

### Validation Commands
- `validate-contracts` - Verify module contracts
- `validate-module-hierarchy` - Check dependency graph
- `validate-artifacts` - Verify build outputs
- `validate-dependencies` - Check dependency integrity

### Configuration Commands
- `get-config` - View configuration
- `show-config` - Display configuration details

## Benefits of MCP Integration

**Auto-Discovery**: Claude uses MCP commands to discover your project structure instead of guessing.

**Real-Time Validation**: Claude can verify changes using `build`, `test`, and `validate` commands.

**Language-Agnostic**: MCP commands work for any project using EAC (Go, Python, JavaScript, etc.).

**Efficient Workflow**: Claude follows systematic workflows using MCP commands at each step.

## Next Steps

1. Choose configuration (production or development)
2. Add `.mcp.json` to your project root
3. Restart Claude Code
4. Try asking Claude to use MCP commands
5. Install generic templates: `r2r eac templates install claude`
