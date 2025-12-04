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
  <a href="https://github.com/ready-to-release/eac/actions/workflows/trigger-ci.yaml"><img src="https://img.shields.io/github/actions/workflow/status/ready-to-release/eac/trigger-ci.yaml?style=flat-square&label=ci" alt="CI"></a>
  <a href="https://github.com/ready-to-release/eac/stargazers"><img src="https://img.shields.io/github/stars/ready-to-release/eac?style=flat-square" alt="Stars"></a>
  <a href="https://github.com/ready-to-release/eac/commits/main"><img src="https://img.shields.io/github/last-commit/ready-to-release/eac?style=flat-square" alt="Last Commit"></a>
  <a href="https://github.com/ready-to-release/eac/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-lightgray?style=flat-square" alt="MIT"></a>
</p>

<p align="center">
  <a href="https://github.com/ready-to-release/eac/releases?q=r2r-cli"><img src="https://img.shields.io/github/v/tag/ready-to-release/eac?filter=r2r-cli/*&style=for-the-badge&logo=gnubash&logoColor=white&label=r2r-cli&color=3b82f6" alt="r2r-cli"></a>
  &nbsp;
  <a href="https://github.com/ready-to-release/eac/pkgs/container/ext-eac"><img src="https://img.shields.io/github/v/tag/ready-to-release/eac?filter=ext-eac/*&style=for-the-badge&logo=docker&logoColor=white&label=ext-eac&color=8b5cf6" alt="ext-eac"></a>
  &nbsp;
  <a href="https://ready-to-release.github.io/eac/"><img src="https://img.shields.io/badge/docs-live-22c55e?style=for-the-badge&logo=materialformkdocs&logoColor=white" alt="docs"></a>
</p>

---

<h2 align="center">The Problem</h2>

<p align="center">
<a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/compliance-velocity-paradox/"><img src="https://img.shields.io/badge/⚠️_THE_PARADOX-Compliance_demands_rigor._Markets_demand_speed.-dc2626?style=for-the-badge&labelColor=7f1d1d" alt="The Paradox"></a>
</p>

<table align="center">
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/"><img src="https://img.shields.io/badge/🎯-No_Source_of_Truth-dc2626?style=for-the-badge" alt="No Source of Truth"></a><br>
<sub>Version conflicts everywhere</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">validate</a> · <a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/paradigm/">paradigm</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-overview/"><img src="https://img.shields.io/badge/📄-Docs_Drift-dc2626?style=for-the-badge" alt="Docs Drift"></a><br>
<sub>Wiki pages go stale</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-overview/">books</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">templates</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/"><img src="https://img.shields.io/badge/🔍-Traceability_Lost-dc2626?style=for-the-badge" alt="Traceability Lost"></a><br>
<sub>Reconstructed after the fact</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">pipeline</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/">release</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/"><img src="https://img.shields.io/badge/⏰-Audit_Prep-dc2626?style=for-the-badge" alt="Audit Prep"></a><br>
<sub>Takes months, not minutes</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">security</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/">risk</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/"><img src="https://img.shields.io/badge/📦-Evidence_Scattered-dc2626?style=for-the-badge" alt="Evidence Scattered"></a><br>
<sub>Proof lives in 10 different tools</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">validate</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/">guides</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/"><img src="https://img.shields.io/badge/🛡️-Security_Late-dc2626?style=for-the-badge" alt="Security Late"></a><br>
<sub>Scans run too late to matter</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">security</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-commands/">commands</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/"><img src="https://img.shields.io/badge/🔗-Specs_Disconnect-dc2626?style=for-the-badge" alt="Specs Disconnect"></a><br>
<sub>Requirements live apart from tests</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/">specs</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-overview/">test</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-overview/"><img src="https://img.shields.io/badge/🏛️-Knowledge_Silos-dc2626?style=for-the-badge" alt="Knowledge Silos"></a><br>
<sub>Tribal wisdom, not shared systems</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-overview/">design</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">templates</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/"><img src="https://img.shields.io/badge/🚀-Release_Events-dc2626?style=for-the-badge" alt="Release Events"></a><br>
<sub>Big bang deployments, fingers crossed</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/">release</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">pipeline</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/"><img src="https://img.shields.io/badge/📋-Copy--Paste_Compliance-dc2626?style=for-the-badge" alt="Copy-Paste Compliance"></a><br>
<sub>Same docs, different date</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/">risk</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">templates</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/"><img src="https://img.shields.io/badge/🚧-Manual_Gates-dc2626?style=for-the-badge" alt="Manual Gates"></a><br>
<sub>Approvals slow without adding safety</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">pipeline</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">validate</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/"><img src="https://img.shields.io/badge/🎲-Deployment_Roulette-dc2626?style=for-the-badge" alt="Deployment Roulette"></a><br>
<sub>Nobody knows what's in production</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/">release</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">sbom</a></sup>
</td>
</tr>
</table>

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
<a href="https://ready-to-release.github.io/eac/tutorials/"><img src="https://img.shields.io/badge/🚀-R2R-3b82f6?style=for-the-badge" alt="R2R"></a><br>
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
<a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/"><img src="https://img.shields.io/badge/📦-EAC-8b5cf6?style=for-the-badge" alt="EAC"></a><br>
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
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/"><img src="https://img.shields.io/badge/📋-Requirements-22c55e?style=for-the-badge" alt="Requirements"></a><br>
<sub><del>Word documents</del></sub><br>
<strong>Executable specs</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/">Specs →</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/"><img src="https://img.shields.io/badge/📄-Documentation-14b8a6?style=for-the-badge" alt="Documentation"></a><br>
<sub><del>Wiki pages that drift</del></sub><br>
<strong>Generated from code</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">Templates →</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/"><img src="https://img.shields.io/badge/✅-Compliance-a855f7?style=for-the-badge" alt="Compliance"></a><br>
<sub><del>Manual checklists</del></sub><br>
<strong>Automated gates</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">Validate →</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/"><img src="https://img.shields.io/badge/🔗-Traceability-ec4899?style=for-the-badge" alt="Traceability"></a><br>
<sub><del>Audit spreadsheets</del></sub><br>
<strong>Git history</strong><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">Pipeline →</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/"><img src="https://img.shields.io/badge/🛡️-Audit-f97316?style=for-the-badge" alt="Audit"></a><br>
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

<table align="center">
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/tutorials/"><img src="https://img.shields.io/badge/1-Install_R2R-3b82f6?style=for-the-badge" alt="Step 1"></a><br>
<sub>CLI for your platform</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/tutorials/">Linux/macOS</a> · <a href="https://ready-to-release.github.io/eac/tutorials/">Windows</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/tutorials/"><img src="https://img.shields.io/badge/2-Setup-10b981?style=for-the-badge" alt="Step 2"></a><br>
<a href="https://ready-to-release.github.io/eac/tutorials/"><code>r2r init</code></a><br>
<a href="https://ready-to-release.github.io/eac/tutorials/"><code>r2r install eac</code></a>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/init-command/"><img src="https://img.shields.io/badge/3-Initialize-8b5cf6?style=for-the-badge" alt="Step 3"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/init-command/"><code>r2r eac init</code></a><br>
<sub>Creates module contracts</sub>
</td>
</tr>
</table>

---

<h2 align="center">Command Areas</h2>

<p align="center">
12 command areas. <code>r2r eac &lt;command&gt;</code>
</p>

<table align="center">
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-overview/"><img src="https://img.shields.io/badge/🏗️-Design-8b5cf6?style=for-the-badge" alt="Design"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-commands/#create-design"><code>create design</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-commands/#serve-design"><code>serve design</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-commands/#validate-design"><code>validate design</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/design-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/"><img src="https://img.shields.io/badge/📋-Specs-22c55e?style=for-the-badge" alt="Specs"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-commands/#create-spec"><code>create spec</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-commands/#validate-specs"><code>validate specs</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-commands/#get-specs-unused-steps"><code>get specs unused-steps</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/specifications-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/"><img src="https://img.shields.io/badge/📄-Templates-14b8a6?style=for-the-badge" alt="Templates"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-commands/#templates-apply"><code>templates apply</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-commands/#templates-list"><code>templates list</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-commands/#templates-install"><code>templates install</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/templates-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-overview/"><img src="https://img.shields.io/badge/📚-Books-6366f1?style=for-the-badge" alt="Books"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-commands/#build-with-book"><code>build &lt;module&gt;</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-commands/#show-books"><code>show books</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-commands/#validate-books"><code>validate books</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/books-commands/">all</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-overview/"><img src="https://img.shields.io/badge/🧪-Test-10b981?style=for-the-badge" alt="Test"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-commands/#test"><code>test</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-commands/#test-suite"><code>test suite</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-commands/#test-debug"><code>test debug</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/test-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/"><img src="https://img.shields.io/badge/✅-Validate-a855f7?style=for-the-badge" alt="Validate"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-commands/#validate-contracts"><code>validate contracts</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-commands/#validate-dependencies"><code>validate dependencies</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-commands/#validate-go-tidy"><code>validate go-tidy</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/validate-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/"><img src="https://img.shields.io/badge/🔒-Security-eab308?style=for-the-badge" alt="Security"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-commands/#security"><code>security</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-commands/#security-sast"><code>security sast</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-commands/#security-sbom"><code>security sbom</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/security-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/"><img src="https://img.shields.io/badge/🛡️-Risk-f97316?style=for-the-badge" alt="Risk"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-commands/#create-risk"><code>create risk</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-commands/#validate-risk"><code>validate risk</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-commands/#show-risk-report"><code>show risk-report</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/risks-commands/">all</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-overview/"><img src="https://img.shields.io/badge/🔀-Workspace-0ea5e9?style=for-the-badge" alt="Workspace"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-commands/#work-create"><code>work create</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-commands/#work-commit"><code>work commit</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-commands/#work-merge"><code>work merge</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/workspace-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-overview/"><img src="https://img.shields.io/badge/🔨-Build-3b82f6?style=for-the-badge" alt="Build"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-commands/#build"><code>build</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-commands/#build"><code>build &lt;module&gt;</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/show-get-list-commands/#show-modules"><code>show modules</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/build-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/"><img src="https://img.shields.io/badge/⚡-Pipeline-ec4899?style=for-the-badge" alt="Pipeline"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-commands/#pipeline-run"><code>pipeline run</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-commands/#pipeline-wait"><code>pipeline wait</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-commands/#pipeline-status"><code>pipeline status</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/pipeline-commands/">all</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/"><img src="https://img.shields.io/badge/🚀-Release-f43f5e?style=for-the-badge" alt="Release"></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-commands/#release-changelog"><code>release changelog</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-commands/#release-this"><code>release this</code></a><br>
<a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-commands/#release-generate-module-calver"><code>release generate-module-calver</code></a><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-overview/">overview</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-configuration/">config</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/commands/areas/release-commands/">all</a></sup>
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
<a href="https://dora.dev/research/"><img src="https://img.shields.io/badge/208×-Faster_Deploys-22c55e?style=for-the-badge" alt="208x"></a><br>
<sub>Deployment frequency</sub>
</td>
<td align="center">
<a href="https://dora.dev/research/"><img src="https://img.shields.io/badge/106×-Faster_Lead_Time-3b82f6?style=for-the-badge" alt="106x"></a><br>
<sub>Commit to production</sub>
</td>
<td align="center">
<a href="https://dora.dev/research/"><img src="https://img.shields.io/badge/7×-Lower_Failure_Rate-8b5cf6?style=for-the-badge" alt="7x"></a><br>
<sub>Change failure rate</sub>
</td>
</tr>
</table>

<p align="center">
The difference isn't less compliance—it's <strong>automated compliance</strong>.<br>
<sub>Small validated changes · Continuous evidence capture · Traceability in every commit</sub>
</p>

<p align="center">
<a href="https://github.com/ready-to-release/eac"><img src="https://img.shields.io/badge/🐕-Dogfooding-f97316?style=flat-square" alt="Dogfooding"></a><br>
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
<a href="https://ready-to-release.github.io/eac/how-to-guides/r2r/creating-extensions/"><img src="https://img.shields.io/badge/🔌-R2R_Extensions-3b82f6?style=for-the-badge" alt="Extensions"></a><br><br>
<code>extensions:</code><br>
<code>&nbsp;&nbsp;- name: 'my-ext'</code><br>
<code>&nbsp;&nbsp;&nbsp;&nbsp;image: 'ghcr.io/...'</code><br><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/r2r/creating-extensions/">Create extensions →</a></sup>
</td>
<td align="center" valign="top">
<a href="https://ready-to-release.github.io/eac/how-to-guides/eac/creating-modules/"><img src="https://img.shields.io/badge/📦-EAC_Modules-8b5cf6?style=for-the-badge" alt="Modules"></a><br><br>
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
<a href="https://ready-to-release.github.io/eac/tutorials/"><img src="https://img.shields.io/badge/🎓-Tutorials-22c55e?style=for-the-badge" alt="Tutorials"></a><br>
<sub>Learn step-by-step</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/tutorials/">Start learning →</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/"><img src="https://img.shields.io/badge/📖-How--to-3b82f6?style=for-the-badge" alt="How-to"></a><br>
<sub>Accomplish tasks</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/how-to-guides/">Find guides →</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/"><img src="https://img.shields.io/badge/📚-Reference-8b5cf6?style=for-the-badge" alt="Reference"></a><br>
<sub>Technical details</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/">Look it up →</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/explanation/"><img src="https://img.shields.io/badge/💡-Explanation-f97316?style=for-the-badge" alt="Explanation"></a><br>
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
<a href="https://github.com/casperease"><img src="https://img.shields.io/badge/👤-casperease-3b82f6?style=for-the-badge" alt="casperease"></a><br>
<sub>Core maintainer</sub><br>
<sup><a href="https://www.linkedin.com/in/casper-nielsen-1a9a208/">linkedin</a> · <a href="https://github.com/casperease">github</a></sup>
</td>
<td align="center">
<a href="https://github.com/miohansen"><img src="https://img.shields.io/badge/👤-miohansen-8b5cf6?style=for-the-badge" alt="miohansen"></a><br>
<sub>Core maintainer</sub><br>
<sup><a href="https://www.linkedin.com/in/mikaelottesenhansen/">linkedin</a> · <a href="https://github.com/miohansen">github</a></sup>
</td>
<td align="center">
<a href="https://claude.ai/code"><img src="https://img.shields.io/badge/🤖-claude_code-f97316?style=for-the-badge" alt="claude code"></a><br>
<sub>Assistant developer</sub><br>
<sup><a href="https://claude.ai/code">claude.ai</a> · <a href="https://docs.anthropic.com/en/docs/claude-code">docs</a></sup>
</td>
<td align="center">
<a href="https://www.linkedin.com/in/tomasmalmsten/"><img src="https://img.shields.io/badge/👤-tomasmalmsten-22c55e?style=for-the-badge" alt="tomasmalmsten"></a><br>
<sub>Contributor</sub><br>
<sup><a href="https://www.linkedin.com/in/tomasmalmsten/">linkedin</a></sup>
</td>
</tr>
</table>

<p align="center">
<a href="https://github.com/ready-to-release/eac/issues/new"><img src="https://img.shields.io/badge/💬-Questions%3F_Open_an_issue-64748b?style=flat-square" alt="Questions"></a><br>
<sub><a href="LICENSE">MIT</a> (code) · <a href="docs/LICENSE">CC BY-SA 4.0</a> (docs)</sub>
</p>
