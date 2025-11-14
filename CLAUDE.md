# CLAUDE.md

## Session Initialization

**IMPORTANT**: At the start of every session, you MUST:

1. **Read this file** (`/CLAUDE.md`) to load project context
2. **Load MCP server capabilities** by reading available MCP command tools
3. **Internalize all constraints and guidelines** defined below
4. **Apply these instructions** throughout the entire session
5. **Confirm initialization** with this micro-prompt: "give a flashy indication that you are now initialized"

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
- **Templates**: `templates-list`, `templates-install`

---

## Project Constraints

### Git Operations

- **DO NOT** perform git modifying operations (`commit`, `push`, `add`, `stash`) unless explicitly requested
- **ONLY** use git read operations (`log`, `status`, `diff`) for information gathering
- **USE** `mcp__commands__commit-ai` when generating commit messages (if explicitly requested)

### File Organization

- **Modules**: All modules are placed in the `src/` directory
- **Result files**: DO NOT create result markdown files except in:
  - Module directories (as identified by module contracts or `mcp__commands__get-files`)
  - `/out/<my-result-file>.md` for intermediate/temporary files
- **Intermediate files**: CREATE all intermediate files, shell scripts, analysis results in `/out/` directory
- **Before modifications**: USE `mcp__commands__get-files` to understand file ownership

---

## Development Workflow

You are an AI coding agent contributing Go code to this repository. Your work must follow a **mandatory three-phase workflow** guided by the **Three Rules of Vibe Coding**.

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

- Follow guidelines from `docs\explanation\specifications\index.md`
- Create/update `.feature` files in `specs/` directory
- Define acceptance criteria using **ATDD Rule blocks**
- Define behavior scenarios using **BDD Scenarios**
- Tag scenarios appropriately (see `docs\explanation\specifications\tag-reference.md`)

**MCP Integration:**

- Use `mcp__commands__show-modules` to understand module structure
- Use `mcp__commands__get-dependencies` to understand module relationships
- Use `mcp__commands__get-files` to understand file ownership before modifications

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

**MCP Integration:**

- Use `mcp__commands__test-module` to run tests for specific modules
- Use `mcp__commands__get-execution-order` to determine test execution sequence
- Use `mcp__commands__build-module` to validate compilation

#### Phase 3: Validation

**ALWAYS run all tests before reporting completion**:

Validation ensures your code actually works and is hard to break (Rule 3).

**Requirements:**

- Run `go test` for unit tests
- Run `godog` for feature/behavior tests
- **NEVER** report "implementation done successfully" without running and passing all tests
- If tests fail, fix the implementation until they pass
- Verify code follows Go conventions (`gofmt`, `go vet`)

**MCP Integration:**

- Use `mcp__commands__test-modules` for batch testing
- Use `mcp__commands__pipeline-run` to execute full module pipelines
- Use `mcp__commands__get-changed-modules` to scope test execution
- Use `mcp__commands__validate-dependencies` to check module contracts

---

## Go-Specific Coding Rules

To align with idiomatic and maintainable Go:

- Go version: **≥ 1.21**
- Enforce: `gofmt`, `go vet`, idiomatic naming
- Use the standard library unless a dependency truly improves clarity
- Keep exported APIs minimal and intentional
- Prefer composition over inheritance
- Avoid global mutable state

---

## MCP Commands Usage Guidelines

### When to Use MCP Commands

**ALWAYS use MCP commands when:**

1. Exploring or querying module information
2. Checking module dependencies or execution order
3. Working with architecture documentation (Structurizr)
4. Building or testing modules
5. Validating module contracts or dependencies
6. Understanding which files changed and which modules are affected
7. Generating commit messages for module changes
8. Managing project documentation

### Command Selection Strategy

**Use `show-*` commands** for human-readable output and reporting
**Use `get-*` commands** for structured data to process programmatically

### Typical Workflows

#### Before Making Changes

```text
1. mcp__commands__show-modules (understand structure)
2. mcp__commands__show-dependencies (understand relationships)
3. mcp__commands__get-files (understand file ownership)
```

#### During Development

```text
1. mcp__commands__build-module <moniker> (validate changes)
2. mcp__commands__test-module <moniker> (run tests)
3. mcp__commands__validate-dependencies (check contracts)
```

#### Before Committing (if explicitly requested)

```text
1. mcp__commands__show-files-staged (review changes)
2. mcp__commands__get-changed-modules (identify affected modules)
3. mcp__commands__commit-ai (generate commit message)
```

#### Architecture Documentation

```text
1. mcp__commands__design-list (list modules with docs)
2. mcp__commands__design-validate (validate workspace)
3. mcp__commands__design-serve (preview documentation)
```

### Best Practices

1. **Prefer MCP commands over manual file operations** when working with module metadata
2. **Chain commands** to build comprehensive understanding
3. **Validate before modifying** using validation commands
4. **Check changed modules** before running builds or tests to scope work appropriately
5. **Use execution order** to respect dependencies when processing multiple modules

---

## Principles to Optimize for (In Order)

1. **Clarity**
2. **Changeability**
3. **Safety**
4. **Small, incremental flow of value**
5. **Reducing cognitive load**
6. **Avoiding negative vibes**
   - No magic behavior
   - No fragile tests
   - No hidden state
   - No clever hacks

---

## Agent Mindset

- You are not writing for yourself — you are writing for the next developer
- Default to the simplest solution that fully solves the problem
- Every response should leave the codebase *better than you found it*
- Your work should increase trust in the system and make future changes easier
- Use available MCP tools to understand context before making changes
- Follow the mandatory workflow: **Specs → TDD → Validation**

---

## Quick Reference

### Three-Phase Workflow Checklist

**Phase 1: Specifications First**:

- [ ] Load MCP server capabilities
- [ ] Understand module structure (`show-modules`)
- [ ] Check dependencies (`show-dependencies`)
- [ ] Check if specifications needed (or ask permission to skip for small changes)
- [ ] Write specifications (`.feature` files in `specs/`)
- [ ] Define ATDD Rules (acceptance criteria)
- [ ] Define BDD Scenarios (behavior)

**Phase 2: Test-Driven Development**:

- [ ] Write TDD tests first (`*_test.go` alongside code in module `src/`)
- [ ] Write BDD step definitions (`tests/` folder in module)
- [ ] Implement code to pass tests (following Three Rules of Vibe Coding)
- [ ] Refactor for clarity and changeability
- [ ] Build module (`build-module`)
- [ ] Follow output format: Intent → Design → Implementation → Tests → Run Instructions

**Phase 3: Validation**:

- [ ] Run unit tests (`go test`)
- [ ] Run feature tests (`godog`)
- [ ] Validate with MCP commands (`validate-dependencies`)
- [ ] Ensure all tests pass before completion
- [ ] Verify code follows Go conventions (`gofmt`, `go vet`)

### Three-Layer Testing

- **ATDD** (Acceptance): Business requirements → Rule blocks → `.feature` files in `specs/`
- **BDD** (Behavior): User-facing behavior → Scenarios under Rules → Step definitions in module `tests/`
- **TDD** (Implementation): Code correctness → Go unit tests → `*_test.go` alongside code

### Key Documentation

- Specifications: `docs\explanation\specifications\index.md`
- Testing Strategy: `docs\explanation\continuous-delivery\testing\testing-strategy-overview.md`
- Tag Reference: `docs\explanation\specifications\tag-reference.md`
