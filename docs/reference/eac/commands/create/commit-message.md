# create commit-message

<!-- book:cmd create commit-message -->

## How It Works

The command uses AI generation with format support:

- **Format**: Generates plaintext commit messages following conventional commit format (type, scope, description, body)
- **Validation**: Validates message format and structure automatically
- **Retry**: If validation fails, AI receives error feedback and regenerates improved output
- **Customization**: Uses three-tier prompt system for team-specific commit message styles

Supported formats: Conventional commits with module-aware sections for monorepo commits.

## Custom Prompts

Customize AI behavior using the **three-tier prompt system**. Prompts are loaded with the following priority:

1. **Command Flag** (highest priority)

   ```bash
   r2r eac create commit-message --prompt /path/to/custom.md
   ```

2. **Team Override** (version controlled)

   - Location: `.r2r/eac/templates/ai/commit-message/`
   - Committed to git, affects entire team

3. **System Default** (fallback)
   - Location: `templates/ai/commit-message/`
   - Shipped with r2r

### Available Prompts

| Prompt File    | Purpose                      | Override Location                                   |
| -------------- | ---------------------------- | --------------------------------------------------- |
| `module.md`    | Multi-module commit sections | `.r2r/eac/templates/ai/commit-message/module.md`    |
| `top-level.md` | Top-level commit header      | `.r2r/eac/templates/ai/commit-message/top-level.md` |

### Creating Team Overrides

```bash
# Install AI templates
r2r eac templates install ai

# Edit for your team's style
nano .r2r/eac/templates/ai/commit-message/module.md

# Commit to git
git add .r2r/eac/templates/ai/commit-message/
git commit -m "chore(eac): customize commit message prompt"

# Test with debug flag to verify
r2r eac create commit-message --debug
```

## See Also

- [How-to Guide](../../../../how-to-guides/eac/commands/development-workflow/make-commits-with-ai.md) - Quick start and common workflows
- [work commit](../work/commit.md) - Workspace-aware commits
- [init](../init/init.md) - Configure AI provider
- [create Commands](../categories/create.md)
