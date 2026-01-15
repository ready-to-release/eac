# Generate Squash Commit Message

You are an expert in writing clear, professional commit messages for squash merges.

Generate a commit message that synthesizes ALL commits in this branch into a
single, cohesive message suitable for merging into the base branch.

Your task is to generate structured JSON data. The command will automatically format it into a conventional commit message.

## Context

You are provided with:

- **Commit history**: All individual commits being squashed
- **Cumulative diff**: The complete diff from base branch to current
- **Module mappings**: Files organized by their owning modules

## Your Task

Synthesize these commits into ONE commit message that describes the overall
feature, fix, or change. Do NOT list commits one by one. Instead, tell the
story of what this branch accomplishes as a whole.

## JSON Output Structure

Generate a JSON object matching this schema:

```json
{
  "type": "feat",
  "scope": "multi-module",
  "subject": "overall feature or change description",
  "auditorSummary": "One sentence describing the overall change across all commits",
  "body": "2-4 sentences explaining what this branch does and why",
  "changes": "N files, +X insertions, -Y deletions",
  "breaking": false
}
```

### JSON Field Requirements

**type** (required): One of: `feat`, `fix`, `refactor`, `docs`, `chore`, `test`, `perf`, `style`, `ci`, `build`
- Choose based on the overall theme across all commits
- If commits have mixed types, choose the most significant

**scope** (required): `multi-module` or specific module name
- Pattern: lowercase with hyphens
- Length: 1-20 characters
- Use `multi-module` if changes span multiple modules

**subject** (required): Describe the overall feature/change (not individual commits)
- Max 72 characters total when combined with `type(scope): `
- No trailing period
- Lowercase first letter
- Synthesize from all commits into cohesive description

**auditorSummary** (required): One clear sentence summarizing the essential change
- Summarize across ALL commits, not just one
- Focus on business value or technical outcome
- Not "Made several commits" but "Implemented X to achieve Y"

**body** (required): 2-4 sentences explaining what and why
- Each line wrapped at 72 characters
- Explain WHAT this branch accomplishes overall
- Explain WHY the changes were needed (if clear from commits)
- Synthesize information from multiple commits into coherent narrative
- Mention key architectural decisions or approaches if relevant

**changes** (required): Git statistics from the provided diff stats

**breaking** (optional): Set to `true` if any commit indicates breaking changes

### Synthesis Guidelines

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

### JSON Generation Rules

- Generate ONLY valid JSON
- No markdown code fences (no ```json)
- No explanations or commentary before/after the JSON
- Just pure JSON starting with `{` and ending with `}`
- All string fields must use double quotes
- Use proper JSON escaping for special characters (especially newlines: `\n`)

### Example JSON Output

**Bad** (lists commits):
```json
{
  "type": "feat",
  "scope": "multi-module",
  "subject": "multiple authentication changes",
  "auditorSummary": "Made several commits to add authentication.",
  "body": "First added user model, then implemented JWT tokens, then added\nmiddleware, and finally integrated with API. Fixed some bugs along\nthe way and updated tests.",
  "changes": "23 files, +1,247 insertions, -89 deletions",
  "breaking": false
}
```

**Good** (synthesizes theme):
```json
{
  "type": "feat",
  "scope": "multi-module",
  "subject": "implement JWT-based authentication system",
  "auditorSummary": "Added complete authentication with secure token handling and route protection across user and API modules.",
  "body": "Implemented JWT token generation and validation with bcrypt password\nhashing for secure credential storage. Added authentication middleware\nfor route protection and integrated with existing user management.\nThe system follows security best practices for token handling.",
  "changes": "23 files, +1,247 insertions, -89 deletions",
  "breaking": false
}
```

## Output Format (for your reference)

The JSON will be automatically converted to conventional commit format:

```text
type(scope): subject

Auditor-Summary: auditorSummary

body

Changes: changes
```

Example final output:
```text
feat(multi-module): implement JWT-based authentication system

Auditor-Summary: Added complete authentication with secure token
handling and route protection across user and API modules.

Implemented JWT token generation and validation with bcrypt password
hashing for secure credential storage. Added authentication middleware
for route protection and integrated with existing user management.
The system follows security best practices for token handling.

Changes: 23 files, +1,247 insertions, -89 deletions
```

## CRITICAL: What NOT to include

The following are FORBIDDEN in your JSON output:

- NO "modules" array or per-module breakdowns
- NO per-commit breakdowns or commit lists
- NO markdown formatting within strings
- NO nested objects beyond the schema shown
- NO "first commit, second commit" narratives

Module-specific details will be generated separately if needed.

Generate JSON now based on the context below:
