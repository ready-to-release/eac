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

---

## Testing Specifications

This project uses a **three-layer testing approach** unified in Gherkin:

- **ATDD** (Acceptance Criteria as Rule blocks) → Godog
- **BDD** (Behavior Scenarios under Rules) → Godog
- **TDD** (Unit tests) → Go test

**Key Principles**:

1. ATDD and BDD are conceptually distinct layers but technically unified in a single `.feature` file using Gherkin's `Rule:` syntax
2. **Specifications (WHAT) vs Implementation (HOW)**: Specifications live in `specs/`, test implementations live in `src/`

**Full documentation**: [Specifications](docs\explanation\specifications\index.md)
