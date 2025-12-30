# Commit Messages

How to write semantic commit messages that maintain traceability.

## Format

```text
<type>(<scope>): <description>

[optional body]

[optional footer]
```

## Type Keywords

| Type     | Purpose          | Example                                 |
| -------- | ---------------- | --------------------------------------- |
| feat     | New feature      | `feat(api): add user authentication`    |
| fix      | Bug fix          | `fix(parser): handle empty input`       |
| docs     | Documentation    | `docs(readme): update install steps`    |
| style    | Formatting       | `style(lint): apply gofmt`              |
| refactor | Code restructure | `refactor(auth): extract token service` |
| test     | Add/modify tests | `test(api): add login tests`            |
| chore    | Maintenance      | `chore(deps): update dependencies`      |

## Scope

The component or module affected:

- `api`, `cli`, `core`, `ui`
- `auth`, `billing`, `users`
- `ci`, `deps`, `config`

## Description Guidelines

- Use imperative mood: "add" not "added" or "adds"
- Keep under 50 characters
- Don't end with period
- Capitalize first letter

## Body (Optional)

Use the body to explain:

- **What** changed and **why**
- Any trade-offs or decisions made
- Context not obvious from code

Wrap at 72 characters per line.

## Footer (Optional)

Link to issues and indicate breaking changes:

```text
Relates to #123
Fixes #456
BREAKING CHANGE: API endpoint renamed
```

## Complete Example

```text
feat(api): add user authentication endpoint

Implements user login with JWT tokens. Uses bcrypt for password
hashing and RS256 for token signing.

- Add /auth/login POST endpoint
- Generate and return JWT tokens
- Store refresh tokens in Redis
- Add rate limiting (10 req/min)

Relates to #123
Co-authored-by: Jane Doe <jane@example.com>
```

## Commit Message for Releases

```text
chore(release): prepare v1.2.0

- Update CHANGELOG.md
- Bump version in package.json
- Generate release notes

Releases: v1.2.0
```

## What to Avoid

- Generic messages: "fix bug", "update code"
- Multiple unrelated changes in one commit
- Missing issue references for tracked work
- Implementation details that belong in code comments

## Related

- [Traceability Requirements](../cd-model/stages.md#stage-4-commit)
- [Pre-commit Setup](../quality-gates/precommit-setup.md)
- [Trunk-Based Development](./trunk-based-development.md)
- [Branching Strategies](./branching-strategies.md)
