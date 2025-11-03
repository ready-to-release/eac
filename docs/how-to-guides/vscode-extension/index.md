# VS Code Extension How-to Guides

**Problem-solving guides**: Step-by-step instructions for specific development tasks with the VS Code extension.

---

## Available Guides

### [Add a New Action](add-action.md)

Learn how to add a new action to the robot button menu in the extension.

**What you'll do:**

- Update Quick Pick menu in extension.ts
- Implement action handler in MCP server
- Test the new action

**Time**: 10-15 minutes

---

### [Add a New Command](add-command.md)

Learn how to register a VSCode command with keyboard shortcuts.

**What you'll do:**

- Register command in package.json
- Implement command handler with input prompts
- Add keyboard shortcuts
- Test via Command Palette and shortcuts

**Time**: 10-15 minutes

---

### [Work with MCP Servers](work-with-mcp-servers.md)

Learn how to test, debug, and create MCP servers.

**What you'll do:**

- Test servers manually with JSON-RPC
- Debug server communication
- Create a new MCP server from scratch
- Configure servers in .mcp.json

**Time**: 20-30 minutes

---

### [Troubleshoot Common Issues](troubleshoot.md)

Solutions for common problems when developing or using the extension.

**Covers:**

- Extension button not visible
- MCP server errors
- Build errors
- Extension not reloading
- Commit message not generated
- Keyboard shortcuts not working
- Action execution failures
- Performance issues

---

### [Customize Commit Agents](customize-commit-agents.md)

Learn how to modify and extend the commit agent pipeline.

**What you'll do:**

- Test individual agents
- Modify agent prompts
- Add a new agent to the pipeline
- Change agent models
- Debug agent issues

**Time**: 15-30 minutes

---

## What are How-to Guides?

How-to guides are **task-oriented** instructions that help you accomplish specific goals. They assume you have basic knowledge and want to solve a particular problem.

**Characteristics:**

- Focus on a specific task or problem
- Step-by-step instructions
- Practical, actionable content
- Assumes basic familiarity with the system

**Not finding what you need?**

- **Just getting started?** See [Quick Start Tutorial](../../tutorials/vscode-extension-quickstart.md)
- **Want to understand how it works?** See [Extension Architecture](../../explanation/vscode-extension-architecture.md)
- **Need technical reference?** See [VS Code Extension Reference](../../reference/vscode-extension.md)
