# Quick Start Guide

Get up and running in 5 minutes!

> 🎬 **Video Demo**: See [docs/assets/quick-start-guide.md](docs/assets/quick-start-guide.md) for a visual walkthrough (GIF to be recorded)

## 1. Initialize (2 minutes)

```bash
# Run the setup script
./automation/sh-vscode/init.sh
```

Wait for "✓ Initialization Complete!"

## 2. Launch Extension (1 minute)

```bash
# Open VSCode
code .
```

Press **F5** to launch Extension Development Host

## 3. Test It (2 minutes)

In the new VSCode window:

1. **Open Source Control**
   - Press `Ctrl+Shift+G` (or `Cmd+Shift+G` on Mac)

2. **Click the Robot Icon** 🤖
   - It's in the toolbar at the top of the Source Control view

3. **Select "Git Commit"**
   - From the menu that appears

4. **Enter a message**
   - Type: "Test commit"
   - Press Enter

5. **See the result!**
   - A notification shows the MCP server response

## Done! 🎉

You now have:

- ✓ 4 MCP servers (pwsh, docs, github, vscode)
- ✓ VSCode extension with Git button
- ✓ MCP integration working

## What's Next?

- **Full guide**: See [USAGE.md](USAGE.md) for detailed instructions
- **Automation**: See [automation/sh-vscode/README.md](automation/sh-vscode/README.md) for scripts
- **Development**: See [README.md](README.md) for architecture and development

## Common Commands

```bash
# Restore dependencies (if issues)
./automation/sh-vscode/restore.sh

# Clean build artifacts
./automation/sh-vscode/clean.sh

# Rebuild everything
./automation/sh-vscode/clean.sh && ./automation/sh-vscode/init.sh
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Robot icon not visible | Make sure you're in a Git repository (`git init`) |
| "Go not found" | Install Go from <https://golang.org/dl/> |
| "Node not found" | Install Node.js from <https://nodejs.org/> |
| Build errors | Run `./automation/sh-vscode/restore.sh` |

## Quick Reference

**Extension Location**: `.vscode/extensions/claude-mcp-vscode/`

**MCP Servers**: `src/mcp/[pwsh|docs|github|vscode]/`

**Scripts**: `automation/sh-vscode/[init|restore|clean].sh`

**Reload Extension**: `Ctrl+R` in Extension Development Host

**View Logs**: Debug Console in original VSCode window
