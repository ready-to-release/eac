# Release Notes

## Overview

Release notes communicate **what changed** to stakeholders. Good release notes help users understand new features, prepare for breaking changes, and know about security fixes.

---

## Template

```markdown
# Release v1.2.0

**Release Date:** YYYY-MM-DD

## Summary

[1-2 sentence overview of this release]

## New Features

- [Feature description with user benefit]

## Enhancements

- [Enhancement description]

## Bug Fixes

- [Fix description] (#issue-number)

## Breaking Changes

- [Breaking change with migration guide]

## Security

- [Security fix or CVE addressed]

## Deprecations

- [Feature being phased out with timeline]

## Known Issues

- [Documented limitations]
```

---

## Section Guidelines

### Summary

1-2 sentences maximum. Highlight the most important change.

**Good:** "This release adds OAuth2 authentication and improves search performance by 50%."

**Bad:** "Various updates and fixes."

### New Features

Focus on user benefit, not implementation. Link to documentation.

```markdown
- **OAuth2 Login**: Users can now sign in with Google, GitHub, or Microsoft.
  See [Authentication Guide](docs/auth.md) for setup.
```

### Bug Fixes

Reference issue numbers. Describe the user-visible problem fixed.

```markdown
- Fixed crash when uploading files larger than 100MB (#234)
- Corrected timezone handling in scheduled reports (#245)
```

### Breaking Changes

Clearly mark as breaking. Provide migration path.

```markdown
- **API Change**: `/api/v1/users` removed. Migrate to `/api/v2/users`.
  Before: `GET /api/v1/users`
  After:  `GET /api/v2/users`
```

### Security

Reference CVE numbers. Describe impact without exploit details.

```markdown
- Updated dependencies to address CVE-2024-1234 (high severity).
  Recommend upgrading immediately.
```

---

## Writing Guidelines

**DO:**

- Write for your audience (technical vs end-user)
- Be specific ("50% faster" not "improved performance")
- Link to issue tracker (#1234)
- Provide migration guidance for breaking changes
- Use active voice ("Added feature" not "Feature was added")

**DON'T:**

- Use jargon for end-user release notes
- Omit breaking changes
- Be vague ("various bug fixes")
- Forget to mention security fixes
- Skip documentation for "obvious" changes

---

## Auto-Generation

Use conventional commits to auto-generate release notes:

```bash
# Generate from git log
git log v1.1.0..HEAD --pretty=format:"%s" | \
  grep -E "^(feat|fix|docs)"
```

Or use the changelog system:

```bash
# Generate changelog entries from commits
r2r release this <module>
```

See [Changelog System](./changelog-system.md) for the full release workflow.

---

## Review Checklist

Before publishing:

- [ ] All sections complete
- [ ] No typos or grammatical errors
- [ ] Breaking changes clearly marked
- [ ] Migration instructions tested
- [ ] Security issues properly disclosed
- [ ] Known issues documented

---

## Next Steps

- [Changelog System](./changelog-system.md) - How to create releases
- [Commit Messages](../workflow/commit-messages.md) - Conventional commit format
