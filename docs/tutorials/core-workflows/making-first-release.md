# Making Your First Release

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Prerequisites:** [Working with Git Worktrees](./working-with-worktrees.md), completed feature ready to release

## Planned Content

This tutorial teaches you the complete release workflow: generating changelogs, validating CI status, creating release tags, and understanding versioning.

### What You'll Learn

- Check if a module has pending changes for release
- Generate changelog from conventional commits
- Validate changelog format and structure
- Check CI status before releasing
- Create release tags using CalVer format
- Understand versioning strategy (CalVer vs. SemVer)
- Handle release validation and rollback

### Tutorial Structure

1. **Understanding release readiness**
   - Check pending changes: `r2r release pending <module>`
   - View unreleased commits
   - Understand conventional commit format
   - Determine release scope

2. **Generating the changelog**
   - Generate: `r2r release changelog <module>`
   - Review generated CHANGELOG.md
   - Understand changelog sections (Added, Changed, Fixed, etc.)
   - Edit manually if needed

3. **Validating the release**
   - Validate format: `r2r validate release <module>`
   - Check version format (CalVer: YYYY.MM.DD.N)
   - Verify changelog structure
   - Ensure all required sections present

4. **Checking CI status**
   - Verify CI passed: `r2r release check-ci <module>`
   - Understand why CI must pass before release
   - View CI status: `r2r pipeline status`
   - Wait for CI if needed: `r2r pipeline wait`

5. **Creating the release tag**
   - Finalize release: `r2r release this <module>`
   - Creates git tag (e.g., `eac-commands/2025.12.10.0`)
   - Updates CHANGELOG.md with release date
   - Commits changelog update
   - Pushes tag to remote (if specified)

6. **Post-release verification**
   - Verify tag exists: `git tag -l`
   - Check GitHub release (if automated)
   - Verify artifacts built in CI
   - Monitor deployment

7. **Understanding versioning**
   - CalVer format: YYYY.MM.DD.N
   - Why CalVer for continuous delivery
   - Module-specific versioning
   - SemVer for libraries (R2R CLI)

### Example Release Workflow

The tutorial will walk through releasing a complete feature:

```bash
# Check if module has unreleased changes
r2r release pending eac-commands
# Output: Module has 5 unreleased commits

# Generate changelog from commits
r2r release changelog eac-commands
# Creates/updates CHANGELOG.md

# Review the generated changelog
cat go/eac/commands/CHANGELOG.md

# Validate changelog format
r2r validate release eac-commands
# Output: ✓ Valid release format

# Check CI status (must be green)
r2r release check-ci eac-commands
# Output: ✓ CI passed for commit abc123

# Create release tag
r2r release this eac-commands
# Creates tag: eac-commands/2025.12.10.0
# Updates CHANGELOG.md
# Commits and pushes
```

### Changelog Example

```markdown
# Changelog

## [2025.12.10.0] - 2025-12-10

### Added
- New `build` command for module compilation
- Support for cross-platform builds

### Changed
- Improved error messages in test command
- Updated dependency validation logic

### Fixed
- Fixed bug in module discovery
- Corrected path handling on Windows
```

### Key Concepts Covered

- Release readiness checks
- Conventional commits and changelog generation
- CalVer versioning strategy
- CI/CD validation before release
- Git tagging conventions
- Module-specific releases in monorepo

### Conventional Commits

Understanding commit message format:

```text
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature (MINOR)
- `fix`: Bug fix (PATCH)
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

### Release Checklist

Before creating a release:

- [ ] All tests pass locally
- [ ] CI is green for latest commit
- [ ] Changelog generated and reviewed
- [ ] Version format validated
- [ ] No pending merge conflicts
- [ ] Documentation updated
- [ ] Breaking changes documented

### Best Practices

- Release frequently (daily or weekly)
- Keep changelog human-readable
- Use conventional commits consistently
- Validate CI before releasing
- Tag releases for traceability
- Document breaking changes clearly

### Troubleshooting

Common issues:

- **CI not passing**: Fix tests before releasing
- **Invalid changelog format**: Run `r2r validate release`
- **No pending changes**: Nothing to release
- **Version conflict**: Check existing tags

### Next Steps

Congratulations! You now understand the complete development lifecycle from feature creation to release. Continue to [Advanced Practices](../advanced-practices/) to learn about compliance automation and CI/CD integration.

{{ diataxis_footer() }}
