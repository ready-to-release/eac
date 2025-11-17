# Customizing Commit Message Prompts

This guide explains how to customize the AI prompts used by `commit-ai` to match your team's commit message style and preferences.

## Overview

The `commit-ai` command uses AI to generate commit messages based on your staged changes. You can customize the prompts that guide the AI to produce commit messages in your preferred style.

## Two-Tier Prompt System

The CLI uses a two-tier system for loading prompts (in order of priority):

1. **User Override** (highest priority)
   - Location: `.r2r/prompts/commit/<name>.md`
   - Use this to customize prompts for your needs
   - Created automatically when you run `r2r init --ai <provider>`

2. **Built-in** (fallback)
   - Embedded in the CLI binary
   - Always available as a default
   - Used only if user override doesn't exist

## Available Prompts

### `top-level.md`

Generates the top-level commit message header and summary.

**Default behavior:**
- Uses `# multi-module:` for commits affecting multiple modules
- Uses `# <module-name>:` for single-module commits
- Follows conventional commit format: `<type>: <description>`
- Wraps body text at 72 characters

### `module.md`

Generates individual module sections for multi-module commits.

**Default behavior:**
- Starts with `## <module-name>`
- Follows conventional commit format: `<module-name>: <type>: <description>`
- Describes changes specific to that module
- Wraps text at 72 characters

## Customizing Prompts

### Step 1: Initialize (if not already done)

The `init` command automatically creates `.r2r/prompts/commit/` and copies the built-in prompts:

```bash
r2r init --ai claude-cli
```

If you've already initialized, the prompts are in `.r2r/prompts/commit/`:
- `top-level.md` - Top-level commit message
- `module.md` - Module section messages

### Step 2: Modify the Prompt

Edit `.r2r/prompts/commit/top-level.md` to match your style:

```markdown
---
description: Generate top-level commit message header and summary
model: claude-3-5-haiku-20241022
temperature: 0.3
max_tokens: 2000
---

# Your Custom Instructions

Modify the instructions to match your commit message style...
```

### Step 3: Test Your Changes

```bash
# Stage some changes
git add .

# Run commit-ai with debug flag
r2r commit-ai --debug
```

The debug output will show:
- Which prompt file was loaded (user override or built-in)
- The generated commit message
- Token usage

### Step 4: Iterate

Adjust your prompt based on the results and re-run `commit-ai` until you're satisfied.

## Prompt Frontmatter Options

```yaml
---
description: Brief description of the prompt
model: claude-3-5-haiku-20241022  # AI model to use
temperature: 0.3                  # Creativity (0.0-1.0, lower = more consistent)
max_tokens: 2000                  # Maximum output length
---
```

**Notes:**
- `model` can be overridden by `.r2r/agent-config.yml`
- `temperature` affects output variation (0.0 = deterministic, 1.0 = creative)
- `max_tokens` limits the response length

## Examples

### Example 1: Shorter Commit Messages

If you prefer concise commits, modify the body instructions:

```markdown
## Body Rules (AFTER BLANK LINE)

- 1-2 sentences maximum describing the overall changes
- Wrap at 72 characters per line
- No code snippets, no code blocks
- Focus on WHAT changed, skip WHY
```

### Example 2: Include Issue Numbers

Add instructions to reference issue numbers:

```markdown
## Header Rules (LINE 1 - MANDATORY)

**Format**: `# <module|multi-module>: <type>: <summary> [#<issue>]`

- Always end with issue number in brackets: [#123]
- Get issue number from branch name (e.g., feature/123-add-auth)
- Max 72 characters including issue reference
```

### Example 3: Team-Specific Commit Types

Customize the allowed commit types:

```markdown
- Types: `feature`, `bugfix`, `hotfix`, `cleanup`, `docs`, `test`
  (Instead of: feat, fix, refactor, etc.)
```

### Example 4: Include Affected Files

Add a section to list changed files:

```markdown
## Body Format

Paragraph 1: High-level summary
Paragraph 2: Implementation details
Paragraph 3: Files changed (list top 5 files)
```

## Sharing Custom Prompts with Your Team

### Option 1: Commit to Repository (Recommended)

Remove `.r2r/prompts/` from `.gitignore` and commit your prompts:

```bash
# Edit .gitignore and remove this line:
# .r2r/prompts/

# Commit the prompts
git add .r2r/prompts/
git commit -m "Add custom commit message prompts"
```

Team members will automatically use your custom prompts after pulling.

### Option 2: Document and Share Manually

Keep `.r2r/prompts/` gitignored but document your customizations:

1. Create `docs/team/commit-message-style.md`
2. Include your prompt customizations
3. Team members run `r2r init` and then copy the prompts manually

## Reverting to Built-in Prompts

To revert to the default built-in prompts:

```bash
# Delete your overrides
rm .r2r/prompts/commit/top-level.md
rm .r2r/prompts/commit/module.md

# The CLI will automatically use built-in prompts
```

Or re-run init to restore the defaults:

```bash
r2r init --ai <your-provider>
```

## Troubleshooting

### How do I know which prompt is being used?

Run with the `--debug` flag:

```bash
r2r commit-ai --debug
```

The debug output shows the prompt source.

### My changes aren't being applied

Check:
1. File name is exactly `top-level.md` or `module.md` (not `.example.md`)
2. File is in `.r2r/prompts/commit/` directory
3. YAML frontmatter is valid
4. Run with `--debug` to see which file is loaded

### The AI isn't following my instructions

Common issues:
- **Instructions are ambiguous**: Be very specific and use examples
- **Temperature too high**: Lower to 0.1-0.3 for more consistent output
- **Conflicting instructions**: Remove contradictions in your prompt
- **Model limitations**: Try a more capable model (opus vs haiku)

## Best Practices

1. **Start Small**: Modify one prompt at a time
2. **Test Frequently**: Run `commit-ai --debug` after each change
3. **Use Examples**: Include examples of desired output in your prompt
4. **Be Specific**: Vague instructions lead to inconsistent results
5. **Version Control**: Commit your prompts to share with team
6. **Document Why**: Add comments explaining your customizations

## Related Documentation

- [AI Providers](ai-providers.md) - Configure AI providers and models
- [Commit Message Contract](../explanation/contracts/commit-message.md) - Formal specification
- [Command Reference](../reference/commands/commit-ai.md) - commit-ai command details

## Feedback

If you create useful prompt customizations, consider:
- Sharing them with the community
- Opening a PR to improve the built-in prompts
- Documenting your style guide for your team
