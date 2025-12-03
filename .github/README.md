<!-- markdownlint-disable MD033 MD045 -->
#

<p align="center">
  <img src="../docs/assets/logo/eac-logo.png" width="200" /><br><br>
  <sub><i>continuous integration</i></sub><br>
  <strong style="font-size: 1.5em;">Ready to Release</strong><br>
  <sub><i>continuous delivery</i></sub><br>
  <strong style="font-size: 1.5em;">Everything as Code</strong><br>
  <sub><i>continuous improvement</i></sub><br><br>
  <em>Uncompromising Continuous Delivery for regulated industries</em>
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

<h2 align="center">The Problem</h2>

<p align="center">
<img src="https://img.shields.io/badge/⚠️_THE_PARADOX-Compliance_demands_rigor._Markets_demand_speed.-dc2626?style=for-the-badge&labelColor=7f1d1d" alt="The Paradox">
</p>

<p align="center">
🎯 <strong>No single source of truth</strong> — version conflicts everywhere<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">validate</a> · <a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/paradigm/">paradigm</a></sup>
</p>
<p align="center">
📄 <strong>Docs drift</strong> — wiki pages go stale<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-overview/">books</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">templates</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">evidence</a></sup>
</p>
<p align="center">
🔍 <strong>Traceability lost</strong> — reconstructed after the fact<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">pipeline</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/">release</a> · <a href="https://ready-to-release.github.io/eac/explanation/continuous-delivery/workflow/branching-strategies/">trunk</a></sup>
</p>
<p align="center">
⏰ <strong>Audit prep</strong> — takes months, not minutes<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">security</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/">risk</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-overview/">design</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/">specs</a></sup>
</p>
<p align="center">
📦 <strong>Evidence scattered</strong> — proof lives in 10 different tools<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">validate</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/">how-to guides</a></sup>
</p>
<p align="center">
🛡️ <strong>Security as afterthought</strong> — scans run too late to matter<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">security</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-commands/">commands</a></sup>
</p>
<p align="center">
🔗 <strong>Requirements disconnect</strong> — specs live apart from tests<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/">specifications</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-commands/">commands</a></sup>
</p>
<p align="center">
🏛️ <strong>Knowledge silos</strong> — tribal wisdom, not shared systems<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-overview/">design</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">templates</a> · <a href="https://ready-to-release.github.io/eac/">docs</a></sup>
</p>
<p align="center">
🚀 <strong>Releases are events</strong> — big bang deployments, fingers crossed<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/">release</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">pipeline</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-commands/">changelog</a></sup>
</p>
<p align="center">
📋 <strong>Copy-paste compliance</strong> — same docs, different date<br>
<sup>solved → <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/">risk</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">templates</a></sup>
</p>

<p align="center">
<a href="https://dora.dev/capabilities/streamlining-change-approval/">DORA research</a> proves heavyweight approvals backfire—they <em>increase</em> risk.<br>
<strong>Elite teams prove you can have both — speed <em>and</em> compliance.</strong>
</p>

---

<h2 align="center">The Solution</h2>

<p align="center">
<a href="https://dora.dev/capabilities/version-control/">DORA</a> proves elite performers store <strong>everything</strong> in version control.<br>
<sub>Not just code—configurations, infrastructure, specifications, tests, and deployment automation.</sub>
</p>

<table align="center">
<tr>
<td align="center" valign="top">
<img src="https://img.shields.io/badge/🚀-R2R-3b82f6?style=for-the-badge" alt="R2R"><br>
<strong>Ready to Release</strong><br>
<em>"Ready to release at any time, for any team."</em><br><br>
<sub>
Extensible CLI that isolates teams from tooling complexity.<br>
Docker encapsulation ensures platform independence.<br>
Every commit validated. Every artifact traceable.
</sub><br><br>
<sup><a href="https://ready-to-release.github.io/eac/tutorials/">Install →</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/r2r/creating-extensions/">Extend →</a></sup>
</td>
<td align="center" valign="top">
<img src="https://img.shields.io/badge/📦-EAC-8b5cf6?style=for-the-badge" alt="EAC"><br>
<strong>Everything as Code</strong><br>
<em>"Trace, version, review, and automate everything."</em><br><br>
<sub>
Requirements, docs, compliance as executable artifacts.<br>
Not documents that drift—code that runs.<br>
Full traceability. Full auditability. Always.
</sub><br><br>
<sup><a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/">Paradigm →</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/">Commands →</a></sup>
</td>
</tr>
</table>

<p align="center">
<strong>One source of truth. One workflow. One audit trail.</strong>
</p>

---

<h2 align="center">How It Works</h2>

<p align="center">
One commit. Validated build. Auditable release.
</p>

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/📋-Requirements-22c55e?style=for-the-badge" alt="Requirements"><br>
<sub><del>Word documents</del></sub><br>
<strong>Executable specs</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/">Specs →</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/📄-Documentation-14b8a6?style=for-the-badge" alt="Documentation"><br>
<sub><del>Wiki pages that drift</del></sub><br>
<strong>Generated from code</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">Templates →</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/✅-Compliance-a855f7?style=for-the-badge" alt="Compliance"><br>
<sub><del>Manual checklists</del></sub><br>
<strong>Automated gates</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">Validate →</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/🔗-Traceability-ec4899?style=for-the-badge" alt="Traceability"><br>
<sub><del>Audit spreadsheets</del></sub><br>
<strong>Git history</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">Pipeline →</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/🛡️-Audit-f97316?style=for-the-badge" alt="Audit"><br>
<sub><del>Months of prep</del></sub><br>
<strong>Always ready</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">Security →</a></sup>
</td>
</tr>
</table>

<p align="center">
<strong>The result</strong>: Audit-ready documentation as a byproduct of your pipeline, not a bottleneck before release.<br>
<a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/">Understand the paradigm →</a> · <a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/compliance-velocity-paradox/">See the compliance-velocity paradox →</a>
</p>

---

<h2 align="center">Get Started</h2>

<p align="center">
Up and running in under 5 minutes.
</p>

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/1-Install_R2R-3b82f6?style=for-the-badge" alt="Step 1"><br>
<sub>CLI for your platform</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/tutorials/">Linux/macOS</a> · <a href="https://ready-to-release.github.io/eac/tutorials/">Windows</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/2-Setup-10b981?style=for-the-badge" alt="Step 2"><br>
<code>r2r init</code><br>
<code>r2r install eac</code>
</td>
<td align="center">
<img src="https://img.shields.io/badge/3-Initialize-8b5cf6?style=for-the-badge" alt="Step 3"><br>
<code>r2r eac init</code><br>
<sub>Creates module contracts</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/4-Build-f43f5e?style=for-the-badge" alt="Step 4"><br>
<code>r2r eac build</code><br>
<code>r2r eac test</code>
</td>
</tr>
</table>

<p align="center">
<details>
<summary><b>View install commands</b></summary>

**Linux / macOS**
```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

</details>
</p>

<p align="center">
<strong>That's it.</strong> Your first validated build is one <code>r2r eac build</code> away.
</p>

---

<h2 align="center">Command Areas</h2>

<p align="center">
<sub>12 command areas organized by delivery flow.</sub>
</p>

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/🏗️-Design-8b5cf6?style=for-the-badge" alt="Design"><br>
<sub>C4 architecture diagrams</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/📋-Specs-22c55e?style=for-the-badge" alt="Specs"><br>
<sub>BDD & Gherkin</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/📄-Templates-14b8a6?style=for-the-badge" alt="Templates"><br>
<sub>Doc generation</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-commands/">commands</a></sup>
</td>
</tr>
<tr>
<td align="center">
<img src="https://img.shields.io/badge/🔨-Build-3b82f6?style=for-the-badge" alt="Build"><br>
<sub>Module compilation</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/🧪-Test-10b981?style=for-the-badge" alt="Test"><br>
<sub>Unit & integration</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/✅-Validate-a855f7?style=for-the-badge" alt="Validate"><br>
<sub>Contracts & schemas</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-commands/">commands</a></sup>
</td>
</tr>
<tr>
<td align="center">
<img src="https://img.shields.io/badge/🔒-Security-eab308?style=for-the-badge" alt="Security"><br>
<sub>SAST, SBOM & scanning</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/🛡️-Risk-f97316?style=for-the-badge" alt="Risk"><br>
<sub>OSCAL compliance</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/📚-Books-6366f1?style=for-the-badge" alt="Books"><br>
<sub>PDF & EPUB generation</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-commands/">commands</a></sup>
</td>
</tr>
<tr>
<td align="center">
<img src="https://img.shields.io/badge/🔀-Workspace-0ea5e9?style=for-the-badge" alt="Workspace"><br>
<sub>Git worktrees</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/⚡-Pipeline-ec4899?style=for-the-badge" alt="Pipeline"><br>
<sub>CI/CD orchestration</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-commands/">commands</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/🚀-Release-f43f5e?style=for-the-badge" alt="Release"><br>
<sub>CalVer & SemVer tagging</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-commands/">commands</a></sup>
</td>
</tr>
</table>

<p align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/"><strong>Browse all commands →</strong></a>
</p>

---

<h2 align="center">Why It Works</h2>

<p align="center">
<a href="https://dora.dev/research/">DORA research</a> proves regulated industries achieve elite performance:
</p>

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/208×-Faster_Deploys-22c55e?style=for-the-badge" alt="208x"><br>
<sub>Deployment frequency</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/106×-Faster_Lead_Time-3b82f6?style=for-the-badge" alt="106x"><br>
<sub>Commit to production</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/7×-Lower_Failure_Rate-8b5cf6?style=for-the-badge" alt="7x"><br>
<sub>Change failure rate</sub>
</td>
</tr>
</table>

<p align="center">
The difference isn't less compliance—it's <strong>automated compliance</strong>.<br>
<sub>Small validated changes · Continuous evidence capture · Traceability in every commit</sub>
</p>

<p align="center">
<img src="https://img.shields.io/badge/🐕-Dogfooding-f97316?style=flat-square" alt="Dogfooding"><br>
<sub>This repository uses R2R and EAC to build itself—living proof of the paradigm.</sub>
</p>

---

<h2 align="center">Extend It</h2>

<p align="center">
Build once. Reuse everywhere.
</p>

<table align="center">
<tr>
<td align="center" valign="top">
<img src="https://img.shields.io/badge/🔌-R2R_Extensions-3b82f6?style=for-the-badge" alt="Extensions"><br><br>
<code>extensions:</code><br>
<code>&nbsp;&nbsp;- name: 'my-ext'</code><br>
<code>&nbsp;&nbsp;&nbsp;&nbsp;image: 'ghcr.io/...'</code><br><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/r2r/creating-extensions/">Create extensions →</a></sup>
</td>
<td align="center" valign="top">
<img src="https://img.shields.io/badge/📦-EAC_Modules-8b5cf6?style=for-the-badge" alt="Modules"><br><br>
<code>modules:</code><br>
<code>&nbsp;&nbsp;- moniker: my-svc</code><br>
<code>&nbsp;&nbsp;&nbsp;&nbsp;type: go-library</code><br><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/eac/creating-modules/">Add modules →</a></sup>
</td>
</tr>
</table>

<p align="center">
<strong>Your standards. Your tooling. One CLI.</strong>
</p>

---

<h2 align="center">Documentation</h2>

<p align="center">
Four paths to understanding—pick your learning style.
</p>

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/🎓-Tutorials-22c55e?style=for-the-badge" alt="Tutorials"><br>
<sub>Learn step-by-step</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/tutorials/">Start learning →</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/📖-How--to-3b82f6?style=for-the-badge" alt="How-to"><br>
<sub>Accomplish tasks</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/">Find guides →</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/📚-Reference-8b5cf6?style=for-the-badge" alt="Reference"><br>
<sub>Technical details</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/">Look it up →</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/💡-Explanation-f97316?style=for-the-badge" alt="Explanation"><br>
<sub>Understand why</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/explanation/">Deep dive →</a></sup>
</td>
</tr>
</table>

<p align="center">
<a href="https://ready-to-release.github.io/eac/"><strong>Full documentation →</strong></a>
</p>

---

<h2 align="center">The Crew</h2>

<p align="center">
Built by practitioners who ship to production daily.
</p>

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/👤-casperease-3b82f6?style=for-the-badge" alt="casperease"><br>
<sub>Core maintainer</sub><br>
<sup><a href="https://www.linkedin.com/in/casper-nielsen-1a9a208/">linkedin</a> · <a href="https://github.com/casperease">github</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/👤-miohansen-8b5cf6?style=for-the-badge" alt="miohansen"><br>
<sub>Core maintainer</sub><br>
<sup><a href="https://www.linkedin.com/in/mikaelottesenhansen/">linkedin</a> · <a href="https://github.com/miohansen">github</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/🤖-claude_code-f97316?style=for-the-badge" alt="claude code"><br>
<sub>Assistant developer</sub><br>
<sup><a href="https://claude.ai/code">claude.ai</a> · <a href="https://docs.anthropic.com/en/docs/claude-code">docs</a></sup>
</td>
<td align="center">
<img src="https://img.shields.io/badge/👤-tomasmalmsten-22c55e?style=for-the-badge" alt="tomasmalmsten"><br>
<sub>Contributor</sub><br>
<sup><a href="https://www.linkedin.com/in/tomasmalmsten/">linkedin</a></sup>
</td>
</tr>
</table>

<p align="center">
<a href="https://github.com/ready-to-release/eac/issues/new"><img src="https://img.shields.io/badge/💬-Questions%3F_Open_an_issue-64748b?style=flat-square" alt="Questions"></a><br>
<sub><a href="LICENSE">MIT</a> (code) · <a href="docs/LICENSE">CC BY-SA 4.0</a> (docs)</sub>
</p>
