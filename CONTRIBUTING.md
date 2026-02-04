# Contributing to EAC

Thank you for your interest in contributing to EAC (Everything as Code)! This project follows a "Vibe Coding" philosophy, prioritizing clarity, changeability, and robustness.

## 🌟 Core Principles

Before you start, please internalize our core development rules (found in `agent.md`):

1. **Three Rules of Vibe Coding**:
    - **Easy to Understand**: Clear, simple, and explicit code.
    - **Easy to Change**: Small functions, stable boundaries, and no hidden state.
    - **Hard to Break**: High test coverage and continuous validation.
2. **Three-Phase Development**:
    - **Specifications First**: Write Gherkin `.feature` files before code.
    - **Test-Driven Development**: Write tests before implementation.
    - **Validation**: Run all tests and quality gates before reporting complete.

## 🚀 Getting Started

1. **Read the Onboarding Tutorials**: Start with our [Getting Started Tutorials](docs/tutorials/getting-started/index.md).
2. **Set Up Local Environment**: Follow the [Local Setup Guide](docs/how-to-guides/local-setup/install-toolchain.md).
3. **Source the Importer**: Always run `source ./importer.sh` in your terminal to set up necessary aliases (`eac`, `run`).

## 🛠 Development Workflow

We use the `eac` CLI for all repository maintenance.

- **Build**: `eac build <module-name>`
- **Test**: `eac test <module-name>`
- **Validation**: Always run `eac validate` before committing. This runs our quality gates (contracts, dependencies, specs, security, and tests).

For detailed command usage, see the [Command Reference](docs/how-to-guides/eac/commands/index.md).

## 🤖 AI-Assisted Contributions

This repository is optimized for AI-assisted development.

- If you are using **Junie**, refer to `junie/README.md` for specific agent rules.
- If you are using **Claude Code**, we have specialized agents and skills in `.claude/`.

## 📮 Pull Requests & Commits

- **Small & Atomic**: We prefer small, focused pull requests that address a single concern.
- **Semantic Commits**: Use clear, descriptive commit messages. We recommend using `eac work-commit` for AI-generated semantic messages.
- **Validation**: Ensure `eac validate` passes locally before submitting a PR.

## ⚖️ Legal

By contributing to this project, you agree that your contributions will be licensed under the project's [LICENSE](LICENSE).
