# Usage Guide

This guide walks you through using the VSCode extension and MCP servers.

## 🎬 Video Demos

Visual learner? Check out these demo guides:

- [Quick Start Demo](docs/assets/quick-start-guide.md) - Complete setup walkthrough
- [Git Commit Demo](docs/assets/git-commit-demo.md) - Using the extension button
- [MCP Server Test](docs/assets/mcp-server-test.md) - Testing the server manually

*Note: GIFs are not yet recorded. See the storyboards above for instructions.*

## Table of Contents

1. [Initial Setup](#initial-setup)
2. [Testing the MCP Server](#testing-the-mcp-server)
3. [Using the VSCode Extension](#using-the-vscode-ext-claude-commit)
4. [Example Workflows](#example-workflows)
5. [Advanced Usage](#advanced-usage)

---

## Initial Setup

### Step 1: Run Initialization Script

From the project root:

```bash
./automation/sh-vscode/init.sh
```

**Expected Output:**

```text
==================================
Initializing CLI Project
==================================

Checking prerequisites...

✓ Go go1.21.0
✓ Node.js v18.17.0
✓ npm 9.6.7

==================================
Installing VSCode Extension
==================================

Installing dependencies...
[npm install output]

Compiling TypeScript...
[compilation output]

==================================
Verifying MCP Servers
==================================

Checking pwsh server...
  ✓ pwsh server ready
Checking docs server...
  ✓ docs server ready
Checking github server...
  ✓ github server ready
Checking vscode server...
  ✓ vscode server ready

==================================
✓ Initialization Complete!
==================================

Next steps:
1. Open this workspace in VSCode
2. Press F5 to launch the Extension Development Host
3. Open the Source Control view (Ctrl+Shift+G)
4. Click the robot icon button in the toolbar
```

---

## Testing the MCP Server

### Manual Test of VSCode MCP Server

You can test the MCP server directly before using it with the extension:

```bash
cd src/mcp/vscode
go run .
```

The server will wait for JSON-RPC input. Test it with:

**1. Initialize the server:**

```json
{"jsonrpc":"2.0","id":1,"method":"initialize"}
```

**Expected Response:**

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "result":{
    "protocolVersion":"2024-11-05",
    "serverInfo":{"name":"mcp-server-vscode","version":"0.1.0"},
    "capabilities":{"tools":{}}
  }
}
```

**2. List available tools:**

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

**Expected Response:**

```json
{
  "jsonrpc":"2.0",
  "id":2,
  "result":{
    "tools":[
      {
        "name":"vscode-action",
        "description":"Execute a VSCode action",
        "inputSchema":{
          "type":"object",
          "properties":{
            "action":{"type":"string","description":"Action to execute (e.g., 'git-commit', 'git-push', 'git-pull')"},
            "message":{"type":"string","description":"Optional message for the action"}
          },
          "required":["action"]
        }
      }
    ]
  }
}
```

**3. Call a tool:**

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vscode-action","arguments":{"action":"git-commit","message":"test commit"}}}
```

**Expected Response:**

```json
{
  "jsonrpc":"2.0",
  "id":3,
  "result":{
    "content":[
      {"type":"text","text":"Executing action: git-commit with message: test commit"}
    ]
  }
}
```

Press `Ctrl+C` to exit the server.

---

## Using the VSCode Extension

### Step 1: Launch Extension Development Host

1. **Open VSCode** in the project directory:

   ```bash
   code .
   ```

2. **Press F5** (or Run → Start Debugging)
   - This opens a new VSCode window titled `[Extension Development Host]`

3. **In the Extension Development Host window**:
   - Open a folder that has a Git repository
   - Or initialize a new Git repo: `git init`

### Step 2: Find the Extension Button

1. **Open Source Control view:**
   - Click the Source Control icon in the Activity Bar (left sidebar)
   - Or press `Ctrl+Shift+G` (Windows/Linux) or `Cmd+Shift+G` (macOS)

2. **Locate the button:**
   - Look for a **robot icon** (🤖) in the Source Control toolbar
   - It's next to the refresh and more actions buttons

### Step 3: Use the Extension

1. **Click the robot icon button**

2. **Select an action** from the Quick Pick menu:
   - `Git Commit` - Stage and commit changes
   - `Git Push` - Push to remote
   - `Git Pull` - Pull from remote
   - `Custom Action` - Custom command

3. **Enter a message** (if prompted):
   - For Git Commit or Custom Action
   - Example: "Initial commit"
   - Press Enter or leave blank

4. **View the result**:
   - A notification appears with the MCP server response
   - Check the Output panel for detailed logs

---

## Example Workflows

### Workflow 1: Quick Git Commit

```text
1. Make changes to files in your workspace
2. Click robot icon in Source Control view
3. Select "Git Commit"
4. Enter commit message: "Add new feature"
5. Press Enter
6. View success notification
```

### Workflow 2: Push to Remote

```text
1. Ensure you have commits to push
2. Click robot icon in Source Control view
3. Select "Git Push"
4. (No message needed)
5. View push result notification
```

### Workflow 3: Custom Action

```text
1. Click robot icon in Source Control view
2. Select "Custom Action"
3. Enter custom message: "deploy-staging"
4. MCP server processes the custom action
5. View result
```

### Workflow 4: Development Cycle

```bash
# 1. Make changes to extension code
vim .vscode/extensions/claude-mcp-vscode/src/extension.ts

# 2. In the Extension Development Host window:
#    Press Ctrl+R (Cmd+R on macOS) to reload the extension

# 3. Test the changes by clicking the robot icon

# 4. View logs in the original VSCode window:
#    Debug Console panel shows extension logs
```

---

## Advanced Usage

### Debugging the Extension

**View Extension Logs:**

1. In the **original VSCode window** (not Extension Development Host)
2. Open Debug Console: `View → Debug Console`
3. See `console.log()` output from the extension

**View MCP Server Output:**

1. The extension captures stdout/stderr from the Go server
2. Check notifications for MCP server responses
3. Add error handling in `extension.ts` for detailed debugging

### Modifying the MCP Server

**Add a new action:**

Edit `src/mcp/vscode/main.go`:

```go
func executeAction(action string, message string) string {
 switch action {
 case "git-commit":
  return handleGitCommit(message)
 case "git-push":
  return handleGitPush()
 case "custom-deploy":
  return handleCustomDeploy(message)
 default:
  return fmt.Sprintf("Executing action: %s with message: %s", action, message)
 }
}

func handleCustomDeploy(message string) string {
 // Your custom logic here
 return fmt.Sprintf("Deploying to: %s", message)
}
```

**Reload the extension** (Ctrl+R in Extension Development Host) to test changes.

### Adding Extension Commands

Edit `.vscode/extensions/claude-mcp-vscode/package.json`:

```json
"contributes": {
  "commands": [
    {
      "command": "claude-mcp-vscode.callMCP",
      "title": "Call Claude MCP",
      "icon": "$(robot)"
    },
    {
      "command": "claude-mcp-vscode.quickCommit",
      "title": "Quick Commit",
      "icon": "$(git-commit)"
    }
  ]
}
```

Implement in `src/extension.ts`:

```typescript
let quickCommit = vscode.commands.registerCommand('claude-mcp-vscode.quickCommit', async () => {
    // Implementation here
});

context.subscriptions.push(quickCommit);
```

### Watch Mode Development

For faster development, use watch mode:

```bash
cd .vscode/extensions/claude-mcp-vscode
npm run watch
```

Now TypeScript auto-recompiles on every save. Just reload the Extension Development Host (Ctrl+R) after changes.

---

## Troubleshooting

### Extension button not visible

**Check:**

```bash
# Ensure you're in a Git repository
git status

# If not, initialize Git
git init
```

The button only appears when `scmProvider == git`.

### MCP Server connection errors

**Test the server manually:**

```bash
cd src/mcp/vscode
go run .
# Paste JSON-RPC test commands (see above)
```

**Check Go installation:**

```bash
go version
# Should show: go version go1.21.0 or higher
```

### Build errors in extension

**Clean and restore:**

```bash
./automation/sh-vscode/clean.sh
./automation/sh-vscode/init.sh
```

### Extension not reloading

**Hard reload:**

1. Close Extension Development Host window
2. Press F5 again in original window
3. Or: `Ctrl+Shift+P` → "Developer: Reload Window"

---

## Next Steps

1. **Implement actual Git operations** in `src/mcp/vscode/main.go`
2. **Add more MCP tools** for different VSCode actions
3. **Create keyboard shortcuts** for common actions
4. **Add status bar items** for quick access
5. **Implement error handling** for better user experience

## Getting Help

- Check [README.md](README.md) for project overview
- Check [automation/sh-vscode/README.md](automation/sh-vscode/README.md) for script details
- Review MCP server code in `src/mcp/vscode/main.go`
- Review extension code in `.vscode/extensions/claude-mcp-vscode/src/extension.ts`
