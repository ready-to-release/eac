# Command Restructure Plan

## Overview

Clean restructure from ~70 commands to ~59 commands with consistent verb-first naming.

## Design Principles

1. **Verb-first**: All commands start with action verbs
2. **show vs get**: `show` = human-readable tables, `get` = structured YAML/JSON for piping
3. **Consistent flags**: Reuse `--format`, `--changed`, `--staged` across commands
4. **AI generation**: All AI-powered commands under `create`
5. **No legacy support**: Clean cut-over, no aliases or deprecation

---

## Command Mappings

### Renames and Merges

| Old Command | New Command | Change |
|-------------|-------------|--------|
| `show files changed` | `show files --changed` | Merge into files.go |
| `show files staged` | `show files --staged` | Merge into files.go |
| `show moduletypes` | `show module-types` | Rename |
| `help` | `show help` | Move to show/ |
| `describe commands` | `get commands` | Move to get/ |
| `work list` | `show workspaces` | Move to show/ |
| `pipeline status` | `show pipeline-status` | Move to show/ |
| `test list-suites` | `test --list-suites` | Merge as flag |
| `get changed-modules-ci` | `get changed-modules --ci` | Merge as flag |
| `pipeline wait` | `run pipeline --wait` | Merge as flag |

### New Verb Categories

#### CREATE (AI-powered generation)

| Old Command | New Command | Old File | New File |
|-------------|-------------|----------|----------|
| `specs create` | `create spec` | `specs/create/create.go` | `create/spec.go` |
| `design create` | `create design` | `design/create/create.go` | `create/design.go` |
| `commit message` | `create commit-message` | `commit/message.go` | `create/commit-message.go` |
| `work pr` | `create pr` | `work/pr.go` | `create/pr.go` |
| `risks assessment` | `create risk-assessment` | `risks/assessment/assessment.go` | `create/risk-assessment.go` |
| `risks create` | `create risk-controls` | `risks/create/create.go` | `create/risk-controls.go` |

#### UPDATE

| Old Command | New Command | Old File | New File |
|-------------|-------------|----------|----------|
| `design update` | `update design` | `design/update/update.go` | `update/design.go` |

#### SERVE (interactive)

| Old Command | New Command | Old File | New File |
|-------------|-------------|----------|----------|
| `docs serve` | `serve docs` | `docs/serve.go` | `serve/docs.go` |
| `design serve` | `serve design` | `design/serve.go` | `serve/design.go` |

#### RUN (orchestration)

| Old Command | New Command | Old File | New File |
|-------------|-------------|----------|----------|
| `pipeline run` | `run pipeline` | `pipeline/run.go` | `run/pipeline.go` |
| `pipeline wait` | `run pipeline --wait` | `pipeline/wait.go` | Merge into run/pipeline.go |

### Phase 3: Validate Consolidation

| Old Command | New Command | Old File | New File |
|-------------|-------------|----------|----------|
| `specs validate` | `validate specs` | `specs/validate/validate.go` | `validate/specs.go` |
| `specs unused-steps` | `validate specs --unused-steps` | `specs/unused_steps.go` | Merge into validate/specs.go |
| `design validate` | `validate design` | `design/validate.go` | `validate/design.go` |

### Phase 4: Removals

| Command | Reason | Alternative |
|---------|--------|-------------|
| `completion` | Broken (targets `go` not `r2r`) | Remove entirely |
| `list commands` | Redundant | `show help` |
| `commit reset` | Git native operation | Use `git reset --soft HEAD~1` |

---

## Implementation Details

### File Structure Changes

```
src/commands/impl/
├── build/           # Keep as-is
├── create/          # NEW - AI generation commands
│   ├── create.go    # Parent command
│   ├── spec.go      # was: specs/create/
│   ├── design.go    # was: design/create/
│   ├── commit-message.go  # was: commit/message.go
│   ├── pr.go        # was: work/pr.go
│   ├── risk-assessment.go # was: risks/assessment/
│   └── risk-controls.go   # was: risks/create/
├── get/             # Keep, add commands.go
├── init/            # Keep as-is
├── release/         # Keep as-is
├── run/             # NEW - orchestration
│   ├── run.go       # Parent command
│   └── pipeline.go  # was: pipeline/run.go + wait.go
├── serve/           # NEW - interactive servers
│   ├── serve.go     # Parent command
│   ├── docs.go      # was: docs/serve.go
│   └── design.go    # was: design/serve.go
├── show/            # Keep, add help.go, workspaces.go, pipeline-status.go
├── templates/       # Keep as-is
├── test/            # Keep, merge list-suites into test.go
├── update/          # NEW
│   ├── update.go    # Parent command
│   └── design.go    # was: design/update/
├── validate/        # Keep, add specs.go, design.go
└── work/            # Keep (minus pr.go, list.go)
```

### Directories to Remove

```
src/commands/impl/
├── ci/              # REMOVE (ci summary-link → standalone utility)
├── commit/          # REMOVE (message→create, reset→remove)
├── completion/      # REMOVE (broken)
├── describe/        # REMOVE (→ get commands)
├── design/          # REMOVE (distributed to create/update/serve/validate)
├── docs/            # REMOVE (→ serve docs)
├── help/            # REMOVE (→ show help)
├── list/            # REMOVE (→ show help)
├── pipeline/        # REMOVE (status→show, run/wait→run)
├── risks/           # REMOVE (→ create/*, show risks)
├── specs/           # REMOVE (create→create, validate→validate)
```

---

## Registry System Changes

### Current: Comment-Based Registration

```go
// Command: show files changed
// Short: Show changed files with module ownership
package show

func init() {
    registry.Register(ShowFilesChanged)
}
```

### Proposed: Same Pattern, New Names

```go
// Command: show files
// Short: Show repository files with module ownership
// Flag.changed: type=bool, usage=Show only changed (unstaged) files
// Flag.staged: type=bool, usage=Show only staged files
package show

func init() {
    registry.Register(ShowFiles)
}
```

---

## Command Count Summary

| Category | Before | After | Diff |
|----------|--------|-------|------|
| get | 10 | 12 | +2 |
| show | 12 | 13 | +1 |
| validate | 7 | 9 | +2 |
| build | 1 | 1 | 0 |
| test | 4 | 1 | -3 |
| create | 0 | 6 | +6 |
| update | 0 | 1 | +1 |
| serve | 0 | 2 | +2 |
| run | 0 | 1 | +1 |
| work | 7 | 5 | -2 |
| release | 3 | 3 | 0 |
| templates | 5 | 3 | -2 |
| init | 1 | 1 | 0 |
| ci | 1 | 1 | 0 |
| **Removed** | 5 | 0 | -5 |
| **TOTAL** | **~70** | **~59** | **-11** |

---

## Detailed File Changes

### Files to Modify (Rename Command)

| File | Old Command | New Command | Notes |
|------|-------------|-------------|-------|
| `show/moduletypes.go` | `show moduletypes` | `show module-types` | Comment change only |
| `get/modules-changed.go` | `get changed-modules` | `get changed-modules` | Add `--ci` flag |

### Files to Merge

| Target File | Source Files | New Flags |
|-------------|--------------|-----------|
| `show/files.go` | `show/files-changed.go`, `show/files-staged.go` | `--changed`, `--staged` |
| `test/test.go` | `test/list-suites.go` | `--list-suites` |
| `get/modules-changed.go` | `get/modules-changed-ci.go` | `--ci` |

### Files to Move

| Old Path | New Path | New Command |
|----------|----------|-------------|
| `help/help.go` | `show/help.go` | `show help` |
| `describe/commands.go` | `get/commands.go` | `get commands` |
| `work/list.go` | `show/workspaces.go` | `show workspaces` |
| `pipeline/status.go` | `show/pipeline-status.go` | `show pipeline-status` |
| `specs/create/create.go` | `create/spec.go` | `create spec` |
| `design/create/create.go` | `create/design.go` | `create design` |
| `design/update/update.go` | `update/design.go` | `update design` |
| `design/serve.go` | `serve/design.go` | `serve design` |
| `design/validate.go` | `validate/design.go` | `validate design` |
| `docs/serve.go` | `serve/docs.go` | `serve docs` |
| `pipeline/run.go` | `run/pipeline.go` | `run pipeline` |
| `commit/message.go` | `create/commit-message.go` | `create commit-message` |
| `work/pr.go` | `create/pr.go` | `create pr` |
| `risks/assessment/*.go` | `create/risk-assessment.go` | `create risk-assessment` |
| `risks/create/*.go` | `create/risk-controls.go` | `create risk-controls` |
| `specs/validate/validate.go` | `validate/specs.go` | `validate specs` |

### Files to Delete

| File | Reason |
|------|--------|
| `completion/completion.go` | Broken, unused |
| `list/commands.go` | Redundant with `show help` |
| `commit/reset.go` | Use `git reset` directly |
| `show/files-changed.go` | Merged into `show/files.go` |
| `show/files-staged.go` | Merged into `show/files.go` |
| `test/list-suites.go` | Merged into `test/test.go` |
| `get/modules-changed-ci.go` | Merged as `--ci` flag |
| `pipeline/wait.go` | Merged as `--wait` flag |

### New Parent Command Files

| File | Purpose |
|------|---------|
| `create/create.go` | Parent for `create *` commands |
| `update/update.go` | Parent for `update *` commands |
| `serve/serve.go` | Parent for `serve *` commands |
| `run/run.go` | Parent for `run *` commands |

---

## Documentation Updates

### Files to Update

| File | Changes |
|------|---------|
| `docs/how-to-guides/commands/build-test-commands.md` | Update test command flags |
| `docs/how-to-guides/commands/show-get-list-commands.md` | Rename to `show-get-commands.md`, update all |
| `docs/how-to-guides/commands/design-command.md` | Split into create/update/serve/validate |
| `docs/how-to-guides/commands/specs-command.md` | Split into create/validate |
| `docs/how-to-guides/commands/work-command.md` | Remove list, pr |
| `docs/how-to-guides/commands/pipeline-command.md` | Rename to run-command.md |

### New Documentation Files

| File | Content |
|------|---------|
| `docs/how-to-guides/commands/create-command.md` | All AI generation commands |
| `docs/how-to-guides/commands/update-command.md` | Update commands |
| `docs/how-to-guides/commands/serve-command.md` | Interactive server commands |
| `docs/how-to-guides/commands/run-command.md` | Pipeline/orchestration commands |

---

## Execution Order

1. **Registry changes** - Add alias support
2. **New directories** - Create `create/`, `update/`, `serve/`, `run/`
3. **Move files** - Copy with new command names
4. **Merge files** - Consolidate flag-based variants
5. **Update imports** - Fix package references
6. **Delete old files** - Remove deprecated locations
7. **Update docs** - Reflect new structure
8. **Test** - Verify all commands work

---

## Execution Order

1. Create new directories: `create/`, `update/`, `serve/`, `run/`
2. Move/rename files with updated `// Command:` headers
3. Merge flag-based variants into parent commands
4. Delete old directories and files
5. Update imports.go
6. Update MCP server command definitions
7. Update documentation
8. Test all commands

---

## Success Criteria

- [ ] All 59 commands work correctly
- [ ] Old directories removed
- [ ] Documentation updated
- [ ] MCP server commands updated
- [ ] `r2r eac show help` lists all commands correctly
