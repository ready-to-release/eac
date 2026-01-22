# Make Commits with AI

## What You'll Accomplish

Generate semantic commit messages automatically from your staged changes using AI analysis of code diffs.

## Prerequisites

- AI provider configured (see [Setup AI Provider](../getting-started/setup-ai-provider.md))
- Changes staged with `git add`
- Working in EAC-managed repository

## How It Works

The command uses AI to generate semantic commit messages:

- **Diff Analysis**: AI analyzes staged changes to understand what changed and why
- **Semantic Format**: Generates messages following conventional commit format (type, scope, description)
- **Structured Output**: AI generates plaintext following commit message conventions
- **Validation**: Output format is validated and retried automatically if needed
- **Quality**: Ensures commit messages are clear, accurate, and follow best practices

This produces professional commit messages that accurately describe your changes.

## Steps

### 1. Stage Your Changes

```bash
git add src/auth/login.go src/auth/login_test.go
```

**What happens**: Stages files for commit

### 2. Generate Commit Message

```bash
r2r eac create commit-message
```

**What happens**: AI analyzes changes, generates a semantic commit message with validation, and opens editor for review

### 3. Review and Save

The generated message follows this format:

```text
feat(auth): add JWT token validation

- Implement token signature verification
- Add expiration time checks
- Include refresh token support
```

Review, edit if needed, save and close editor.

**What happens**: Git commit is created with the message

## Auto-Commit (Skip Editor)

```bash
r2r eac create commit-message --commit
```

**What happens**: Generates message and commits automatically without opening editor

## Example Scenario

You've added login functionality and want a good commit message:

```bash
# Stage changes
git add src/auth/

# Generate and commit
r2r eac create commit-message

# AI generates:
# feat(auth): implement user login with session management
#
# - Add login endpoint with email/password
# - Create session tokens with 24h expiration
# - Implement bcrypt password hashing
# - Add login integration tests
```

## Common Issues

| Problem                     | Solution                              |
| --------------------------- | ------------------------------------- |
| "No AI provider configured" | Run `r2r eac init` to setup           |
| "No staged changes"         | Stage files with `git add` first      |
| Message too generic         | Use `--debug` flag to see AI analysis |

## Next Steps

- [Create Pull Request](./create-pull-request.md) → Generate PR description

## Related Commands

- [`create commit-message`](../../../../reference/eac/commands/create/commit-message.md) - Full command reference
- [`work commit`](../../../../reference/eac/commands/work/commit.md) - Commit in workspace
- [`create squash-message`](../../../../reference/eac/commands/create/squash-message.md) - Generate squash message
