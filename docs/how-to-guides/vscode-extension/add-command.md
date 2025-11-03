# Add a New Command

Learn how to register a VSCode command with keyboard shortcuts for the extension.

---

## Overview

**What you'll do:**

- Register a new VSCode command in package.json
- Implement the command handler in extension.ts
- Add keyboard shortcuts (optional)
- Test the command

**Time Required:** 10-15 minutes

**Difference from Actions:**

- **Actions** (robot button menu) - Best for quick selections, auto-generated workflows
- **Commands** (this guide) - Best for keyboard shortcuts, manual input prompts, command palette

---

## Prerequisites

- Extension running in development mode
- Basic understanding of TypeScript
- See [Quick Start](../../tutorials/vscode-extension-quickstart.md) for setup

---

## Step 1: Register Command in package.json

Edit the extension manifest to declare your command.

**File:** `.vscode/extensions/claude-mcp-vscode/package.json`

```json
{
  "contributes": {
    "commands": [
      {
        "command": "claude-mcp-vscode.quickCommit",
        "title": "Quick Commit with Manual Message",
        "icon": "$(git-commit)"
      }
    ],
    "keybindings": [
      {
        "command": "claude-mcp-vscode.quickCommit",
        "key": "ctrl+shift+c",
        "mac": "cmd+shift+c",
        "when": "scmProvider == git"
      }
    ]
  }
}
```

**Key fields:**

- `command`: Unique identifier (use extension name as prefix)
- `title`: Display name in Command Palette
- `icon`: VSCode icon (see [icon reference](https://code.visualstudio.com/api/references/icons-in-labels))
- `key`: Windows/Linux keyboard shortcut
- `mac`: macOS keyboard shortcut
- `when`: Optional context condition

---

## Step 2: Implement Command Handler

Add the command implementation to your extension code.

**File:** `.vscode/extensions/claude-mcp-vscode/src/extension.ts`

### 2a. Register the Command

```typescript
export function activate(context: vscode.ExtensionContext) {
    // ... existing code ...

    let quickCommit = vscode.commands.registerCommand(
        'claude-mcp-vscode.quickCommit',
        async () => {
            // Command implementation here
            const message = await vscode.window.showInputBox({
                prompt: 'Enter commit message',
                placeHolder: 'feat: add new feature'
            });

            if (!message) {
                return; // User cancelled
            }

            // Call MCP server
            const result = await callMCPServer(
                mcpServerPath,
                'git-commit',
                message
            );

            vscode.window.showInformationMessage(`Committed: ${result}`);
        }
    );

    context.subscriptions.push(quickCommit);
}
```

### 2b. Command with Multiple Prompts

```typescript
let quickCommit = vscode.commands.registerCommand(
    'claude-mcp-vscode.quickCommit',
    async () => {
        // Step 1: Get commit type
        const type = await vscode.window.showQuickPick(
            ['feat', 'fix', 'docs', 'refactor', 'test'],
            { placeHolder: 'Select commit type' }
        );

        if (!type) return;

        // Step 2: Get description
        const description = await vscode.window.showInputBox({
            prompt: 'Enter commit description',
            placeHolder: 'add user authentication'
        });

        if (!description) return;

        // Step 3: Build message and commit
        const message = `${type}: ${description}`;
        const result = await callMCPServer(mcpServerPath, 'git-commit', message);

        vscode.window.showInformationMessage(`Committed: ${result}`);
    }
);
```

---

## Step 3: Test the Command

### 3a. Reload Extension

In Extension Development Host:

- Press `Ctrl+R` (or `Cmd+R`)
- Or close and press F5 again

### 3b. Test via Command Palette

1. Press `Ctrl+Shift+P` (or `Cmd+Shift+P`)
2. Type your command title: "Quick Commit with Manual Message"
3. Press Enter
4. Follow the prompts

### 3c. Test via Keyboard Shortcut

1. Press `Ctrl+Shift+C` (or `Cmd+Shift+C`)
2. Follow the prompts

### 3d. Verify Results

- Check notification messages
- Check Debug Console for errors
- Verify the action completed (e.g., commit was created)

---

## Command Patterns

### Command with No Input

```typescript
let showStatus = vscode.commands.registerCommand(
    'claude-mcp-vscode.showStatus',
    async () => {
        const result = await callMCPServer(mcpServerPath, 'git-status', '');
        vscode.window.showInformationMessage(result);
    }
);
```

### Command with File Selection

```typescript
let processFile = vscode.commands.registerCommand(
    'claude-mcp-vscode.processFile',
    async () => {
        const uris = await vscode.window.showOpenDialog({
            canSelectFiles: true,
            canSelectFolders: false,
            canSelectMany: false
        });

        if (!uris || uris.length === 0) return;

        const filePath = uris[0].fsPath;
        const result = await callMCPServer(
            mcpServerPath,
            'process-file',
            filePath
        );

        vscode.window.showInformationMessage(result);
    }
);
```

### Command with Confirmation

```typescript
let dangerousAction = vscode.commands.registerCommand(
    'claude-mcp-vscode.dangerousAction',
    async () => {
        const confirm = await vscode.window.showWarningMessage(
            'This will delete all local changes. Continue?',
            { modal: true },
            'Yes, delete',
            'Cancel'
        );

        if (confirm !== 'Yes, delete') return;

        const result = await callMCPServer(mcpServerPath, 'reset-hard', '');
        vscode.window.showInformationMessage(result);
    }
);
```

---

## Auto-Generated vs Manual Input

**Understanding the difference:**

### Main Robot Button "Git Commit" (Auto-Generated)

- Reads staged changes automatically
- Generates commit message via MCP server analysis
- No user input required
- Best for quick workflows

### Command "Quick Commit" (Manual Input)

- Prompts user for commit message
- User types the message manually
- More control over message format
- Best for specific commit messages

**Example comparison:**

```typescript
// AUTO-GENERATED (robot button action)
// No input prompt - analyzes staged changes
const diff = await git.diff(['--staged']);
const result = await callMCPServer(mcpServerPath, 'git-commit', diff);
// Server generates message from diff

// MANUAL INPUT (command)
// Prompts for message - user types it
const message = await vscode.window.showInputBox({
    prompt: 'Enter commit message'
});
const result = await callMCPServer(mcpServerPath, 'git-commit', message);
// Server uses message as-is
```

---

## Best Practices

1. **Use meaningful command IDs** - Prefix with extension name
2. **Provide clear prompts** - Users should know what to enter
3. **Handle cancellation** - Check if user pressed ESC
4. **Add keyboard shortcuts** - For frequently used commands
5. **Use `when` clauses** - Prevent shortcuts from conflicting
6. **Validate input** - Check for empty or invalid values

---

## Troubleshooting

**Problem:** Command doesn't appear in Command Palette

**Solution:**

- Check package.json syntax (valid JSON)
- Reload extension (Ctrl+R)
- Verify command ID matches registration

---

**Problem:** Keyboard shortcut doesn't work

**Solution:**

- Check for conflicting shortcuts (File → Preferences → Keyboard Shortcuts)
- Verify `when` clause is satisfied
- Try different key combination

---

**Problem:** Command crashes with error

**Solution:**

- Check Debug Console for stack trace
- Ensure all async operations use `await`
- Handle null/undefined cases for user input

---

## Related Documentation

- **Add Actions**: [Add a New Action](add-action.md) - For robot button menu items
- **Architecture**: [VS Code Extension Architecture](../../explanation/vscode-extension-architecture.md)
- **Reference**: [VS Code Extension Reference](../../reference/vscode-extension.md)
- **VSCode API**: [Commands API](https://code.visualstudio.com/api/extension-guides/command)
