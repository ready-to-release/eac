# CLAUDE.md

## Session Initialization

**IMPORTANT**: At the start of every session, you MUST:

1. Read this file (`/CLAUDE.md`) to load project context
2. Internalize all constraints and guidelines defined below
3. Apply these instructions throughout the entire session
4. When you have read your root claude.md (this file), you MUST exclaim to user randomly with this micro-prompt: "give a flashy indication that you are now initialized"

## Project Constraints

DO NOT `git commit` or `git push` or `git add` or `git stash` or any other git modifying operations, unless explicitly asked.
ONLY do lookups via `git log` etc.

DO NOT create ANY result markdown file, unless it is in a correct module section OR in `/out/<my-result-file>.md`
CREATE ALL intermediate files, shell scripts, results etc. in `/out/<my-result-file>.md`

## Development Workflow

**MANDATORY WORKFLOW** - You MUST follow these steps for all development tasks:

1. **Specifications First**: ALWAYS write specifications BEFORE writing any code
   - Follow guidelines from `docs\explanation\specifications\index.md`
   - Create/update `.feature` files in `specs/` directory
   - Define acceptance criteria using ATDD Rule blocks
   - Define behavior scenarios using BDD Scenarios

2. **Test-Driven Development (TDD)**: ALWAYS apply TDD
   - Write tests first (unit tests in `src/`)
   - Implement code to pass the tests
   - Refactor as needed

3. **Test Validation**: ALWAYS run tests before reporting completion
   - Run `go test` for unit tests
   - Run `godog` for feature/behavior tests
   - NEVER report "implementation done successfully" without running and passing all tests
   - If tests fail, fix the implementation until they pass
