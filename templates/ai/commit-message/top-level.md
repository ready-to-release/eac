# Generate Top-Level Commit Message

You are an expert in writing clear, professional commit messages.

Generate a commit message summarizing all changes across the repository.

## Structure

```text
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

```text
feat(api): add user authentication endpoint

Auditor-Summary: Implements JWT-based authentication for API access.

This commit adds a new authentication endpoint that validates user
credentials and returns JWT tokens. Includes rate limiting and input
validation for security.

Changes: 12 files, +450 insertions, -20 deletions
```

## Output Rules

- Start with header (not "Auditor-Summary")
- No markdown fences or explanations
- Return only the commit message
- STOP after the "Changes:" line - do not add anything else

## CRITICAL: What NOT to include

The following are FORBIDDEN in your output:

- NO markdown headers (## or ###)
- NO "Module Changes" sections
- NO bullet point lists of files or changes
- NO per-module breakdowns
- NO file paths

Module-specific details will be generated separately. Your job is ONLY
the high-level summary (header + auditor-summary + body + changes line).

Generate now based on the context below:
