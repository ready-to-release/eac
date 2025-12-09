# Release Notes

{{ page_breadcrumb() }}

How to create comprehensive release notes for production releases.

## Release Notes Template

```markdown
# Release v1.2.0

**Release Date:** YYYY-MM-DD
**Release Manager:** [Name]

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

## Upgrade Instructions

[Steps for upgrading from previous version]
```

## Section Guidelines

### Summary

- 1-2 sentences maximum
- Highlight the most important change
- Use user-focused language

**Good:** "This release adds OAuth2 authentication and improves search performance by 50%."

**Bad:** "Various updates and fixes."

### New Features

- Focus on user benefit, not implementation
- Include brief "how to use" if not obvious
- Link to documentation

```markdown
- **OAuth2 Login**: Users can now sign in with Google, GitHub, or Microsoft accounts.
  See [Authentication Guide](docs/auth.md) for setup.
```

### Bug Fixes

- Reference issue numbers
- Describe the user-visible problem fixed
- Keep technical details minimal

```markdown
- Fixed crash when uploading files larger than 100MB (#234)
- Corrected timezone handling in scheduled reports (#245)
```

### Breaking Changes

- Clearly mark as breaking
- Provide migration path
- Include code examples if applicable

**Example:**

```text
- **API Change**: `/api/v1/users` endpoint removed. Migrate to `/api/v2/users`.

  Before: curl https://api.example.com/v1/users
  After:  curl https://api.example.com/v2/users
```

### Security

- Reference CVE numbers when applicable
- Describe impact without disclosing exploit details
- Recommend immediate upgrade for critical fixes

**Example:**

```text
- Updated dependencies to address CVE-2024-1234 (high severity).
  Recommend upgrading immediately.
```

## Auto-Generation from Commits

Use semantic commits to auto-generate release notes:

```bash
# Generate from git log
git log v1.1.0..HEAD --pretty=format:"%s" | \
  grep -E "^(feat|fix|docs)" | \
  sed 's/^feat:/- **New:**/' | \
  sed 's/^fix:/- **Fixed:**/'
```

Or use tools like:

- `semantic-release`
- `conventional-changelog`
- `release-it`

## Review Checklist

Before publishing release notes:

- [ ] All sections complete
- [ ] No typos or grammatical errors
- [ ] Issue numbers linked
- [ ] Breaking changes clearly marked
- [ ] Migration instructions tested
- [ ] Security issues properly disclosed
- [ ] Known issues documented
- [ ] Links to documentation valid

## Publishing

1. **Internal review** - Technical lead and product owner approve
2. **Staging** - Publish to staging environment
3. **Production** - Publish with release deployment
4. **Notify** - Email, Slack, or other channels

## Related

- [Start Release](../cd-model/cd-model-stages-7-12.md#stage-8-start-release)
- [Release Documentation](./release-documentation.md)
- [Commit Messages](../workflow/commit-messages.md)

{{ diataxis_footer() }}
