# ACD and the CD Model

This article maps the [Agentic CD framework](https://migration.minimumcd.org/docs/agentic-cd/) onto this project's CD Model. For ACD definitions and the adoption roadmap, read the source. This article adds only the project-specific mapping.

---

## Prerequisites

The base CD Model (Stages 2–5) must be established before ACD applies. See the [ACD Adoption Roadmap](https://migration.minimumcd.org/docs/agentic-cd/) on the source site for the phased adoption sequence.

---

## Where the Six Artifacts Live

ACD requires six first-class artifacts per change. This table shows where each lives in this project:

| Artifact             | Authority                                | Where it lives in this project                                          |
| -------------------- | ---------------------------------------- | ----------------------------------------------------------------------- |
| System Constraints   | Organization                             | Pipeline gates: `.golangci.yml`, SAST, contract validation              |
| Intent Description   | Human (via create-spec prompt)           | `# Intent:` comment at top of `specification.feature`, AI-derived       |
| Feature Description  | Engineering (via C4 model + description) | `# Architecture:` comment at top of `specification.feature`, AI-derived |
| User-Facing Behavior | Human defines; agent generates           | Feature + Rules + Scenarios in `specification.feature`                  |
| Executable Truth     | Pipeline                                 | `go/**/*_test.go` + CI stages 2–5                                       |
| Implementation       | Agent / developer                        | `go/**/*.go`                                                            |

Key observations:

- All three human-owned artifacts live in one file. The engineer's description drives the intent; the C4 model provides the architectural context; the AI derives and writes both into the spec before writing any scenarios.
- `validate-specs` enforces that `# Intent:` and `# Architecture:` are present and non-empty. Missing comments are a pipeline failure, not a style suggestion.
- The Structurizr C4 model (`specs/<module>/.design/workspace.dsl`) is the architectural source of truth the AI reads to derive `# Architecture:`. It is not the per-change artifact — the comment in the spec file is.

---

## The Specification File as the Agent's Brief

The `specification.feature` file is structured so that the agent implementing the feature reads intent and architecture first, before the scenarios:

```gherkin
# Intent: Add pre-commit validation for AI-generated commit messages so that non-conforming messages are rejected before they reach the repository
# Architecture: Affects eac work-commit container; reads git staged changes; calls commit message validator; depends on validation/formats/commit package and git adapter

@deps:go @L2 @ov @env:isolated-test-project
Feature: eac-work_work-commit

  As a developer working in a workspace
  I want to commit changes with AI-generated messages
  So that my commits have consistent, high-quality messages

  Rule: Validation prevents invalid commits
    ...
```

The `# Intent:` and `# Architecture:` lines are the agent's brief. They communicate the problem and constraints before the scenarios provide the concrete behavior contract.

---

## The Eight Constraints in Our Pipeline

Each ACD constraint maps to existing mechanisms, with identified gaps:

| #   | Constraint                                             | How enforced                                                        | Gap                                                          |
| --- | ------------------------------------------------------ | ------------------------------------------------------------------- | ------------------------------------------------------------ |
| 1   | Every change has explicit, human-owned intent          | `# Intent:` in spec file; engineer's description drives it          | Not human-written directly — AI-derived from description     |
| 2   | Intent and architecture are first-class artifacts      | Versioned in `specification.feature` in VCS                         | —                                                            |
| 3   | All artifacts versioned and delivered with each change | Git tracks spec file                                                | No pipeline check that comments match the change             |
| 4   | Intended behavior independent of implementation        | `specification.feature` scenarios are spec-derived                  | Test decoupling checked in Stage 3 review only               |
| 5   | Consistency between artifacts enforced                 | CI gates (tests, linting); `validate-specs` checks comment presence | No semantic consistency check between comments and scenarios |
| 6   | Agent changes comply with documented constraints       | Stage 2 SAST/lint; Stage 3 review                                   | Constraints in `# Architecture:` not yet machine-checked     |
| 7   | Agents cannot promote their own changes                | Stage 9 RA: human approval                                          | RA variant satisfies; CDe requires additional gate           |
| 8   | Pipeline red → agents fix only                         | Process constraint                                                  | Not mechanically enforced                                    |

---

## Expert Agents in Our Pipeline

ACD defines expert agent roles that inspect changes at specific stages. The `# Intent:` and `# Architecture:` comments in the spec file are exactly the inputs these agents need — they are already versioned in the repository.

| Expert Agent Role         | Fits in our pipeline at        | Input artifact            |
| ------------------------- | ------------------------------ | ------------------------- |
| Test fidelity             | Stage 3 MR; Stage 5 acceptance | `# Intent:` + scenarios   |
| Implementation coupling   | Stage 2 or Stage 3             | Test files                |
| Architectural conformance | Stage 3 MR                     | `# Architecture:` comment |
| Intent alignment          | Stage 3 MR; Stage 9            | `# Intent:` comment       |
| Constraint compliance     | Stage 2 (static); Stage 3      | System constraint gates   |

---

## Cross-references

- [The 12 Stages](../cd-model/stages.md)
- [Pre-commit Quality Gates](../quality-gates/precommit-gates.md)
- [Merge Request Quality Gates](../quality-gates/merge-request-gates.md)
- [AI-Assisted Detection](../quality-gates/defect-prevention/ai-detection.md)
