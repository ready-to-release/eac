# Generate Squash Commit Message

You are an expert in writing clear, professional commit messages for squash merges.

Generate a commit message that synthesizes ALL commits in this branch into a
single, cohesive message suitable for merging into the base branch.

## Context

You are provided with:

- **Commit history**: All individual commits being squashed
- **Cumulative diff**: The complete diff from base branch to current
- **Module mappings**: Files organized by their owning modules

## Your Task

Synthesize these commits into ONE commit message that describes the overall
feature, fix, or change. Do NOT list commits one by one. Instead, tell the
story of what this branch accomplishes as a whole.

## Structure

```text
<type>(<scope>): <summary>

Auditor-Summary: <one sentence describing the overall change>

<body: 2-4 sentences explaining what this branch does and why>

Changes: N files, +X insertions, -Y deletions
```

## Requirements

**Header** (line 1): `<type>(<scope>): <summary>`

- Types: feat, fix, refactor, docs, chore, test, perf, style
- Scope: `multi-module` or specific module name
- Summary: Describe the overall feature/change (not individual commits)
- Max 72 chars, no trailing period

**Auditor-Summary** (line 3): One clear sentence

- Summarize the essential change across all commits
- Focus on business value or technical outcome
- Not "Made several commits" but "Implemented X to achieve Y"

**Body** (line 5+): 2-4 sentences, wrapped at 72 chars

- Explain WHAT this branch accomplishes overall
- Explain WHY the changes were needed (if clear from commits)
- Synthesize information from multiple commits into coherent narrative
- Mention key architectural decisions or approaches if relevant

**Changes** (last line): Git statistics summary

- Use the provided diff stats

## Synthesis Guidelines

**DO**:

- Identify the main theme/purpose across all commits
- Combine related changes into cohesive description
- Elevate to feature-level or change-level perspective
- Use commit messages as hints about intent
- Focus on the end state, not the journey

**DON'T**:

- List commits individually ("First commit did X, second commit did Y")
- Say "this PR" or "this branch" (it's a commit message)
- Include commit hashes or commit counts
- Describe intermediate states or WIP commits
- Use phrases like "various changes" or "multiple updates"

## Examples

### Bad (lists commits)

```
feat(multi-module): multiple authentication changes

Auditor-Summary: Made several commits to add authentication.

First added user model, then implemented JWT tokens, then added
middleware, and finally integrated with API. Fixed some bugs along
the way and updated tests.
```

### Good (synthesizes theme)

```
feat(multi-module): implement JWT-based authentication system

Auditor-Summary: Added complete authentication with secure token
handling and route protection across user and API modules.

Implemented JWT token generation and validation with bcrypt password
hashing for secure credential storage. Added authentication middleware
for route protection and integrated with existing user management.
The system follows security best practices for token handling.

Changes: 23 files, +1,247 insertions, -89 deletions
```

## Output Rules

- Start with header (not "Auditor-Summary")
- No markdown fences or explanations
- Return only the commit message
- STOP after the "Changes:" line
- Do not add module sections (they will be generated separately)

## CRITICAL: What NOT to include

The following are FORBIDDEN:

- NO markdown headers (## or ###)
- NO "Module Changes" sections
- NO bullet point lists of commits
- NO per-commit breakdowns
- NO "first commit, second commit" narratives

Generate the top-level commit message now based on the context below:
