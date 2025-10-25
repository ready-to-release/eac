# Claude Code Logs

This directory is for Claude Code log files.

## Configuration

Logs are stored in this directory to keep them organized and separate from code.

### Structure

```text
.claude/logs/
├── .gitignore         # Ignores all log files except .gitkeep and .gitignore
├── .gitkeep           # Ensures directory is tracked by git
├── README.md          # This file
└── (log files here)   # Ignored by git
```

## Git Ignore

The `.gitignore` in this directory ensures that:

- ✅ Log files are NOT committed to git
- ✅ The directory structure IS tracked
- ✅ `.gitignore` and `.gitkeep` are tracked

Pattern:

```gitignore
# Ignore all files
*

# Except these
!.gitignore
!.gitkeep
!README.md
```

## Log Files

Claude Code may generate various log files here:

- Debug logs
- Error logs
- Session logs
- MCP communication logs

All log files are automatically ignored by git for privacy and to avoid bloating the repository.

## Viewing Logs

```bash
# List all logs
ls -la .claude/logs/

# View latest log
tail -f .claude/logs/latest.log

# Search logs
grep "error" .claude/logs/*.log
```

## Cleaning Logs

```bash
# Remove all logs (keeps .gitignore and .gitkeep)
cd .claude/logs
find . -type f ! -name '.gitignore' ! -name '.gitkeep' ! -name 'README.md' -delete

# Or with age filter (logs older than 7 days)
find . -type f ! -name '.gitignore' ! -name '.gitkeep' ! -name 'README.md' -mtime +7 -delete
```

## Notes

- Logs are local and not shared
- Safe to delete at any time
- Regenerated as needed by Claude Code
- Useful for debugging issues
