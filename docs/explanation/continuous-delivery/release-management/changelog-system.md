# Changelog-Based Release System

## Overview

This repository uses a **changelog-based release system**.

Releases are triggered by updating changelogs and merging to main - no manual tagging required.

**Key principle:** The changelog is the source of truth for releases.

When a new version appears in a changelog and merges to main,
the release workflow automatically creates a git tag and triggers the appropriate release pipeline.

---

## How It Works

The release process follows a simple flow:

1. **Check for changes** - Are there unreleased changes since the last release?
2. **Update changelog** - Finalize the changelog with a new version entry
3. **Create PR** - Submit the changelog update for review
4. **Merge to main** - After approval, merge triggers automation
5. **Automatic tagging** - `release-auto.yml` workflow creates git tag
6. **Build and publish** - Module-specific workflow builds artifacts

```text
Developer
    │
    ▼
┌──────────────────┐
│ Check pending    │  Are there changes to release?
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Update changelog │  Finalize version entry
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Create PR        │  Submit for review
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Merge to main    │  After approval
└────────┬─────────┘
         │
─────────┼─────────────────────────────────────────
         │  CI/CD (automated)
         ▼
┌──────────────────┐
│ Auto-tagging     │  Creates git tag from changelog
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Build & publish  │  Module-specific release workflow
└──────────────────┘
```

---

## Why Changelog-Based?

### Traditional Approach Problems

- Manual git tagging is error-prone
- Easy to forget to update changelog
- Tag creation and changelog updates can get out of sync
- Multiple steps to coordinate

### Changelog-Based Benefits

- **Single source of truth** - Changelog drives the release
- **Reviewed releases** - Changelog changes go through PR review
- **Automatic tagging** - No manual git tag commands
- **Audit trail** - Every release decision is a reviewed commit

---

## Reference Documentation

For complete CLI commands and workflow details, see:

**[Release Commands Reference](../../../reference/eac/continuous-delivery/changelog/release-commands.md)**

Complete implementation guide including:

- `eac release pending` - Check for unreleased changes
- `eac release this` - Finalize changelog for release
- `eac release tag-pending` - Check for untagged versions
- `eac release validate` - Validate changelog format
- Workflow automation details
- Troubleshooting guide

**[Changelog Format Specification](../../../reference/eac/continuous-delivery/changelog/format-specification.md)**

Keep a Changelog format details

---

## Related Documentation

- [Release Notes](./release-notes.md) - How to write effective release notes
- [Release Evidence](./release-evidence.md) - What evidence is required for approval
- [Commit Messages](../workflow/commit-messages.md) - Conventional commit format
