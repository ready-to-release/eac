<!-- EDITOR
# Editor: how-to-guides/commands/squash-message-command.md

## Soul

Guide for AI-powered squash commit message generation for GitHub PR workflow that synthesizes branch commits into cohesive narrative suitable for squash merge UI.

## Sections

1. Problem/Solution
2. Key Benefits
3. Quick Start
4. Command Reference
5. Differences from commit-message
6. Message Format
7. Typical Workflows
8. Example Outputs
9. Synthesis Guidelines
10. Debug Mode
11. Best Practices
12. Troubleshooting
13. Configuration
14. Integration
15. Advanced Usage
16. Summary

## Notes

- Auditor-Summary section is unique compliance feature
- ">>>>>>OUTPUT START<<<<<<" marker is unusual UX pattern

-->

# Squash Message Command

**Problem**: When preparing to squash merge a pull request in GitHub, you need to create a cohesive commit message that synthesizes all branch commits into a single narrative, rather than just listing individual commits.

**Solution**: Use `create squash-message` to generate AI-powered squash commit messages that analyze your branch commits and create a comprehensive message suitable for GitHub PR squash merge UI.

## Key Benefits

- Synthesizes multiple commits into cohesive narrative
- Analyzes cumulative branch changes vs base branch
- Follows conventional commit format
- Includes Auditor-Summary for compliance tracking
- Provides change statistics (files, insertions, deletions)
- Focuses on "what" and "why" rather than commit-by-commit details
- Designed for copying into GitHub PR squash merge UI

## Quick Start

```bash
# On your feature branch with multiple commits
r2r eac create squash-message

# Specify custom base branch
r2r eac create squash-message --base=develop

# Enable debug mode to inspect AI generation
r2r eac create squash-message --debug
```

## Command Reference

### create squash-message

Generate AI-powered squash commit messages from branch commits.

```bash
r2r eac create squash-message [options]

# Options:
--base=BRANCH         # Base branch for comparison (default: main)
--debug, -d           # Save intermediate outputs to out/ directory

# Examples:
r2r eac create squash-message                    # Compare against main
r2r eac create squash-message --base=develop    # Compare against develop
r2r eac create squash-message --debug            # Debug AI generation process
```

**What it does:**

1. **Branch Analysis**: Identifies commits from baseBranch..HEAD
2. **Diff Analysis**: Analyzes cumulative diff and changed files
3. **Module Mapping**: Identifies affected modules
4. **Context Building**: Creates comprehensive context for AI
5. **AI Generation**: Synthesizes commits into cohesive narrative
6. **Output**: Displays message ready for GitHub PR squash merge

**Generated format:**

```text
<type>(<scope>): <summary>

Auditor-Summary: <one sentence describing the overall change>

<body: 2-4 sentences explaining what this branch does and why>

Changes: N files, +X insertions, -Y deletions
```

## Differences from create commit-message

| Aspect | create commit-message | create squash-message |
|--------|----------------------|----------------------|
| Input | Staged changes | Branch commits |
| Comparison | Working tree vs HEAD | Current branch vs base branch |
| Output | Commit message for immediate use | Message for GitHub PR squash UI |
| --commit flag | ✅ Available | ❌ Not available (output only) |
| Use case | Regular development commits | PR squash merge preparation |
| Narrative style | Detailed per-module | Synthesized overall feature |

## Message Format

### Header Line

```text
<type>(<scope>): <summary>
```

- **Type**: feat, fix, refactor, docs, chore, test, perf, style
- **Scope**: Module name or `multi-module`
- **Summary**: Concise description of overall change (max 72 chars)

### Auditor-Summary

```text
Auditor-Summary: <one clear sentence>
```

- Summarizes the essential change across all commits
- Focuses on business value or technical outcome
- Required for compliance tracking

### Body

```text
<2-4 sentences, wrapped at 72 chars>
```

- Explains WHAT this branch accomplishes overall
- Explains WHY the changes were needed
- Synthesizes information from multiple commits
- Mentions key architectural decisions if relevant

### Changes Line

```text
Changes: N files, +X insertions, -Y deletions
```

- Git statistics summary from branch diff
- Automatically generated from diff stats

## Typical Workflows

### GitHub PR Workflow

```bash
# 1. Complete feature branch with multiple commits
git log --oneline main..HEAD
# abc123 feat: add JWT validation
# def456 fix: handle edge case
# ghi789 docs: update authentication guide

# 2. Generate squash message
r2r eac create squash-message --base=main

# 3. Copy output starting from >>>>>>OUTPUT START<<<<<<

# 4. Open GitHub PR and click "Squash and merge"

# 5. Paste the generated message into GitHub's squash merge UI

# 6. Complete the merge
```

### Custom Base Branch

```bash
# Working from develop branch
git checkout -b feature/authentication develop

# ... make multiple commits ...

# Generate squash message against develop
r2r eac create squash-message --base=develop
```

### Debug AI Generation

```bash
# Generate with debug output
r2r eac create squash-message --debug

# Inspect debug files
ls out/logs/commit/
# squash-context.md      - Branch info, commits, diff
# squash-prompt.md       - Full AI prompt
# squash-ai-response.md  - Raw AI response
# squash-final.md        - Final formatted message
```

## Example Outputs

### Single Module Feature

```text
feat(eac-commands): implement squash message generation

Auditor-Summary: Added new command to generate AI-powered squash commit messages for GitHub PR workflow.

Implemented complete squash message generation using AI to synthesize
branch commits into cohesive narratives. The command analyzes branch
history and cumulative diffs to create messages suitable for GitHub
PR squash merge UI.

Changes: 5 files, +412 insertions, -0 deletions
```

### Multi-Module Refactoring

```text
refactor(multi-module): consolidate risk assessment reporting

Auditor-Summary: Unified risk assessment report generation across risk-assess and risk-profile commands.

Extracted common template rendering logic into shared template package
to eliminate duplication between risk assessment commands. The
refactoring improves maintainability and ensures consistent report
formatting across all risk-related operations.

Changes: 8 files, +234 insertions, -189 deletions
```

### Bug Fix

```text
fix(eac-core): resolve git branch comparison edge cases

Auditor-Summary: Fixed branch comparison to properly handle remote branch references and empty diffs.

Corrected GetBranchCommits to check for origin/ prefix when resolving
base branch references. Added validation to return clear error when
no commits exist ahead of base branch instead of empty result.

Changes: 3 files, +45 insertions, -12 deletions
```

## Synthesis Guidelines

The AI follows these principles when generating squash messages:

### DO:
- Identify the main theme/purpose across all commits
- Combine related changes into cohesive description
- Elevate to feature-level or change-level perspective
- Use commit messages as hints about intent
- Focus on the end state, not the journey

### DON'T:
- List commits individually ("First commit did X, second commit did Y")
- Say "this PR" or "this branch" (it's a commit message)
- Include commit hashes or commit counts
- Describe intermediate states or WIP commits
- Use phrases like "various changes" or "multiple updates"

## Debug Mode

Use `--debug` to inspect the AI generation process:

```bash
r2r eac create squash-message --debug
```

Creates debug files in `out/logs/commit/`:

```text
out/logs/commit/
├── squash-context.md      # Branch info, commits, modules, diff
├── squash-prompt.md       # Full AI prompt with context
├── squash-ai-response.md  # Raw AI response
└── squash-final.md        # Final formatted message
```

**Use debug mode when:**

- AI generates unexpected messages
- Understanding how commits are synthesized
- Message doesn't capture the feature essence
- Troubleshooting formatting issues
- Customizing AI prompts

## Best Practices

### Branch Preparation

```bash
# ✅ Good: Clean branch with logical commits
git log --oneline main..HEAD
# feat: add user authentication
# test: add authentication tests
# docs: update authentication guide

# ❌ Avoid: Branch with WIP commits
git log --oneline main..HEAD
# WIP
# fix typo
# WIP again
# actually works now
```

### Base Branch Selection

```bash
# ✅ Good: Match your PR target branch
r2r eac create squash-message --base=main     # For PR to main
r2r eac create squash-message --base=develop  # For PR to develop

# ❌ Avoid: Wrong base branch
r2r eac create squash-message --base=main     # When PR targets develop
```

### Message Quality

The AI generates better messages when:

- Branch has focused, related commits
- Commit messages are descriptive
- Changes affect logical feature boundaries
- Test commits accompany implementation
- Branch is relatively recent (not stale)

## Troubleshooting

| Problem | Solution |
|---------|----------|
| No commits ahead of base | Ensure you're on feature branch: `git log main..HEAD` |
| AI API error | Check API key configuration: `r2r eac init` |
| Message lists commits | Branch may have too many disparate changes; consider splitting |
| Wrong base branch detected | Explicitly specify with `--base=BRANCH` |
| Message too generic | Ensure commit messages are descriptive |
| Missing module information | Verify modules.yml is up to date |

## Configuration

### AI Provider Setup

```bash
# Configure AI provider (first time)
r2r eac init

# Select provider: openai, anthropic, azure, etc.
# Enter API key when prompted
```

### Custom Prompts

Customize AI behavior by editing the squash prompt:

```text
.r2r/eac/ai/commit-message/
└── squash.md       # How to synthesize commits into squash message
```

The prompt template includes:
- Structure guidelines (header, Auditor-Summary, body, Changes)
- Synthesis instructions (DO/DON'T)
- Examples of good vs bad messages
- Output format requirements

## Integration with GitHub PR Workflow

### Standard PR Process

```bash
# 1. Create feature branch
git checkout -b feature/new-capability

# 2. Make multiple commits during development
git commit -m "feat: add capability X"
git commit -m "test: add tests for X"
git commit -m "docs: document X"

# 3. Push to remote
git push origin feature/new-capability

# 4. Create PR in GitHub

# 5. After review approval, generate squash message
r2r eac create squash-message

# 6. Copy message output

# 7. In GitHub PR, click "Squash and merge"

# 8. Replace default message with generated message

# 9. Complete merge
```

### Why Use Squash Message for PRs?

**Without squash-message:**
```text
feat: add capability X (#123)

* feat: add capability X
* test: add tests for X
* fix: handle edge case
* docs: document X
* fix typo in docs
```

**With squash-message:**
```text
feat(eac-core): implement capability X with comprehensive testing

Auditor-Summary: Added capability X with full test coverage and documentation.

Implemented capability X to handle new requirement Y. The feature
includes edge case handling and comprehensive documentation for
maintainers.

Changes: 8 files, +234 insertions, -12 deletions
```

## Advanced Usage

### Review-then-Squash Workflow

```bash
# Generate and save message to file
r2r eac create squash-message > squash-message.txt

# Edit manually if needed
vim squash-message.txt

# Use in GitHub when ready
```

### CI/CD Integration

```bash
# Generate squash message in CI for PR checks
r2r eac create squash-message > /tmp/squash-message.txt

# Validate format
if ! grep -q "Auditor-Summary:" /tmp/squash-message.txt; then
  echo "Generated message missing Auditor-Summary"
  exit 1
fi
```

### Pre-merge Hook

```bash
# .github/workflows/pr-checks.yml
- name: Generate squash message
  run: |
    r2r eac create squash-message
    # Could post as PR comment for reviewer reference
```

## Summary

1. **Complete feature branch**: Multiple logical commits
2. **Generate message**: `r2r eac create squash-message`
3. **Copy output**: Starting from `>>>>>>OUTPUT START<<<<<<`
4. **Open GitHub PR**: Click "Squash and merge"
5. **Paste message**: Replace GitHub's default
6. **Complete merge**: Finalize the PR

The squash-message command analyzes your entire branch history to create a synthesized, cohesive commit message that tells the story of your feature as a whole - perfect for GitHub PR squash merges.
