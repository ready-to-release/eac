Generate a commit message using this exact format:

```text
refactor(multi-module): simplify commit message prompts

Auditor-Summary: Removed template embedding for clearer AI instructions.

This commit simplifies the prompts by removing YAML template variables
and using direct format instructions. Updated both top-level and module
prompts to show concrete examples rather than abstract schemas.

Changes: 8 files, +95 insertions, -150 deletions
```

Your output must:
- Start with `<type>(<scope>): <summary>` (types: feat, fix, refactor, docs, chore, test, perf, style)
- Include `Auditor-Summary: <sentence>` on line 3
- Include body text (2-4 sentences, wrapped at 72 chars)
- End with `Changes: N files, +X insertions, -Y deletions`

Output only the commit message. No markdown fences, no extra text.
