<p align="center">
  <img src="../docs/assets/logo/eac-logo.png" width="200" /><br>
  <strong style="font-size: 2em;">Everything as Code</strong><br>
  <em>Continuous delivery for regulated industries—without the compliance bottleneck</em>
</p>

<p align="center">
  <a href="https://github.com/ready-to-release/eac/actions/workflows/change-trigger.yaml"><img src="https://github.com/ready-to-release/eac/actions/workflows/change-trigger.yaml/badge.svg" alt="CI Status"></a>
  <a href="https://github.com/ready-to-release/eac/stargazers"><img src="https://img.shields.io/github/stars/ready-to-release/eac?style=social" alt="GitHub Stars"></a>
</p>

<p align="center">
  <a href="https://ready-to-release.github.io/eac/"><img src="https://img.shields.io/badge/docs-ready--to--release.github.io%2Feac-blue" alt="docs"></a>
  <a href="https://github.com/ready-to-release/eac/releases?q=src-cli"><img src="https://img.shields.io/github/v/tag/ready-to-release/eac?filter=src-cli/*&label=src-cli&color=green" alt="src-cli release"></a>
  <a href="https://github.com/ready-to-release/eac/pkgs/container/ext-eac"><img src="https://img.shields.io/github/v/tag/ready-to-release/eac?filter=ext-eac/*&label=ext-eac&color=green" alt="ext-eac release"></a>
  <a href="https://github.com/ready-to-release/eac/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-yellow" alt="MIT license"></a>
</p>

---

## The Problem

Regulated industries face a paradox: **compliance demands rigor, markets demand speed**.

Traditional approaches force a choice—move fast and risk audit failures, or stay compliant and watch competitors ship. Manual documentation drifts. Traceability is reconstructed after the fact. Audit prep takes months. Release cycles stretch to weeks.

**This is a false choice.**

---

## The Solution

Two complementary ideas that solve the paradox together:

### r2r — Ready to Release

> *"We want to be ready to release at any time, for any team."*

An extensible CLI that makes release-readiness the default state. Every commit validated. Every artifact traceable. Every deployment repeatable.

### EAC — Everything as Code

> *"We treat everything as code—to trace, version, review, and automate."*

A paradigm where requirements, documentation, compliance rules, and architecture live as executable, version-controlled artifacts. Not documents that drift—code that runs.

---

## How It Works

| What Changes | From | To |
|--------------|------|-----|
| **Requirements** | Word documents | Executable Gherkin specifications |
| **Documentation** | Wiki pages that drift | Generated from code and tests |
| **Compliance** | Manual checklists | Automated pipeline gates |
| **Traceability** | Spreadsheets compiled for audits | Git history—always complete |
| **Audit evidence** | Months of preparation | Continuous capture—always ready |

**The result**: Compliance becomes a byproduct of your pipeline, not a bottleneck before release.

[Understand the paradigm →](docs/explanation/everything-as-code/index.md) · [See the compliance-velocity paradox →](docs/explanation/everything-as-code/compliance-velocity-paradox.md)

---

## Get Started

### 1. Install r2r

<details>
<summary><b>Linux / macOS</b></summary>

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

</details>

<details>
<summary><b>Windows (PowerShell)</b></summary>

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

</details>

### 2. Install the EAC extension

<details>
<summary><b>Commands</b></summary>

```bash
r2r install eac
```

</details>

### 3. Start using it

<details>
<summary><b>Commands</b></summary>

```bash
r2r init                          # Initialize configuration in your repo
r2r eac show modules              # Discover module structure
r2r eac specs create "..."        # Generate executable specifications
r2r eac validate contracts        # Validate compliance continuously
```

</details>

[See all 67+ commands →](docs/how-to-guides/index.md)

---

## Why It Works

[DORA research](https://dora.dev/research/) proves that high performers in **regulated industries** achieve the same metrics as tech companies:

- **208×** faster deployment frequency
- **106×** faster lead time
- **7×** lower change failure rate

The difference isn't less compliance—it's **automated compliance**. Small, validated changes. Continuous evidence capture. Traceability built into every commit.

**This repository is the proof.** We use r2r and EAC to build r2r and EAC. Study `.github/workflows/`, `specs/`, and `.r2r/eac/` to see it in action.

---

## Extend It

<details>
<summary><b>Create r2r extensions</b></summary>

```yaml
# .r2r/r2r-cli.yml
extensions:
  - name: 'my-extension'
    image: 'ghcr.io/my-org/my-extension:latest'
```

[Full guide →](docs/how-to-guides/r2r/creating-extensions.md)

</details>

<details>
<summary><b>Add EAC modules</b></summary>

```yaml
# .r2r/eac/modules.yml
modules:
  - moniker: my-service
    type: go-library
    files:
      root: src/my-service
```

[Full guide →](docs/how-to-guides/eac/creating-modules.md)

</details>

---

## Documentation

| I want to... | Go to |
|--------------|-------|
| Learn step-by-step | [Tutorials](docs/tutorials/index.md) |
| Accomplish a specific task | [How-to Guides](docs/how-to-guides/index.md) |
| Look up technical details | [Reference](docs/reference/index.md) |
| Understand why things work | [Explanation](docs/explanation/index.md) |

**Full documentation**: [ready-to-release.github.io/eac](https://ready-to-release.github.io/eac/)

---

## Maintainers

- Casper Leon Nielsen ([@casperease](https://github.com/casperease))
- Mikael Ottesen Hansen ([@miohansen](https://github.com/miohansen))

**Questions?** [Open an issue](https://github.com/ready-to-release/eac/issues/new) · **License**: [MIT](LICENSE) (code) / [CC BY-SA 4.0](docs/LICENSE) (docs)
