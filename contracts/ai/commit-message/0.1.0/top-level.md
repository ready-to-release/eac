# Generate Top-Level Commit Message

You are an expert in writing clear, professional commit messages.

Generate a commit message summarizing all changes across the repository.

## Structure

```
<type>(<scope>): <summary>

Auditor-Summary: <one sentence>

<body: 2-4 sentences, wrapped at 72 chars>

Changes: N files, +X insertions, -Y deletions
```

## Requirements

**Header** (line 1): `<type>(<scope>): <summary>`
- Types: feat, fix, refactor, docs, chore, test, perf, style
- Scope: `multi-module` or specific module
- Max 72 chars, no trailing period

**Auditor-Summary** (line 3): One clear sentence

**Body** (line 5+): 2-4 sentences, wrapped at 72 chars

**Changes** (last line): Git statistics summary

## Example

```
refactor(multi-module): simplify commit message prompts

Auditor-Summary: Removed template embedding for clearer AI instructions.

This commit simplifies the prompts by removing YAML template variables
and using direct format instructions. Updated both top-level and module
prompts to show concrete examples rather than abstract schemas.

Changes: 8 files, +95 insertions, -150 deletions
```

## Output Rules

- Start with header (not "Auditor-Summary")
- No markdown fences or explanations
- Return only the commit message

Generate now based on the context below:
