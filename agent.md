# Agent

## Session Initialization

**IMPORTANT**: At the start of every session, you MUST:

1. **Detect current workspace** by checking the git branch in the current working directory
2. **Verify workspace context**: Check if the current directory path matches the detected branch (may be a mismatch in multi-worktree setups)
3. **Read this file** (`/agent.md`) to load project context
4. **Load MCP server capabilities** by reading available MCP command tools
5. **Internalize all constraints and guidelines** defined below
6. **Apply these instructions** throughout the entire session
7. **Confirm initialization** with a flashy initialization report using this format:

```text
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  ⚡ SYSTEM INITIALIZED ⚡                                      ┃
┃  Project context loaded from agent.md                         ┃
┃  MCP servers: ACTIVE                                          ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

Workspace Context:
- Current branch: [branch name]
- Working directory: [current path]
- Status: [✓ Match / ⚠ MISMATCH - Expected path: [expected path for this branch]]

Project Context Loaded:

Active Constraints:
- Git: READ-ONLY by default. No commits/pushes/branches without explicit user request
- Multi-Worktree Aware: Operating in [current directory] ([branch])
- File Organization: Modules in /src, intermediate files in /out
- MCP-First: Always prefer mcp__commands__* tools over manual file operations

MCP Server Status:
- Module Discovery: ✓
- Dependency Management: ✓
- Architecture Docs: ✓
- Build & Test: ✓
- Git Operations: ✓
- Specifications: ✓
- Templates: ✓
- Workspace Management: ✓
- Documentation: ✓

Ready to assist with project tasks.
```

**Git Worktree Context**: This repository uses git worktrees for parallel development. The initialization process automatically detects the current branch based on the working directory. If there's a mismatch between the expected worktree path and the current directory, it will be highlighted in the initialization report.

### MCP Server Initialization

This project uses **MCP (Model Context Protocol) servers** to provide specialized commands for managing the modular monorepo architecture. During initialization, you MUST:

1. Recognize available `mcp__commands__*` tools
2. Understand their purpose and when to use them
3. Prefer MCP commands over manual file operations for module-related tasks

**Available MCP Command Categories:**

- **Module Discovery**: `get-modules`, `show-modules`, `show-moduletypes`, `get-files`, `show-files`
- **Dependency Management**: `get-dependencies`, `show-dependencies`, `validate-dependencies`, `get-execution-order`
- **Architecture Documentation**: `design-*` (Structurizr integration)
- **Build & Test**: `build-module`, `build-modules`, `test-module`, `test-modules`, `pipeline-run`
- **Documentation**: `docs-serve` (MkDocs integration)
- **Git Operations**: `commit-ai`, `show-files-changed`, `show-files-staged`, `get-changed-modules`
- **Specifications**: `specs-create`, `specs-validate`
- **Templates**: `templates-list`, `templates-install`, `templates-apply`
- **Workspace Management**: `work-create`, `work-list`, `work-commit`

---

## MCP-First Execution Policy

**ALWAYS use `mcp__commands__*` tools for project operations.**

## Project Constraints

### Git Operations

- **DO NOT** perform ANY git operations unless explicitly requested by the user
- **READ-ONLY git operations** (`log`, `status`, `diff`) are permitted ONLY when explicitly needed for information gathering
- **NEVER** run git modifying operations (`commit`, `push`, `add`, `stash`, `checkout`, `branch`, `merge`, `rebase`, `worktree`, etc.)
- **USER CONTROLS GIT**: The user manages all git operations manually, especially in multi-worktree setups
- **USE** `commit-ai` ONLY when the user explicitly requests commit message generation

#### Git Worktree Awareness

This repository may be running multiple Claude Code sessions in parallel using **git worktrees**. Each worktree is an isolated working directory on a different branch.

**When operating in a worktree environment:**

1. **Identify your location**: Check `git worktree list` (read-only) ONLY if needed to understand context
2. **Stay in your lane**: Work only on files in your current worktree directory
3. **No cross-worktree operations**: Never attempt to modify files in other worktrees
4. **User coordinates branches**: The user manages branch switching, merging, and synchronization
5. **Report your worktree**: When asked about git state, mention which worktree you're operating in

Each Claude session works independently. The user handles all git coordination.

### File Organization

- **Modules**: All modules are placed in the `src/` directory
- **Result files**: DO NOT create result markdown files except in:
  - Module directories (as identified by module contracts or `get-files`)
  - `/out/<my-result-file>.md` for intermediate/temporary files
- **Intermediate files**: CREATE all intermediate files, shell scripts, analysis results in `/out/` directory
- **Before modifications**: USE `get-files` **on-demand** to understand file ownership
  - WARNING: This command loads ~2690 files (~19k tokens). Only call when you need to determine module ownership before making changes.
  - Alternative: Use `show-files-changed` or `show-files-staged` for smaller, targeted queries

---

## Development Workflow

You are an AI coding agent contributing Go code to this repository.

Your work must follow a **mandatory three-phase workflow** guided by the **Three Rules of Vibe Coding**.

Everything you produce must increase clarity, reduce cognitive load, and enable fast, safe iteration.

### Guiding Principles: The Three Rules of Vibe Coding

All code you generate must embody these three rules:

#### Rule 1: Make the code easy to understand

Your outputs must be:

- Direct, explicit, and idiomatic Go
- Free of unnecessary abstractions, cleverness, or surprises
- Structured with minimal branching and clear data flow
- Written using clear, intention-revealing names
- Supported by comments only when explaining **why**, not **what**

**Agent behaviors:**

- Prefer small, single-purpose functions
- Keep files cohesive
- Avoid complexity unless it directly adds clarity

**If the team can understand your code in a single pass, you have succeeded.**

#### Rule 2: Make the code easy to change

Your code must be designed so future changes are safe and simple.

**Agent behaviors:**

- Break work into small, safe, incremental steps
- Produce complete, compilable code every time — no stubs or placeholders
- Use stable, predictable package boundaries
- Avoid deep dependency chains or hidden side effects
- Prefer pure functions where possible
- Use `context.Context` consistently for operations that may block or be canceled

**Good code gives the next developer freedom while minimizing mental load.**

#### Rule 3: Make it hard to break

All non-trivial code you generate must include tests.

**Agent behaviors:**

- Produce **table-driven unit tests** for all new logic
- Tests must be deterministic, fast, and free of external I/O
- Validate inputs; reject unexpected states early
- Use clear error messages and wrap errors with context
- Avoid concurrency unless needed, and when used, design so races are impossible

**Your code should fail safely and visibly when incorrect.**

---

### Three-Phase Development Process

**MANDATORY** - Follow these phases for all development tasks:

#### Phase 1: Specifications First

**Write specifications BEFORE writing any code**:

Specifications make the intended behavior easy to understand (Rule 1) and provide a contract that makes future changes safer (Rule 2).

**When specifications are required:**

- New features or functionality
- Changes to business logic
- Modifications to user-facing behavior
- Any non-trivial code changes

**For small changes** (bug fixes, typo corrections, minor refactoring):

1. Investigate if existing specifications need updates
2. Inform the user that you're not writing new specifications
3. Ask permission to continue without specifications
4. Proceed only after user approval

**Requirements when writing specifications:**

- **USE** `mcp__commands__specs-create` to generate new specifications from natural language descriptions
- **USE** `mcp__commands__specs-validate` to validate specifications against contracts before proceeding
- Create/update `.feature` files in `specs/` directory

#### Phase 2: Test-Driven Development (TDD)

**ALWAYS write tests BEFORE implementation**:

TDD embodies all three rules: tests document behavior (Rule 1), enable safe refactoring (Rule 2), and catch regressions (Rule 3).

**Test file organization:**

- **TDD unit tests**: Place `*_test.go` files alongside the code they test in module `src/` directories
- **BDD step definitions**: Place step implementation files in a dedicated `tests/` folder within each module
- **Feature files**: Place `.feature` files in the project's `specs/` directory

**Requirements:**

- Write tests first before any implementation
- Produce **table-driven unit tests** for all new logic
- Tests must be deterministic, fast, and free of external I/O
- Implement code to pass the tests
- Refactor to improve clarity and changeability

**Apply Vibe Coding principles in implementation:**

- **Easy to understand**: Use clear names, simple control flow, minimal abstraction
- **Easy to change**: Small functions, pure where possible, stable boundaries, no hidden state
- **Hard to break**: Input validation, early returns, clear errors, comprehensive tests

**Output format for code deliverables:**

Every code implementation must include:

1. **Intent**: One sentence describing what you are implementing or improving
2. **Design Explanation**: 2–5 bullets linking your design to the Three Rules of Vibe Coding
   - How does this make code **easy to understand**?
   - How does this make code **easy to change**?
   - How does this make code **hard to break**?
3. **Full Go Implementation**: Complete, compilable, idiomatic Go code blocks (no missing pieces, no pseudocode)
4. **Unit Tests**: Full table-driven tests in `*_test.go` files, runnable with `go test ./...`
5. **Run Instructions**: Commands to build and test, including relevant MCP commands

**Every deliverable must be ready to paste into the codebase without modification.**

#### Phase 3: Validation

**ALWAYS run all tests before reporting completion**:

Validation ensures your code actually works and is hard to break (Rule 3).

**Requirements:**

- Use `mcp__commands__test-module` for unit tests
- Use `mcp__commands__test-suite` for feature/behavior tests
- Use `mcp__commands__validate-dependencies` to check module contracts
- Use `mcp__commands__specs-validate` to validate specifications
- **NEVER** report "implementation done successfully" without running and passing all tests
- If tests fail, fix the implementation until they pass
- Verify code follows Go conventions (`gofmt`, `go vet`)

---

## Go-Specific Coding Rules

To align with idiomatic and maintainable Go:

- Go version: **≥ 1.21**
- Enforce: `gofmt`, `go vet`, idiomatic naming
- Use the standard library unless a dependency truly improves clarity
- Keep exported APIs minimal and intentional
- Prefer composition over inheritance
- Avoid global mutable state

## Go Code Structure Guidelines

To keep the codebase clear, maintainable, and aligned with idiomatic Go, follow these structural guidelines:

### File Length

Go does not enforce strict limits, but large files reduce clarity and increase cognitive load.

**Guidelines:**

- Prefer files in the range of **200–400 lines**
- Review files that grow beyond **600 lines**
- Avoid files over **1000 lines**, as they usually indicate missing modularity or unclear responsibilities

Files should be organized by **behavior and domain**, not by generic categories (avoid `utils.go`, `helpers.go`, etc.).

### Function Size & Modularity

Functions should be small, intention-revealing, and focused.

**Guidelines:**

- Aim for functions that fit on **one screen (~20–40 lines)**
- Revisit functions that exceed **50–80 lines**
- Avoid functions over **100 lines** unless genuinely necessary

Split functions when:

- They mix multiple concerns
- They require long comments to explain the flow
- They contain deeply nested logic

Prefer several small, well-named functions over one large “god function.”

### Responsibilities & Composition

- Keep files cohesive: one clear responsibility per file
- Use interfaces and small helpers to isolate behavior
- Prefer composition over deep inheritance or nested dependencies
- Avoid “god files” with many unrelated types or functions

**Goal:** Keep code easy to understand, easy to change, and hard to break—the same core principles that guide this agent’s workflow.
