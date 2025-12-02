<!-- markdownlint-disable MD033 MD045 -->
#

<p align="center">
  <img src="../docs/assets/logo/eac-logo.png" width="200" /><br>
  <strong style="font-size: 2em;">Everything as Code</strong><br>
  <em>Continuous delivery for regulated industries—without the compliance bottleneck</em>
</p>

<p align="center">
  <a href="https://github.com/ready-to-release/eac/actions/workflows/trigger-ci.yaml"><img src="https://github.com/ready-to-release/eac/actions/workflows/trigger-ci.yaml/badge.svg" alt="CI Status"></a>
  <a href="https://github.com/ready-to-release/eac/stargazers"><img src="https://img.shields.io/github/stars/ready-to-release/eac?style=social" alt="GitHub Stars"></a>
</p>

<p align="center">
  <a href="https://ready-to-release.github.io/eac/"><img src="https://img.shields.io/badge/docs-ready--to--release.github.io%2Feac-blue" alt="docs"></a>
  <a href="https://github.com/ready-to-release/eac/releases?q=r2r-cli"><img src="https://img.shields.io/github/v/tag/ready-to-release/eac?filter=r2r-cli/*&label=r2r-cli&color=green" alt="r2r-cli release"></a>
  <a href="https://github.com/ready-to-release/eac/pkgs/container/ext-eac"><img src="https://img.shields.io/github/v/tag/ready-to-release/eac?filter=ext-eac/*&label=ext-eac&color=green" alt="ext-eac release"></a>
  <a href="https://github.com/ready-to-release/eac/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-yellow" alt="MIT license"></a>
</p>

---

## The Problem

Regulated industries face a paradox: **compliance demands rigor, markets demand speed**.

[DORA's research on change approval](https://dora.dev/capabilities/streamlining-change-approval/) shows that heavyweight approval processes backfire—they "slow down the delivery process leading to the release of larger batches less frequently," which *increases* production risk. External review boards show no correlation with lower failure rates.

Traditional approaches force a choice—move fast and risk audit failures, or stay compliant and watch competitors ship.
Manual documentation drifts.
Traceability is reconstructed after the fact.
Audit prep takes months.

**This is a false choice.**

---

## The Solution

[DORA's Version Control capability](https://dora.dev/capabilities/version-control/) proves that elite performers store **everything** in version control—not just code, but configurations, infrastructure, specifications, tests, and deployment automation.

This enables complete traceability: "the path backward from every deployment to the elements it came from."

We implement this with two complementary ideas:

### R2R — Ready to Release

> *"We want to be ready to release at any time, for any team."*

An extensible CLI that isolates teams from tooling and platform dependencies. Through Docker encapsulation, R2R runs almost entirely independent of local setups—ensuring every commit is validated, every artifact traceable, and every deployment repeatable.

### EAC — Everything as Code

> *"We treat everything as code—to trace, version, review, and automate."*

A paradigm where requirements, documentation, compliance rules, and architecture live as executable, version-controlled artifacts. Not documents that drift—code that runs. Full traceability. Full auditability. Always.

The EAC extension helps organizations design reusable templates for apps, tools, workflows, and repositories—standardizing delivery across teams.

---

## How It Works

| What Changes       | From                             | To                                |
| ------------------ | -------------------------------- | --------------------------------- |
| **Requirements**   | Word documents                   | Executable Gherkin specifications |
| **Documentation**  | Wiki pages that drift            | Audit-ready, generated from code  |
| **Compliance**     | Manual checklists                | Automated pipeline gates          |
| **Traceability**   | Spreadsheets compiled for audits | Git history—always complete       |
| **Audit evidence** | Months of preparation            | Continuous capture—always ready   |

**The result**: Audit-ready documentation as a byproduct of your pipeline, not a bottleneck before release.

[Understand the paradigm →](https://ready-to-release.github.io/eac/explanation/everything-as-code/) · [See the compliance-velocity paradox →](https://ready-to-release.github.io/eac/explanation/everything-as-code/compliance-velocity-paradox/)

---

## Get Started

### 1. Install R2R

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

### 2. Initialize and install EAC extension

<details>
<summary><b>Commands</b></summary>

```bash
# In your repository root
r2r init
r2r install eac
# Creates .r2r/r2r-cli.yml
```

</details>

### 3. Initialize EAC in your repository

<details>
<summary><b>Commands</b></summary>

```bash
r2r eac init
# Creates .r2r/eac/modules.yml, module-types.yml, ...
```

</details>

### 4. Start using it

<details>
<summary><b>Commands</b></summary>

```bash
r2r eac show modules              # Discover module structure
r2r eac create spec "..."         # Generate executable specifications
r2r eac validate contracts        # Validate compliance continuously
```

</details>

[See all 67+ commands →](https://ready-to-release.github.io/eac/how-to-guides/)

---

## Why It Works

[DORA research](https://dora.dev/research/) proves that high performers in **regulated industries** achieve the same metrics as tech companies:

- **208×** faster deployment frequency
- **106×** faster lead time
- **7×** lower change failure rate

The difference isn't less compliance—it's **automated compliance**. Small, validated changes. Continuous evidence capture. Traceability built into every commit.

**This repository is a demonstration of the tooling in action**: We use R2R and EAC to build R2R and EAC.

---

## Extend It

<details>
<summary><b>Create R2R extensions</b></summary>

```yaml
# .r2r/r2r-cli.yml
extensions:
  - name: 'my-extension'
    image: 'ghcr.io/my-org/my-extension:latest'
```

[Full guide →](https://ready-to-release.github.io/eac/how-to-guides/r2r/creating-extensions/)

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

[Full guide →](https://ready-to-release.github.io/eac/how-to-guides/eac/creating-modules/)

</details>

---

## Documentation

| I want to...               | Go to                                        |
| -------------------------- | -------------------------------------------- |
| Learn step-by-step         | [Tutorials](https://ready-to-release.github.io/eac/tutorials/)         |
| Accomplish a specific task | [How-to Guides](https://ready-to-release.github.io/eac/how-to-guides/) |
| Look up technical details  | [Reference](https://ready-to-release.github.io/eac/reference/)         |
| Understand why things work | [Explanation](https://ready-to-release.github.io/eac/explanation/)     |

**Full documentation**: [ready-to-release.github.io/eac](https://ready-to-release.github.io/eac/)

---

## Maintainers

- Casper Leon Nielsen ([@casperease](https://github.com/casperease))
- Mikael Ottesen Hansen ([@miohansen](https://github.com/miohansen))

**Questions?** [Open an issue](https://github.com/ready-to-release/eac/issues/new) · **License**: [MIT](LICENSE) (code) / [CC BY-SA 4.0](docs/LICENSE) (docs)
