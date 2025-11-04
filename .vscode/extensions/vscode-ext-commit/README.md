# Commit Message AI - VSCode Extension

A VSCode extension that generates AI-powered semantic commit messages using the 7-lever commit-ai system.

## Features

- Adds a button to the Git source control view (SCM)
- Generates commit messages using `commit-ai` command
- Automatic validation and cleanup
- Self-healing via auto-fix
- Contract-driven quality assurance
- Real-time progress updates with whimsical messages

## Installation

1. Install dependencies:

   ```bash
   npm install
   ```

2. Compile the extension:

   ```bash
   npm run compile
   ```

3. Press F5 in VSCode to open a new window with the extension loaded

## Usage

1. Stage your changes in Git
2. Open the Source Control view in VSCode (Ctrl+Shift+G)
3. Click the robot icon button in the toolbar
4. Wait for the AI to generate your commit message
5. Review and commit!

## Requirements

- VSCode 1.80.0 or higher
- Go installed (for running commit-ai)
- The `commit-ai` command must be available at `src/commands/` in your workspace
- Claude Code authentication (subscription)

## Development

To watch for changes and recompile automatically:

```bash
npm run watch
```

## Extension Settings

This extension doesn't require any specific settings.

## Known Issues

None at this time.

## How It Works

The extension calls the `commit-ai` command which orchestrates a 7-lever system:

1. **LEVER 1**: Contract specification (formal rules)
2. **LEVER 2**: Agent instructions (AI guidance)
3. **LEVER 3**: Command orchestration
4. **LEVER 3.1**: Pre-verification cleanup
5. **LEVER 4**: Contract verification
6. **LEVER 5**: Feedback loop (AI auto-fix)
7. **LEVER 5.1**: Pre-re-verification cleanup

See [commit-ai documentation](../../../docs/reference/commands/commit-ai.md) for details.

## Release Notes

### 0.2.0

Simplified architecture - now calls `commit-ai` command directly instead of using MCP server.

### 0.1.0

Initial release with MCP server integration (deprecated).
