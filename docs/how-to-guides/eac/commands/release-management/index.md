# Release Management

{{ page_breadcrumb() }}

Learn how to prepare and publish releases with changelog generation and version tagging.

## In This Section

| Guide | What You'll Accomplish |
|-------|------------------------|
| [Prepare Module Release](./prepare-module-release.md) | Complete pre-release checklist and create release |
| [Generate Changelog](./generate-changelog.md) | Create changelog from Git commits |
| [View Changelog and Release Notes](./view-changelog-release-notes.md) | Display changelog and release notes for modules |
| [View Release Specifications](./view-specifications.md) | View and analyze specification files for a release |
| [Create Release Tag](./create-release-tag.md) | Tag release with proper version |
| [Check CI Before Release](./check-ci-before-release.md) | Verify CI passes before releasing |

## Release Workflow

### Standard Release Process

1. [Check for pending changes](./prepare-module-release.md)
2. [Generate changelog](./generate-changelog.md) from commits
3. Validate changelog format
4. [Verify CI passes](./check-ci-before-release.md)
5. [Create release tag](./create-release-tag.md)
6. Push to remote

### Version Management

- CalVer for modules (date-based)
- SemVer for CLI tools (semantic versioning)
- Automated version extraction

{{ diataxis_footer() }}
