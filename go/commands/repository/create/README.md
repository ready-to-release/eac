# create Commands

AI-powered artifact generation commands that produce commit messages, specifications,
designs, and risk documentation from repository content.

## Command Index

| Command | Package | Purpose |
| --- | --- | --- |
| `eac create commit-message` | [commit-message/](./commit-message/) | Generate AI-powered commit messages from staged changes |
| `eac create design` | [design/](./design/) | Generate Structurizr DSL workspace files using AI source analysis |
| `eac create risk-assess` | [risk-assess/](./risk-assess/) | Create OSCAL assessment-results from test and security evidence |
| `eac create risk-profile` | [risk-profile/](./risk-profile/) | Create OSCAL profile from risk assessment using AI |
| `eac create squash-message` | [squash-message/](./squash-message/) | Generate squash commit message from branch commits |

## Small Commands

Commands documented inline rather than in their own README.

### risk

Shared tag extraction utilities for OSCAL risk commands. Provides `tag_extractor.go`
which scans module source files for control tags used by `risk-assess` and `risk-profile`.

### spec

Generates Gherkin specifications from natural language descriptions using AI.
The command transforms feature descriptions into properly formatted Gherkin files
following Rules/Scenarios patterns suitable for BDD testing with Godog.

## Architecture Notes

Most create commands follow a common pattern: gather context from the repository,
construct an AI prompt with relevant source material, invoke the AI model, and write
the structured output. The `risk-assess` and `risk-profile` commands share tag extraction
logic from the `risk` package. The `commit-message` and `squash-message` commands both
analyze git history but differ in scope (staged changes vs. full branch).

### Common Patterns

- **AI-driven commands** (`commit-message`, `design`, `risk-profile`, `spec`): Use the
  `core/ai` package for prompt construction and model invocation
- **Evidence-based commands** (`risk-assess`): Aggregate existing test and scan artifacts
  rather than generating new content
- **Git-analysis commands** (`commit-message`, `squash-message`): Read git diff/log output
  and produce structured commit messages following conventional commit format
