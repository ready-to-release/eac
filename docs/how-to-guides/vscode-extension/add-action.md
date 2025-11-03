# Add a New Action

Learn how to add a new action to the robot button menu in the VS Code extension.

---

## Overview

**What you'll do:**

- Add a new action to the Quick Pick menu in the extension
- Implement the action handler in the MCP server
- Test the new action

**Time Required:** 10-15 minutes

---

## Prerequisites

- Extension running in development mode
- Basic understanding of TypeScript and Go
- See [Quick Start](../../tutorials/vscode-extension-quickstart.md) for setup

---

## Step 1: Update Quick Pick Menu

Edit the extension code to add your new action to the menu.

**File:** `.vscode/extensions/claude-mcp-vscode/src/extension.ts`

```typescript
const action = await vscode.window.showQuickPick([
    { label: 'Git Commit', value: 'git-commit' },
    { label: 'Git Push', value: 'git-push' },
    { label: 'Git Pull', value: 'git-pull' },
    { label: 'Deploy Staging', value: 'deploy-staging' },  // New action!
], {
    placeHolder: 'Select an action'
});
```

**Key points:**

- `label`: What the user sees in the menu
- `value`: Internal identifier sent to MCP server

---

## Step 2: Implement Action Handler in MCP Server

Add the server-side logic to handle your new action.

**File:** `src/mcp/vscode/main.go`

### 2a. Add Case to Switch Statement

```go
func executeAction(action string, message string) string {
    switch action {
    case "git-commit":
        return handleGitCommit(message)
    case "git-push":
        return handleGitPush()
    case "git-pull":
        return handleGitPull()
    case "deploy-staging":
        return handleDeployStaging(message)  // New handler!
    default:
        return fmt.Sprintf("Executing: %s", action)
    }
}
```

### 2b. Implement Handler Function

```go
func handleDeployStaging(env string) string {
    // Your deployment logic here

    // Example: Run deployment script
    cmd := exec.Command("bash", "./deploy.sh", "staging")
    output, err := cmd.CombinedOutput()

    if err != nil {
        return fmt.Sprintf("Deployment failed: %s", err)
    }

    return fmt.Sprintf("Successfully deployed to staging: %s", string(output))
}
```

**Tips:**

- Return user-friendly messages
- Handle errors gracefully
- Keep handlers focused on one task

---

## Step 3: Test the New Action

### 3a. Reload the Extension

In the Extension Development Host window:

- Press `Ctrl+R` (or `Cmd+R` on Mac)
- Or close and press F5 again in original window

### 3b. Use the New Action

1. Click the robot button (🤖)
2. See your new action in the list: "Deploy Staging"
3. Click it
4. Verify the result notification

### 3c. Check for Errors

If something goes wrong:

- Check **Debug Console** in original VSCode window
- Look for stderr output from MCP server
- Add logging to your handler:

```go
func handleDeployStaging(env string) string {
    fmt.Fprintf(os.Stderr, "Starting deployment to: %s\n", env)
    // ... rest of function
}
```

---

## Common Patterns

### Action That Reads Git State

```go
func handleCheckBranch() string {
    cmd := exec.Command("git", "branch", "--show-current")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Sprintf("Error: %s", err)
    }
    return fmt.Sprintf("Current branch: %s", strings.TrimSpace(string(output)))
}
```

### Action That Prompts for Input

**Note:** For actions requiring input, consider creating a [VSCode Command](add-command.md) instead, which supports input prompts.

### Action That Runs Multiple Steps

```go
func handleRunTests() string {
    steps := []struct {
        name string
        cmd  *exec.Cmd
    }{
        {"Lint", exec.Command("npm", "run", "lint")},
        {"Test", exec.Command("npm", "test")},
        {"Build", exec.Command("npm", "run", "build")},
    }

    for _, step := range steps {
        if output, err := step.cmd.CombinedOutput(); err != nil {
            return fmt.Sprintf("%s failed: %s", step.name, string(output))
        }
    }

    return "All checks passed ✓"
}
```

---

## Best Practices

1. **Keep actions simple** - One clear purpose per action
2. **Return informative messages** - Users should know what happened
3. **Handle errors gracefully** - Don't crash, return error messages
4. **Use descriptive labels** - Make it clear what the action does
5. **Test thoroughly** - Try success cases, error cases, edge cases

---

## Troubleshooting

**Problem:** New action doesn't appear in menu

**Solution:**

- Reload extension (Ctrl+R)
- Check for syntax errors in extension.ts
- Verify you saved the file

---

**Problem:** Action appears but does nothing

**Solution:**

- Check Debug Console for errors
- Verify the `value` in extension.ts matches the `case` in main.go
- Add logging to see if handler is called

---

**Problem:** MCP server crashes when action runs

**Solution:**

- Check for nil pointer errors in Go
- Ensure external commands exist (e.g., `git`, `bash`)
- Add error handling for all operations

---

## Related Documentation

- **Add Commands**: [Add a New Command](add-command.md) - For actions requiring input prompts
- **MCP Servers**: [Work with MCP Servers](work-with-mcp-servers.md) - Testing and debugging servers
- **Architecture**: [VS Code Extension Architecture](../../explanation/vscode-extension-architecture.md)
- **Reference**: [VS Code Extension Reference](../../reference/vscode-extension.md)
