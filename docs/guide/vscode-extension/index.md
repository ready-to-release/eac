# VSCode Extension Guide

The VSCode extension adds MCP integration to VSCode with a Git toolbar button that communicates with the local MCP server.

## Overview

The extension provides:
- **Robot button** (🤖) in the Git Source Control toolbar
- **MCP integration** - Calls the local `vscode` MCP server
- **Git actions** - Commit, push, pull through MCP
- **Extensible** - Easy to add new actions

## Quick Start

### Prerequisites

- VSCode 1.80.0+
- Extension initialized: `./automation/sh-vscode/init.sh`
- Git repository open

### Using the Extension

1. **Open Source Control**
   - Press `Ctrl+Shift+G` (Windows/Linux) or `Cmd+Shift+G` (macOS)
   - Or click the Source Control icon in the Activity Bar

2. **Click the Robot Button** (🤖)
   - Located in the toolbar at the top of Source Control view

3. **Select an Action**
   - **Git Commit** - Stage and commit changes
   - **Git Push** - Push to remote
   - **Git Pull** - Pull from remote
   - **Custom Action** - Run custom command

4. **Enter Message** (if prompted)
   - Type your message
   - Press Enter

5. **View Result**
   - Notification shows MCP server response

## Architecture

### Extension Structure

```
.vscode/extensions/claude-mcp-vscode/
├── src/
│   └── extension.ts          # Main extension code
├── out/                       # Compiled JavaScript
├── package.json               # Extension manifest
├── tsconfig.json             # TypeScript config
└── .vscode/
    ├── launch.json           # Debug configuration
    └── tasks.json            # Build tasks
```

### How It Works

1. **User clicks button** in VSCode Source Control view
2. **Extension spawns** Go MCP server process
3. **Sends JSON-RPC** request with action and message
4. **MCP server processes** the request
5. **Returns response** via stdout
6. **Extension displays** notification with result

### Communication Flow

```
VSCode Extension
    ↓ (spawn process)
Go MCP Server (src/mcp/vscode/main.go)
    ↓ (JSON-RPC via stdin/stdout)
Tool: vscode-action
    ↓ (execute action)
Result
    ↓ (JSON response)
VSCode Notification
```

## Extension Development

### Running in Debug Mode

```bash
# Open VSCode in project root
code .

# Press F5 to launch Extension Development Host
# Or: Run → Start Debugging
```

This opens a new VSCode window titled `[Extension Development Host]` where you can test the extension.

### Making Changes

**Edit the extension:**
```typescript
// .vscode/extensions/claude-mcp-vscode/src/extension.ts

export function activate(context: vscode.ExtensionContext) {
    // Your changes here
}
```

**Reload the extension:**
- Press `Ctrl+R` (Windows/Linux) or `Cmd+R` (macOS) in Extension Development Host
- Or: `Ctrl+Shift+P` → "Developer: Reload Window"

### Viewing Logs

**Extension Logs:**
- Original VSCode window → Debug Console
- Shows `console.log()` output from extension

**MCP Server Logs:**
- Captured in extension via stdout/stderr
- Check notifications for MCP server responses
- Add logging in `extension.ts`:
  ```typescript
  console.log('MCP Server Response:', response);
  ```

### Building

```bash
cd .vscode/extensions/claude-mcp-vscode

# Compile once
npm run compile

# Watch mode (auto-recompile on save)
npm run watch
```

## Customization

### Adding New Actions

**1. Add action to Quick Pick menu:**

Edit `.vscode/extensions/claude-mcp-vscode/src/extension.ts`:

```typescript
const action = await vscode.window.showQuickPick([
    { label: 'Git Commit', value: 'git-commit' },
    { label: 'Git Push', value: 'git-push' },
    { label: 'Git Pull', value: 'git-pull' },
    { label: 'Deploy Staging', value: 'deploy-staging' },  // New!
    { label: 'Custom Action', value: 'custom' }
], {
    placeHolder: 'Select an action to execute'
});
```

**2. Handle action in MCP server:**

Edit `src/mcp/vscode/main.go`:

```go
func executeAction(action string, message string) string {
    switch action {
    case "git-commit":
        return handleGitCommit(message)
    case "git-push":
        return handleGitPush()
    case "deploy-staging":
        return handleDeployStaging(message)  // New!
    default:
        return fmt.Sprintf("Executing action: %s with message: %s", action, message)
    }
}

func handleDeployStaging(environment string) string {
    // Your custom logic here
    return fmt.Sprintf("Deploying to: %s", environment)
}
```

**3. Test:**
- Reload extension (Ctrl+R)
- Click robot button
- See new action in menu

### Adding New Commands

**1. Register in package.json:**

```json
{
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
}
```

**2. Implement in extension.ts:**

```typescript
let quickCommit = vscode.commands.registerCommand(
    'claude-mcp-vscode.quickCommit',
    async () => {
        // Your implementation
        const message = await vscode.window.showInputBox({
            prompt: 'Commit message'
        });

        if (message) {
            const result = await callMCPServer(mcpServerPath, 'git-commit', message);
            vscode.window.showInformationMessage(`Committed: ${result}`);
        }
    }
);

context.subscriptions.push(quickCommit);
```

**3. Add keyboard shortcut (optional):**

```json
{
  "contributes": {
    "keybindings": [
      {
        "command": "claude-mcp-vscode.quickCommit",
        "key": "ctrl+shift+c",
        "mac": "cmd+shift+c"
      }
    ]
  }
}
```

### Changing the Button Icon

Edit `package.json`:

```json
{
  "contributes": {
    "commands": [
      {
        "command": "claude-mcp-vscode.callMCP",
        "title": "Call Claude MCP",
        "icon": "$(git-commit)"  // Change this
      }
    ]
  }
}
```

Available icons: https://code.visualstudio.com/api/references/icons-in-labels

### Button Location

The button is added to the Source Control (SCM) toolbar. To change location:

```json
{
  "contributes": {
    "menus": {
      "scm/title": [  // Current location
        {
          "command": "claude-mcp-vscode.callMCP",
          "group": "navigation"
        }
      ]
      // Or try:
      // "view/title": [...],
      // "editor/title": [...],
      // "commandPalette": [...]
    }
  }
}
```

## Troubleshooting

### Extension Button Not Visible

**Symptoms:** Can't find the robot button in Source Control view

**Solutions:**

1. **Check Git repository:**
   ```bash
   git status
   ```
   If not a Git repo, initialize:
   ```bash
   git init
   ```

2. **Check Source Control view:**
   - Press `Ctrl+Shift+G`
   - Look at the toolbar (not sidebar)
   - Button only shows when `scmProvider == git`

3. **Reload extension:**
   - Press `Ctrl+R` in Extension Development Host

4. **Check extension is running:**
   - Original VSCode window → Debug Console
   - Should see: "Claude MCP VSCode extension is now active"

### MCP Server Errors

**Symptoms:** "MCP Server error" notification

**Solutions:**

1. **Test server manually:**
   ```bash
   cd src/mcp/vscode
   go run .
   # Paste: {"jsonrpc":"2.0","id":1,"method":"initialize"}
   # Should respond with server info
   ```

2. **Check Go installation:**
   ```bash
   go version
   # Should show: go version go1.21.0 or higher
   ```

3. **Check server path:**
   - Extension looks for server at: `workspace/src/mcp/vscode/`
   - Verify path exists

4. **View error details:**
   - Check Debug Console for stderr output
   - Add logging to `extension.ts`

### Build Errors

**Symptoms:** TypeScript compilation fails

**Solutions:**

```bash
cd .vscode/extensions/claude-mcp-vscode

# Clean and restore
rm -rf node_modules out
npm install
npm run compile

# Or use automation script
cd ../../..
./automation/sh-vscode/restore.sh
```

### Extension Not Reloading

**Symptoms:** Changes not appearing after edit

**Solutions:**

1. **Hard reload:**
   - Close Extension Development Host window
   - Press F5 again in original window

2. **Check compilation:**
   ```bash
   cd .vscode/extensions/claude-mcp-vscode
   npm run compile
   # Check for errors
   ```

3. **Use watch mode:**
   ```bash
   npm run watch
   # Leave running, auto-compiles on save
   ```

### Response Not Showing

**Symptoms:** Click button, no notification appears

**Solutions:**

1. **Check MCP server is responding:**
   - Add logging in `callMCPServer` function
   - Check Debug Console

2. **Check notification permissions:**
   - VSCode → Settings → Notifications
   - Ensure notifications are enabled

3. **Check error handling:**
   ```typescript
   try {
       const result = await callMCPServer(...);
       console.log('Result:', result);
   } catch (error) {
       console.error('Error:', error);
       vscode.window.showErrorMessage(`Error: ${error}`);
   }
   ```

## Testing

### Manual Testing

1. **Create test repository:**
   ```bash
   mkdir test-repo
   cd test-repo
   git init
   echo "test" > file.txt
   git add file.txt
   ```

2. **Open in Extension Development Host:**
   - Press F5 in main window
   - File → Open Folder → `test-repo`

3. **Test each action:**
   - Git Commit ✓
   - Git Push ✓
   - Git Pull ✓
   - Custom Action ✓

4. **Verify responses:**
   - Check notifications
   - Check Debug Console
   - Verify expected behavior

### Automated Testing

Add tests in `.vscode/extensions/claude-mcp-vscode/src/test/`:

```typescript
import * as assert from 'assert';
import * as vscode from 'vscode';

suite('Extension Test Suite', () => {
    test('Extension activates', async () => {
        const ext = vscode.extensions.getExtension('your-publisher.claude-mcp-vscode');
        await ext?.activate();
        assert.ok(ext?.isActive);
    });

    test('Command is registered', async () => {
        const commands = await vscode.commands.getCommands();
        assert.ok(commands.includes('claude-mcp-vscode.callMCP'));
    });
});
```

Run tests:
```bash
npm test
```

## Packaging & Distribution

### Create VSIX Package

```bash
cd .vscode/extensions/claude-mcp-vscode

# Install vsce
npm install -g @vscode/vsce

# Package
vsce package

# Creates: claude-mcp-vscode-0.1.0.vsix
```

### Install Locally

```bash
# From VSIX
code --install-extension claude-mcp-vscode-0.1.0.vsix

# Or from directory
code --install-extension .
```

### Publish to Marketplace

```bash
# Create publisher account at https://marketplace.visualstudio.com/

# Login
vsce login <publisher-name>

# Publish
vsce publish
```

## Advanced Topics

### Status Bar Integration

Add status bar item:

```typescript
const statusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Left,
    100
);
statusBar.text = "$(robot) MCP";
statusBar.command = 'claude-mcp-vscode.callMCP';
statusBar.show();

context.subscriptions.push(statusBar);
```

### Configuration Options

Add settings:

```json
{
  "contributes": {
    "configuration": {
      "title": "Claude MCP",
      "properties": {
        "claudeMcp.serverPath": {
          "type": "string",
          "default": "src/mcp/vscode",
          "description": "Path to MCP server"
        }
      }
    }
  }
}
```

Access in code:
```typescript
const config = vscode.workspace.getConfiguration('claudeMcp');
const serverPath = config.get<string>('serverPath');
```

### Multiple MCP Servers

Call different servers based on action:

```typescript
async function callMCPServer(serverName: string, action: string, message: string) {
    const serverPath = path.join(workspacePath, 'src', 'mcp', serverName);
    // ... rest of implementation
}

// Usage
const result = await callMCPServer('vscode', 'git-commit', message);
const docsResult = await callMCPServer('docs', 'generate-docs', '');
```

## Related Documentation

- [Recording Demo Videos](recording-demos.md) - Create GIF demonstrations
- [MCP Server Development](../mcp-servers/index.md) - Developing MCP servers
- [Automation Scripts](../automation.md) - Scripts for building and testing

## Resources

**VSCode Extension API:**
- [Extension API](https://code.visualstudio.com/api)
- [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)

**MCP Protocol:**
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)

**Extension Development:**
- [Your First Extension](https://code.visualstudio.com/api/get-started/your-first-extension)
- [Publishing Extensions](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
