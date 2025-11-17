# Generate Commit Message

Write a commit message following this format (do NOT wrap in ```):

<type>(<scope>): <one-line summary>

Auditor-Summary: <one audit-ready sentence>

<2-4 sentences describing changes>

Changes: <N> files, +<X> insertions, -<Y> deletions

Rules:

- Start with `feat(`, `fix(`, `chore(`, `docs(`, `refactor(`, `test(`, or `perf(`
- If multiple modules: use `feat(multi-module):`
- If single module: use `feat(module-name):`
- Max 72 characters per line
- No trailing periods on header
- Wrap body text at 72 characters

Example:

```text
feat(multi-module): add validation pipeline

Auditor-Summary: Implemented contract-based validation for commits.

This commit introduces formal specifications and validation rules
for commit messages across multiple modules. Changes include contract
definitions, CLI implementation, and enforcement logic.

Changes: 5 files, +330 insertions, -129 deletions
```

Generate the commit message now (output ONLY the commit message, no explanations):
