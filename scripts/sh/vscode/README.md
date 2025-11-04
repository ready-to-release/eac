# VSCode Extension Automation Scripts

This directory contains shell scripts to automate common tasks for the VSCode extension and MCP servers.

## Available Scripts

### `init.sh`

**Purpose**: Initialize the project for first-time setup.

**What it does**:
- Checks prerequisites (Go, Node.js, npm)
- Installs VSCode extension dependencies
- Compiles the TypeScript extension
- Verifies all MCP servers are ready

**Usage**:
```bash
./automation/sh/vscode/init.sh
```

**When to use**:
- First time setting up the project
- After cloning the repository
- When you want to verify everything is properly configured

---

### `restore.sh`

**Purpose**: Restore dependencies with a fresh installation.

**What it does**:
- Removes existing `node_modules`
- Removes `package-lock.json`
- Removes `out` directory
- Installs fresh dependencies
- Recompiles the extension

**Usage**:
```bash
./automation/sh/vscode/restore.sh
```

**When to use**:
- When you have dependency issues
- After updating `package.json`
- When npm dependencies are corrupted
- To fix "module not found" errors

---

### `clean.sh`

**Purpose**: Clean all build artifacts and dependencies.

**What it does**:
- Removes `node_modules` from extension
- Removes compiled `out` directory
- Removes `package-lock.json`
- Removes `.vscode-test` directory
- Removes packaged `*.vsix` files
- Removes MCP server binaries (*.exe, *.test)
- Removes TypeScript build info

**Usage**:
```bash
./automation/sh/vscode/clean.sh
```

**When to use**:
- Before committing to ensure no artifacts are tracked
- When you want a completely fresh build
- To free up disk space
- Before running `init.sh` for a clean setup

---

## Typical Workflows

### First Time Setup
```bash
# Clone the repo, then:
./automation/sh/vscode/init.sh
```

### Fix Dependency Issues
```bash
./automation/sh/vscode/restore.sh
```

### Complete Reset
```bash
./automation/sh/vscode/clean.sh
./automation/sh/vscode/init.sh
```

### Before Committing
```bash
# Make sure no build artifacts are included
./automation/sh/vscode/clean.sh
git status
```

---

## Requirements

All scripts require:
- **Bash** shell (Git Bash on Windows, native on Linux/macOS)
- **Go** 1.21+ (for MCP servers)
- **Node.js** 18.x+ (for VSCode extension)
- **npm** (comes with Node.js)

## Permissions

All scripts are executable (`chmod +x`). If you encounter permission errors:

```bash
chmod +x automation/sh/vscode/*.sh
```

## Exit Codes

All scripts use `set -e` to exit immediately on error. If a script fails:
1. Read the error message
2. Fix the issue (e.g., install missing prerequisites)
3. Re-run the script

## Troubleshooting

### "command not found: go"
- Install Go from https://golang.org/dl/

### "command not found: node"
- Install Node.js from https://nodejs.org/

### "npm install" fails
- Check your network connection
- Try running `./automation/sh/vscode/clean.sh` first
- Delete `package-lock.json` manually and retry

### Scripts won't run on Windows
- Use Git Bash (comes with Git for Windows)
- Or use WSL (Windows Subsystem for Linux)

## Notes

- Scripts are designed to be idempotent (safe to run multiple times)
- All scripts output progress messages with ✓ and ❌ symbols
- Scripts automatically detect the project root directory
- Safe to run from any location in the repository
