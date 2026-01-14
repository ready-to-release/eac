# CLI Docs

```text
description: "Update CLI documentation and help text"
```

You are updating CLI documentation and help text.

## Process

1. **Identify what changed**:
   - New commands or flags added?
   - Command behavior modified?
   - Use MCP `show-files-changed` to see modified files

2. **Update command help text**:
   - Delegate to go-cli-ux agent if needed
   - Use Task tool with subagent_type="go-cli-ux"
   - For eac/commands: Update command registration and help
   - For r2r/cli: Update Cobra command Short/Long descriptions
   - Add usage examples
   - Document all flags with descriptions

3. **Update how-to guides** (if CLI surface changed):
   - Check `docs/how-to-guides/eac/commands/`
   - Add new guide for new commands
   - Update existing guides if behavior changed
   - Follow existing guide structure

4. **Validate documentation**:
   - Run the command with `--help` to verify output
   - Ensure examples work
   - Check that MkDocs site builds (if documentation in docs/)

## Example Usage

User: `/go:cli-docs update docs for the new validate-config command`
