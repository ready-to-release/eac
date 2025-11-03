# Troubleshoot Common Issues

Solutions for common problems when developing or using the VS Code extension.

---

## Extension Button Not Visible

### Problem

The robot button (🤖) doesn't appear in the Source Control toolbar.

### Solution 1: Verify Git Repository

```bash
git status
```

If you see "not a git repository":

```bash
git init
```

**Why:** The button only appears in Git repositories (`scmProvider == git` condition in package.json).

### Solution 2: Reload Extension

In Extension Development Host:

- Press `Ctrl+R` (or `Cmd+R` on Mac)
- Or close the window and press F5 again in the original window

### Solution 3: Check Extension Activation

Open Debug Console in the original VSCode window and look for:

```text
Extension 'claude-mcp-vscode' is now active!
```

If you don't see this:

- Check for errors in Debug Console
- Verify package.json has correct `activationEvents`

---

## MCP Server Errors

### Problem

Extension shows errors like "Failed to execute action" or "Server not responding".

### Solution 1: Test Server Manually

```bash
cd src/mcp/vscode
go run .
```

Send test command:

```json
{"jsonrpc":"2.0","id":1,"method":"initialize"}
```

If server doesn't respond or crashes:

- Check for compilation errors: `go build`
- Review server logs in terminal
- Fix syntax errors in main.go

### Solution 2: Check Go Installation

```bash
go version  # Should be 1.21.0 or higher
```

If Go isn't installed or version is too old:

- Install from [golang.org](https://golang.org/dl/)
- Update PATH environment variable

### Solution 3: View Error Details

Check Debug Console in original VSCode window for stderr output:

```text
[MCP Server stderr] Error: failed to execute git command
[MCP Server stderr] exit status 1
```

Add more logging to server:

```go
func executeAction(action string, message string) string {
    fmt.Fprintf(os.Stderr, "[DEBUG] Action: %s, Message: %s\n", action, message)
    // ... rest of function
}
```

### Solution 4: Verify .mcp.json Configuration

**File:** `.mcp.json`

```json
{
  "mcpServers": {
    "vscode": {
      "command": "bash",
      "args": ["src/mcp/vscode/run.sh"]
    }
  }
}
```

**Check:**

- Path to run.sh is correct
- run.sh has execute permissions: `chmod +x src/mcp/vscode/run.sh`
- run.sh changes to correct directory

---

## Build Errors

### Problem

Extension fails to compile with TypeScript errors or npm errors.

### Solution 1: Clean and Restore

```bash
cd .vscode/extensions/claude-mcp-vscode
rm -rf node_modules out
npm install
npm run compile
```

Or use automation:

```bash
./automation/sh/vscode/restore.sh
```

### Solution 2: Check Node.js Version

```bash
node --version  # Should be 18.0.0 or higher
npm --version
```

If versions are too old:

- Install latest LTS from [nodejs.org](https://nodejs.org/)
- Restart terminal after installation

### Solution 3: Fix TypeScript Errors

Open the file with errors and check:

- Missing imports
- Type mismatches
- Undefined variables

Example fix:

```typescript
// Before (error: Cannot find name 'vscode')
const editor = vscode.window.activeTextEditor;

// After (add import)
import * as vscode from 'vscode';
const editor = vscode.window.activeTextEditor;
```

---

## Extension Not Reloading

### Problem

Changes to extension code don't take effect after editing.

### Solution 1: Hard Reload

1. Close Extension Development Host window
2. Return to original VSCode window
3. Press F5 to launch again

### Solution 2: Use Watch Mode

Auto-recompile on file save:

```bash
cd .vscode/extensions/claude-mcp-vscode
npm run watch
```

Keep this running in a terminal. Now:

- Edit extension.ts
- Save file
- Watch mode recompiles automatically
- Press Ctrl+R to reload

### Solution 3: Clear VSCode Cache

If issues persist:

```bash
# Close VSCode completely
# Then clear extension cache
rm -rf ~/.vscode/extensions/.obsolete
rm -rf .vscode/extensions/claude-mcp-vscode/out
```

Recompile and launch:

```bash
npm run compile
# Press F5 in VSCode
```

---

## Commit Message Not Generated

### Problem

Clicking "Git Commit" doesn't auto-generate a message, or generates empty message.

### Solution 1: Verify Changes Are Staged

**Critical:** The extension only analyzes **staged** changes.

```bash
git status
```

You should see:

```text
Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
        modified:   file1.txt
        modified:   file2.txt
```

If you see "Changes not staged for commit":

- Stage files in Source Control (click + icon)
- Or run: `git add <file>`

### Solution 2: Check MCP Server Response

Add logging to see what the server receives:

**File:** `src/mcp/vscode/main.go`

```go
func handleGitCommit(message string) string {
    fmt.Fprintf(os.Stderr, "[DEBUG] Received diff length: %d\n", len(message))
    fmt.Fprintf(os.Stderr, "[DEBUG] First 100 chars: %s\n", message[:min(100, len(message))])
    // ... rest of function
}
```

Check Debug Console for this output.

### Solution 3: Test Message Generation Manually

```bash
cd src/mcp/vscode
go run .
```

Send commit request with real diff:

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vscode-action","arguments":{"action":"git-commit","diff":"diff --git a/file.txt b/file.txt\n..."}}}
```

Verify response contains a message.

---

## Keyboard Shortcuts Not Working

### Problem

Custom keyboard shortcuts (e.g., Ctrl+Shift+C) don't trigger commands.

### Solution 1: Check for Conflicts

File → Preferences → Keyboard Shortcuts (or `Ctrl+K Ctrl+S`)

Search for your key combination (e.g., "ctrl+shift+c") and look for conflicts.

If another command uses the same shortcut:

- Remove the conflicting binding
- Or use a different key combination

### Solution 2: Verify `when` Clause

Check package.json:

```json
{
  "keybindings": [
    {
      "command": "claude-mcp-vscode.quickCommit",
      "key": "ctrl+shift+c",
      "when": "scmProvider == git"
    }
  ]
}
```

The `when` clause must be true for the shortcut to work:

- `scmProvider == git` - Only works in Git repositories
- `editorFocus` - Only works when editor has focus

Test without `when` clause to isolate the issue.

### Solution 3: Restart VSCode

Some keybinding changes require a full restart:

1. Close all VSCode windows
2. Reopen VSCode
3. Test shortcut again

---

## Action Execution Fails Silently

### Problem

Action runs but produces no result or error message.

### Solution 1: Add Error Handling

**File:** `src/mcp/vscode/main.go`

```go
func handleGitCommit(message string) string {
    cmd := exec.Command("git", "commit", "-m", message)
    output, err := cmd.CombinedOutput()

    // Log everything
    fmt.Fprintf(os.Stderr, "[DEBUG] Command: git commit -m %s\n", message)
    fmt.Fprintf(os.Stderr, "[DEBUG] Output: %s\n", string(output))

    if err != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
        return fmt.Sprintf("Commit failed: %s", string(output))
    }

    return fmt.Sprintf("Committed: %s", message)
}
```

### Solution 2: Check Return Values

Ensure your handler returns a string:

```go
// Bad - returns empty string
func handleAction() string {
    doSomething()
    return ""  // Extension shows nothing!
}

// Good - returns informative message
func handleAction() string {
    result := doSomething()
    return fmt.Sprintf("Action completed: %v", result)
}
```

### Solution 3: Verify Extension Displays Result

**File:** `.vscode/extensions/claude-mcp-vscode/src/extension.ts`

```typescript
const result = await callMCPServer(mcpServerPath, action, message);

// Add logging
console.log('MCP Server result:', result);

// Ensure notification is shown
if (result) {
    vscode.window.showInformationMessage(result);
} else {
    vscode.window.showWarningMessage('Action completed with no output');
}
```

---

## Performance Issues

### Problem

Extension is slow or VSCode becomes unresponsive.

### Solution 1: Don't Block the UI Thread

Use async/await properly:

```typescript
// Bad - blocks UI
const result = syncCall();

// Good - doesn't block UI
const result = await asyncCall();
```

### Solution 2: Limit MCP Server Lifetime

Ensure servers exit after each request:

```go
func main() {
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        line := scanner.Text()
        handleRequest(line)
        // Process one request and exit (extension will spawn again if needed)
        return
    }
}
```

### Solution 3: Reduce Logging

In production, minimize stderr output:

```go
const DEBUG = false  // Set to false for production

func log(format string, args ...interface{}) {
    if DEBUG {
        fmt.Fprintf(os.Stderr, format, args...)
    }
}
```

---

## Related Documentation

- **Quick Start**: [VS Code Extension Quick Start](../../tutorials/vscode-extension-quickstart.md)
- **MCP Servers**: [Work with MCP Servers](work-with-mcp-servers.md)
- **Add Actions**: [Add a New Action](add-action.md)
- **Architecture**: [VS Code Extension Architecture](../../explanation/vscode-extension-architecture.md)
- **Reference**: [VS Code Extension Reference](../../reference/vscode-extension.md)

---

## Still Having Issues?

1. Check Debug Console for detailed error messages
2. Review MCP server logs (stderr output)
3. Test MCP server manually to isolate the problem
4. Verify all prerequisites are installed and up-to-date
5. Check GitHub issues for similar problems
