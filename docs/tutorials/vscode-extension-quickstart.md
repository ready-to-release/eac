# VS Code Extension Quick Start

Get the VSCode extension up and running in 5 minutes with AI-generated commit messages.

## What You'll Learn

- Set up and run the extension in 5 minutes
- Use the robot button for Git operations with auto-generated commit messages
- Make your first commit with AI-generated message

**Time Required:** 5 minutes

---

## Prerequisites

- VSCode 1.80.0+
- Git repository
- Go 1.21+ and Node.js 18+

---

## Quick Start (5 Minutes)

### Step 1: Initialize (2 minutes)

```bash
./automation/sh/vscode/init.sh
```

Wait for "✓ Initialization Complete!"

### Step 2: Launch Extension (1 minute)

```bash
code .    # Open VSCode
# Press F5 to launch Extension Development Host
```

### Step 3: Test It (2 minutes)

In the Extension Development Host window:

1. **Make some changes**: Edit a file and save it
2. **Open Source Control**: `Ctrl+Shift+G` (or `Cmd+Shift+G` on Mac)
3. **Stage your changes**: Click the `+` icon next to changed files
4. **Click the Robot Icon** (🤖) in the toolbar
5. **Select "Git Commit"**
6. **See the result**: Extension auto-generates the commit message and commits

**Done!** You now have a working MCP-integrated extension with AI-generated commit messages.

---

## Using the Extension

### Available Actions

Click the robot button to see:

| Action | Description | Auto-generates Message |
|--------|-------------|------------------------|
| **Git Commit** | Commits staged changes | Yes - analyzes staged files |
| **Git Push** | Push to remote | N/A |
| **Git Pull** | Pull from remote | N/A |
| **Custom Action** | Run custom command | No - prompts for input |

### Quick Commit Workflow

```text
1. Make changes to files
2. Open Source Control (Ctrl+Shift+G)
3. Stage your changes (click + icon)
4. Click robot icon → "Git Commit"
5. Extension analyzes staged changes and generates commit message
6. Commit is created automatically
```

**Important**: The extension analyzes only **staged** changes. Make sure to stage files before committing.

### Viewing Results

- **Notifications**: Success/error messages with auto-generated commit message
- **Debug Console**: Original VSCode window → Debug Console
- **Output Panel**: View → Output → Select extension output

---

## Next Steps

Now that you have the extension running, you can:

1. **Understand the architecture**: See [VS Code Extension Architecture](../explanation/vscode-extension-architecture.md)
2. **Add custom actions**: See [Add a New Action](../how-to-guides/vscode-extension/add-action.md)
3. **Work with MCP servers**: See [Work with MCP Servers](../how-to-guides/vscode-extension/work-with-mcp-servers.md)
4. **Troubleshoot issues**: See [Troubleshooting Guide](../how-to-guides/vscode-extension/troubleshoot.md)
5. **Reference documentation**: See [VS Code Extension Reference](../reference/vscode-extension.md)

---

## Need Help?

- **Troubleshooting**: See [Troubleshooting Guide](../how-to-guides/vscode-extension/troubleshoot.md)
- **How-to Guides**: See [VS Code Extension How-to Guides](../how-to-guides/vscode-extension/)
- **Technical Details**: See [Reference Documentation](../reference/vscode-extension.md)
