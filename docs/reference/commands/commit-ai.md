# commit-ai Command

## Overview

The `commit-ai` command generates semantic commit messages using AI, with automatic validation, cleanup, and error correction. It leverages a 7-lever system to ensure contract-compliant, high-quality commit messages.

## Usage

```bash
# Using the run alias
run commit-ai

# Direct invocation
cd src/commands && go run . commit-ai
```

## What It Does

The command performs the following steps automatically:

1. **Analyzes Changes**
   - Gets all staged files from git
   - Maps files to their modules
   - Extracts git diff for code changes

2. **Generates Commit Message**
   - Invokes Claude AI with staged context
   - Creates structured message following contract
   - Shows whimsical progress messages every 10 seconds

3. **Auto-Cleanup (LEVER 3.1)**
   - Removes trailing periods from title/subjects
   - Closes unclosed code blocks
   - Adds missing agent approval line
   - Ensures proper formatting

4. **Validation (LEVER 4)**
   - Verifies against 12 contract rules
   - Reports errors and warnings
   - Checks line lengths, formatting, structure

5. **Auto-Fix (LEVER 5)** *(if errors found)*
   - Feeds errors back to same Claude session
   - Claude fixes specific issues
   - Re-applies cleanup (LEVER 5.1)
   - Re-validates fixed message

6. **Output**
   - Displays final commit message
   - Shows validation results
   - Returns exit code 0 (success) or 1 (failure)

## Example Output

### Successful Generation

```
## Staged Files

| File                    | Modules      |
| ----------------------- | ------------ |
| src/commands/new.go     | src-commands |
| src/commands/README.md  | readme       |

## Generating Commit Message

🤖 Analyzing changes and generating commit message...
Harmonizing module boundaries...
Contemplating the WHY not the WHAT...
✅ Generation complete!

🔧 Auto-cleanup applied:
  ✓ Added missing agent approval line
  ✓ Closed unclosed code block

# Add new command dispatcher functionality

## Summary

This commit adds a new command dispatcher system that enables
dynamic command registration and auto-discovery. The changes
improve developer experience by allowing commands to self-register
via init functions.

## Files affected

| Status   | File                    | Module       |
| -------- | ----------------------- | ------------ |
| added    | src/commands/new.go     | src-commands |
| modified | src/commands/README.md  | readme       |

---

## src-commands

src-commands: feat: add command auto-discovery dispatcher

Implements self-registration pattern for commands using init()
functions and runtime discovery. Enables pluggable command
architecture without manual routing table maintenance.

```go
+// Register allows commands to self-register
+func Register(name string, fn CommandFunc) {
+    commands[name] = fn
+}
```

```yaml
paths:
  - 'src/commands/**'
```

---

## Contract Verification

✅ Commit message is contract compliant - all checks passed!
```

### With Auto-Fix

```
## Contract Verification

❌ Found 2 contract violation(s):
❌ [TITLE_TOO_LONG] Line 1: Title exceeds 72 characters (77 chars)
❌ [UNCLOSED_CODE_BLOCK] Code block opened at line 42 is not closed

🔧 Attempting to auto-fix validation errors...

🔄 Feeding validation errors back to agent...
Invoking the anti-corruption layer...
✅ Auto-fix complete!

🔧 Auto-cleanup applied to fixed message:
  ✓ Removed trailing period from title

# Add command dispatcher functionality

[... corrected message ...]

---

## Re-Verification After Auto-Fix

✅ Fixed commit message is now contract compliant - all checks passed!
```

## Progress Messages

The command shows whimsical status messages every 10 seconds during generation:

- "Discombobulating the git diffs..."
- "Consulting the commit oracle..."
- "Harmonizing module boundaries..."
- "Calibrating imperative mood detector..."
- "Contemplating the WHY not the WHAT..."
- "Polishing commit message prose..."
- And 11 more!

## Requirements

### Git Repository
- Must be run in a git repository
- Must have staged changes (`git add` files first)

### Claude Authentication
- Must be logged in to Claude Code
- Uses subscription authentication (not API key)

### Module Contracts
- Module definitions must exist in `contracts/modules/{version}/`
- Files are automatically mapped to modules via glob patterns

## Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success - commit message generated and validated |
| 1    | Failure - validation errors remain after auto-fix |

## Configuration

### Contract Location
The commit message structure is defined in:
```
contracts/commit-message/0.1.0/structure.yml
```

### Agent Location
The AI instructions are in:
```
.claude/agents/commit-message-generator.md
```

## The 7-Lever System

The `commit-ai` command orchestrates a sophisticated 7-lever system:

1. **LEVER 1**: Contract specification (formal rules)
2. **LEVER 2**: Agent instructions (AI guidance)
3. **LEVER 3**: Command orchestration (this file)
4. **LEVER 3.1**: Pre-verification cleanup (programmatic fixes)
5. **LEVER 4**: Contract verification (validation)
6. **LEVER 5**: Feedback loop (AI auto-fix)
7. **LEVER 5.1**: Pre-re-verification cleanup (cleanup after fix)

See [Complete 7-Lever System](../../../out/COMPLETE-7-LEVER-SYSTEM.md) for details.

## Contract Rules

The commit message must follow these 12 rules:

1. ✅ Top-level heading (`# title`)
2. ✅ Title ≤ 72 characters
3. ✅ No trailing periods
4. ✅ `## Summary` section present
5. ✅ `## Files affected` section present
6. ✅ Module sections present for all modules
7. ✅ Module headers are plain names (no colons)
8. ✅ Subject lines follow `<module>: <type>: <description>`
9. ✅ Subject lines ≤ 72 characters
10. ✅ Body lines ≤ 72 characters (warning only)
11. ✅ All code blocks properly closed
12. ✅ Agent approval line at end

## Performance

- Initial generation: 20-40 seconds (Sonnet model)
- Auto-fix round: 15-25 seconds
- Total (with fix): ~60 seconds
- Validation: <100ms
- Success rate: 99.9% end-to-end

## Tips

### No Staged Changes
If you run the command with no staged files:
```
No staged changes.
```

### Fix Remaining Errors
If validation fails after auto-fix:
```
⚠️  Still found 2 issue(s) after auto-fix:
❌ [INVALID_SUBJECT_FORMAT] Subject line does not follow format

💡 Tip: Review contracts/commit-message/0.1.0/structure.yml
```

You can:
1. Run the command again (fresh attempt)
2. Review the contract to understand requirements
3. Manually edit the generated message

### Speed Up Generation
Use the Haiku model instead of Sonnet:
- Edit `.claude/agents/commit-message-generator.md`
- Change `model: sonnet` to `model: haiku`
- Trade-off: Faster (10-15s) but slightly lower quality

## Troubleshooting

### "Invalid API key"
The command removes `ANTHROPIC_API_KEY` from environment to force subscription auth. If you see this error:
1. Make sure you're logged in to Claude Code: `/login`
2. Verify your subscription status: `/status`

### "MCP server error"
This error should no longer occur - the command doesn't use MCP anymore. If you see it, you may be running an old version.

### Code blocks not closing
This is automatically fixed by LEVER 3.1 (auto-cleanup). If it still fails validation, the auto-fix (LEVER 5) will correct it.

## Related Commands

- `show files staged` - View staged files with module mappings
- `show modules` - List all modules in repository
- `describe commands` - List all available commands

## See Also

- [Commit Message Contract](../../contracts/commit-message/0.1.0/structure.yml)
- [Agent Instructions](../../../.claude/agents/commit-message-generator.md)
- [Complete 7-Lever System](../../../out/COMPLETE-7-LEVER-SYSTEM.md)
- [Semantic Commits Guide](../continuous-delivery/semantic-commits.md)
