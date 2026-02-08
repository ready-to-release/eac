# Understanding Release Types

**Status**: Active
**Introduced**: 2026-01-21 (Phase 1-4 implementation)
**Related**: [Module Changelogs](changelog/module-changelogs.md), [Release Workflows](workflows/release-workflows.md)

## Overview

The EAC repository uses a **release type system** to distinguish between modules that are released to users and modules that are internal implementation details.

This classification drives workflow automation, changelog location, and architectural boundaries.

## The Four Release Types

### 1. `published` - Public Releases

**Purpose**: Modules that are released as standalone artifacts for external consumption.

**Characteristics**:

- Released to GitHub Releases with version tags
- Have public-facing changelogs in `release/<module>/CHANGELOG.md`
- Trigger GitHub release workflows on version changes
- Follow semantic versioning (SemVer)
- Represent stable public APIs

**Examples**:

- `clie-cli` - CLI binary for end users
- `eac-ext` - VS Code extension
- `docs` - HTML documentation site
- `books` - PDF/EPUB documentation books

**Workflow**:

```yaml
versioning:
  scheme: SemVer
  changelog: release/clie-cli/CHANGELOG.md
  release_type: published
```

---

### 2. `internal` - Internal Modules

**Purpose**: Modules that are part of the implementation but not released independently.

**Characteristics**:

- Changelogs in module root (e.g., `go/cli/eac/CHANGELOG.md`)
- Track changes for development history
- Do NOT trigger release workflows
- May or may not follow semantic versioning
- Represent internal architecture boundaries

**Examples**:

- `eac-cli` - Go command implementations
- `eac-mcp-server` - MCP server command bindings
- `clie-installer` - Installation scripts
- `vscode-commit` - VS Code commit message extension

**Workflow**:

```yaml
versioning:
  scheme: SemVer
  changelog: go/cli/eac/CHANGELOG.md
  release_type: internal
```

**Why Track Internal Changelogs?**

Even though internal modules aren't released, their changelogs serve important purposes:

1. **Development history**: Track why changes were made
2. **Dependency management**: Understand when to bump bundle versions
3. **Testing coordination**: Know what needs retesting when modules change
4. **Onboarding**: Help new developers understand module evolution

---

### 3. `bundle` - Meta-Releases

**Purpose**: Modules that bundle multiple internal modules into a single release.

**Characteristics**:

- Released as a meta-artifact (e.g., combined installer, multi-platform binary)
- Have changelogs in `release/<module>/CHANGELOG.md`
- Version bumps when **any dependency changes**
- Follow semantic versioning
- Represent coordinated releases of multiple components

**Examples**:

- `clie-eac-bundle` - Combined release of EAC CLI + extensions

**Workflow**:

```yaml
versioning:
  scheme: SemVer
  changelog: release/clie-eac-bundle/CHANGELOG.md
  release_type: bundle
dependencies:
  build_deps:
    - clie-cli
    - eac-ext
```

**Bundle Version Strategy**:

When a dependency changes:

- **Major change in dependency** → Bump bundle major
- **Minor change in dependency** → Bump bundle minor
- **Patch change in dependency** → Bump bundle patch

This ensures users can track which component versions are included in each bundle release.

---

### 4. `none` - No Releases

**Purpose**: Modules that are never released and have no version tracking.

**Characteristics**:

- No changelog file
- No version numbers
- No release workflows
- Pure library or shared code
- Embedded into other modules during build

**Examples**:

- `eac-core` - Core Go libraries used by commands

**When to Use**:

Use `release_type: none` when:

- Module is a pure library with no standalone value
- Changes are always accompanied by changes in dependent modules
- Versioning would add overhead without value

**Note**: Most modules should use `internal` instead. Only use `none` when the module truly has no identity outside its dependents.

---

## Changelog Location Rules

The release type determines changelog location:

| Release Type | Changelog Location              | Example                               |
| ------------ | ------------------------------- | ------------------------------------- |
| `published`  | `release/<module>/CHANGELOG.md` | `release/clie-cli/CHANGELOG.md`        |
| `bundle`     | `release/<module>/CHANGELOG.md` | `release/clie-eac-bundle/CHANGELOG.md` |
| `internal`   | `<module-root>/CHANGELOG.md`    | `go/cli/eac/CHANGELOG.md`        |
| `none`       | No changelog                    | N/A                                   |

**Rationale**:

- **Published/Bundle** → `release/` folder because these are public-facing releases
- **Internal** → Module root to keep implementation details close to code
- **None** → No changelog because changes are tracked by dependents

---

## Workflow Validation

Release workflows **automatically filter** by release type:

```yaml
# .github/workflows/release-clie-cli.yaml
jobs:
  approve:
    steps:
      - uses: ./.github/actions/approve-release
        with:
          module: clie-cli
          # Only proceeds if module.versioning.release_type == "published"
```

**Validation Rules**:

1. ✅ **Published/Bundle** modules CAN trigger release workflows
2. ❌ **Internal/None** modules CANNOT trigger release workflows
3. ✅ All modules with versioning MUST have valid `release_type`
4. ✅ Changelog location MUST match release type

Tests enforce these rules:

- `go/core/contracts/modules/release_type_test.go`
- `go/cli/eac/impl/release/release_type_workflow_test.go`

---

## Migration Guide

### Converting Internal → Published

To promote an internal module to published status:

1. **Update `repository.yml`**:

   ```yaml
   versioning:
     scheme: SemVer
     changelog: release/my-module/CHANGELOG.md # Move from module root
     release_type: published # Change from internal
   ```

2. **Move changelog**:

   ```bash
   mkdir -p release/my-module
   git mv <module-root>/CHANGELOG.md release/my-module/CHANGELOG.md
   ```

3. **Create release workflow**:

   ```bash
   cp .github/workflows/release-clie-cli.yaml .github/workflows/release-my-module.yaml
   # Update module references in the workflow
   ```

4. **Update tests**: Add to published module list in test expectations

### Converting Published → Internal

To demote a published module (rare):

1. Remove release workflow (`.github/workflows/release-<module>.yaml`)
2. Move changelog from `release/<module>/` to module root
3. Update `release_type: published` → `internal` in `repository.yml`
4. Archive any existing releases with clear deprecation notice

---

## Decision Tree

Use this to determine the correct release type:

```text
Is the module released to external users?
├─ YES → Is it a standalone artifact?
│        ├─ YES → published
│        └─ NO  → Is it bundling other modules?
│                 ├─ YES → bundle
│                 └─ NO  → internal (embedded component)
└─ NO → Does it have any independent identity/versioning?
         ├─ YES → internal
         └─ NO  → none (pure library)
```

**Examples**:

- "CLI binary for end users" → **published**
- "Internal command package" → **internal**
- "Combined installer with CLI + extensions" → **bundle**
- "Shared utility library" → **none**

---

## Anti-Patterns

### ❌ Using `published` for Internal Modules

**Problem**: Creates unnecessary public-facing changelogs and release workflows.

**Solution**: Use `internal` for implementation modules that aren't released independently.

### ❌ Using `none` for Modules with Independent Changes

**Problem**: Loses development history and makes dependency tracking harder.

**Solution**: Use `internal` to track changes even if the module isn't released.

### ❌ Inconsistent Changelog Locations

**Problem**: Published module with changelog in module root, or internal module with changelog in `release/`.

**Solution**: Follow the changelog location rules strictly. Validation tests will catch violations.

### ❌ Creating Release Workflows for Internal Modules

**Problem**: Workflow tries to release but module isn't meant for public consumption.

**Solution**: Only create release workflows for `published` and `bundle` modules.

---

## Related Architecture

- [DR-011: Calendar Versioning Policy](../decision-records/dr011.md) - CalVer usage for docs
- [Module Changelogs](changelog/module-changelogs.md) - Changelog format and structure
- [Release Workflows](workflows/release-workflows.md) - Automated release process
- [Supporting Modules](../modules/index.md) - Module catalog

---

## History

| Date       | Change                                                     |
| ---------- | ---------------------------------------------------------- |
| 2026-01-21 | Introduced release type system (Phase 1-4 implementation)  |
| 2026-01-21 | Moved internal module changelogs to module roots (Phase 2) |
| 2026-01-21 | Added workflow validation by release type (Phase 3)        |
