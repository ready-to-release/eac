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
  <a href="https://github.com/ready-to-release/eac/actions/workflows/change-trigger.yaml"><img src="https://img.shields.io/github/actions/workflow/status/ready-to-release/eac/change-trigger.yaml?style=flat-square&label=ci" alt="CI"></a>
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

<p align="center">
  <a href="https://github.com/ready-to-release/eac/releases?q=tutorials+pdf"><img src="https://img.shields.io/badge/tutorials-pdf-f97316?style=flat-square&logo=adobeacrobatreader&logoColor=white" alt="tutorials"></a>
  <a href="https://github.com/ready-to-release/eac/releases?q=howto+pdf"><img src="https://img.shields.io/badge/how--to-pdf-f97316?style=flat-square&logo=adobeacrobatreader&logoColor=white" alt="howto"></a>
  <a href="https://github.com/ready-to-release/eac/releases?q=explanation+pdf"><img src="https://img.shields.io/badge/explanation-pdf-f97316?style=flat-square&logo=adobeacrobatreader&logoColor=white" alt="explanation"></a>
  <a href="https://github.com/ready-to-release/eac/releases?q=specifications+pdf"><img src="https://img.shields.io/badge/specifications-pdf-f97316?style=flat-square&logo=adobeacrobatreader&logoColor=white" alt="specifications"></a>
  <a href="https://github.com/ready-to-release/eac/releases?q=repository-report+pdf"><img src="https://img.shields.io/badge/repo--report-pdf-f97316?style=flat-square&logo=adobeacrobatreader&logoColor=white" alt="repository-report"></a>
</p>

---

<h2 align="center">The Problem</h2>

<p align="center">
<a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/compliance-velocity-paradox/"><img src="https://img.shields.io/badge/⚠️_THE_PARADOX-Compliance_demands_rigor._Markets_demand_speed.-dc2626?style=for-the-badge&labelColor=7f1d1d" alt="The Paradox"></a>
</p>

<div align="center">
<table>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/validate/"><img src="https://img.shields.io/badge/🎯-No_Source_of_Truth-dc2626?style=for-the-badge" alt="No Source of Truth"></a><br>
<sub>Version conflicts everywhere</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/validate/">validate</a> · <a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/paradigm/">paradigm</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/show/books/"><img src="https://img.shields.io/badge/📄-Docs_Drift-dc2626?style=for-the-badge" alt="Docs Drift"></a><br>
<sub>Wiki pages go stale</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/show/books/">books</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/templates/">templates</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/"><img src="https://img.shields.io/badge/🔍-Traceability_Lost-dc2626?style=for-the-badge" alt="Traceability Lost"></a><br>
<sub>Reconstructed after the fact</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/">pipeline</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/release/">release</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/scan/"><img src="https://img.shields.io/badge/⏰-Audit_Prep-dc2626?style=for-the-badge" alt="Audit Prep"></a><br>
<sub>Takes months, not minutes</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/scan/">scan</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/create/risk-assess/">risk-assess</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/validate/"><img src="https://img.shields.io/badge/📦-Evidence_Scattered-dc2626?style=for-the-badge" alt="Evidence Scattered"></a><br>
<sub>Proof lives in 10 different tools</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/validate/">validate</a> · <a href="https://ready-to-release.github.io/eac/how-to-guides/">guides</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/scan/"><img src="https://img.shields.io/badge/🛡️-Scan_Late-dc2626?style=for-the-badge" alt="Scan Late"></a><br>
<sub>Scans run too late to matter</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/scan/">scan</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/scan/">commands</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/create/spec/"><img src="https://img.shields.io/badge/🔗-Specs_Disconnect-dc2626?style=for-the-badge" alt="Specs Disconnect"></a><br>
<sub>Requirements live apart from tests</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/create/spec/">specs</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/test/">test</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/create/design/"><img src="https://img.shields.io/badge/🏛️-Knowledge_Silos-dc2626?style=for-the-badge" alt="Knowledge Silos"></a><br>
<sub>Tribal wisdom, not shared systems</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/create/design/">design</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/templates/">templates</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/release/"><img src="https://img.shields.io/badge/🚀-Release_Events-dc2626?style=for-the-badge" alt="Release Events"></a><br>
<sub>Big bang deployments, fingers crossed</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/release/">release</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/">pipeline</a></sup>
</td>
</tr>
<tr>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/create/"><img src="https://img.shields.io/badge/📋-Copy--Paste_Compliance-dc2626?style=for-the-badge" alt="Copy-Paste Compliance"></a><br>
<sub>Same docs, different date</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/create/risk-profile/">risk-profile</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/templates/">templates</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/"><img src="https://img.shields.io/badge/🚧-Manual_Gates-dc2626?style=for-the-badge" alt="Manual Gates"></a><br>
<sub>Approvals slow without adding safety</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/">pipeline</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/validate/">validate</a></sup>
</td>
<td align="center">
<a href="https://ready-to-release.github.io/eac/reference/commands/release/"><img src="https://img.shields.io/badge/🎲-Deployment_Roulette-dc2626?style=for-the-badge" alt="Deployment Roulette"></a><br>
<sub>Nobody knows what's in production</sub><br>
<sup><a href="https://ready-to-release.github.io/eac/reference/commands/release/">release</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/scan/">scan</a></sup>
</td>
</tr>
</table>
</div>

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

<div align="center">
<table>
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
</div>

<p align="center">
<strong>One source of truth. One workflow. One audit trail.</strong>
</p>

---

<h2 align="center">How It Works</h2>

<p align="center">
One commit. Validated build. Auditable release.
</p>

<div align="center">
<table>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/create/spec/"><img src="https://img.shields.io/badge/Requirements-22c55e?style=for-the-badge" alt="Requirements"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/templates/"><img src="https://img.shields.io/badge/Documentation-14b8a6?style=for-the-badge" alt="Documentation"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/validate/"><img src="https://img.shields.io/badge/Compliance-a855f7?style=for-the-badge" alt="Compliance"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/"><img src="https://img.shields.io/badge/Traceability-ec4899?style=for-the-badge" alt="Traceability"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/scan/"><img src="https://img.shields.io/badge/Audit-f97316?style=for-the-badge" alt="Audit"></a></td>
</tr>
<tr>
<td align="center"><del>Word documents</del></td>
<td align="center"><del>Wiki pages that drift</del></td>
<td align="center"><del>Manual checklists</del></td>
<td align="center"><del>Audit spreadsheets</del></td>
<td align="center"><del>Months of prep</del></td>
</tr>
<tr>
<td align="center"><strong>Executable specs</strong></td>
<td align="center"><strong>Generated from code</strong></td>
<td align="center"><strong>Automated gates</strong></td>
<td align="center"><strong>Git history</strong></td>
<td align="center"><strong>Always ready</strong></td>
</tr>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/create/spec/">Specs →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/templates/">Templates →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/validate/">Validate →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/">Pipeline →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/scan/">scan</a> · <a href="https://ready-to-release.github.io/eac/reference/commands/update/evidence/">evidence</a></td>
</tr>
</table>
</div>

<p align="center">
<strong>The result</strong>: Audit-ready documentation as a byproduct of your pipeline, not a bottleneck before release.<br>
<a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/">Understand the paradigm →</a> · <a href="https://ready-to-release.github.io/eac/explanation/everything-as-code/compliance-velocity-paradox/">See the compliance-velocity paradox →</a>
</p>

---

<h2 align="center">Get Started</h2>

<p align="center">
Up and running in under 5 minutes.
</p>

<div align="center">
<details>
<summary><b>View install commands</b></summary>
<br>

> **Linux / macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

> **Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

</details>
</div>

<div align="center">
<table>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/tutorials/"><img src="https://img.shields.io/badge/1-Install_R2R-3b82f6?style=for-the-badge" alt="1 Install R2R"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/tutorials/getting-started/quick-start"><img src="https://img.shields.io/badge/2-Setup-10b981?style=for-the-badge" alt="2 Setup"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/other/init"><img src="https://img.shields.io/badge/3-Initialize-8b5cf6?style=for-the-badge" alt="3 Initialize"></a></td>
</tr>
<tr>
<td align="center">CLI for your platform</td>
<td align="center"><code>r2r init</code></td>
<td align="center"><code>r2r eac init</code></td>
</tr>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/tutorials/getting-started/quick-start">Linux/macOS</a> · <a href="https://ready-to-release.github.io/eac/tutorials/getting-started/quick-start">Windows</a></td>
<td align="center"><code>r2r install eac</code></td>
<td align="center">Creates module contracts</td>
</tr>
</table>
</div>

---

<h2 align="center">Command Areas</h2>

<p align="center">
12 command areas. <code>r2r eac &lt;command&gt;</code>
</p>

<div align="center">
<table>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/create/design/"><img src="https://img.shields.io/badge/Design-8b5cf6?style=for-the-badge" alt="Design"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/create/spec/"><img src="https://img.shields.io/badge/Specs-22c55e?style=for-the-badge" alt="Specs"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/templates/"><img src="https://img.shields.io/badge/Templates-14b8a6?style=for-the-badge" alt="Templates"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/show/books/"><img src="https://img.shields.io/badge/Books-6366f1?style=for-the-badge" alt="Books"></a></td>
</tr>
<tr>
<td align="center"><code>create design</code></td>
<td align="center"><code>create spec</code></td>
<td align="center"><code>templates install</code></td>
<td align="center"><code>build &lt;module&gt;</code></td>
</tr>
<tr>
<td align="center"><code>validate design</code></td>
<td align="center"><code>validate specs</code></td>
<td align="center"><code>templates install ai</code></td>
<td align="center"><code>show books</code></td>
</tr>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/test/"><img src="https://img.shields.io/badge/Test-10b981?style=for-the-badge" alt="Test"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/validate/"><img src="https://img.shields.io/badge/Validate-a855f7?style=for-the-badge" alt="Validate"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/scan/"><img src="https://img.shields.io/badge/Scan-eab308?style=for-the-badge" alt="Scan"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/create/"><img src="https://img.shields.io/badge/Risk-f97316?style=for-the-badge" alt="Risk"></a></td>
</tr>
<tr>
<td align="center"><code>test</code></td>
<td align="center"><code>validate contracts</code></td>
<td align="center"><code>scan</code></td>
<td align="center"><code>create risk-profile</code></td>
</tr>
<tr>
<td align="center"><code>test suite</code></td>
<td align="center"><code>validate deps</code></td>
<td align="center"><code>scan vuln</code></td>
<td align="center"><code>create risk-assess</code></td>
</tr>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/work/"><img src="https://img.shields.io/badge/Workspace-0ea5e9?style=for-the-badge" alt="Workspace"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/other/build/"><img src="https://img.shields.io/badge/Build-3b82f6?style=for-the-badge" alt="Build"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/pipeline/"><img src="https://img.shields.io/badge/Pipeline-ec4899?style=for-the-badge" alt="Pipeline"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/commands/release/"><img src="https://img.shields.io/badge/Release-f43f5e?style=for-the-badge" alt="Release"></a></td>
</tr>
<tr>
<td align="center"><code>work create</code></td>
<td align="center"><code>build</code></td>
<td align="center"><code>pipeline run</code></td>
<td align="center"><code>release changelog</code></td>
</tr>
<tr>
<td align="center"><code>work merge</code></td>
<td align="center"><code>build &lt;module&gt;</code></td>
<td align="center"><code>pipeline wait</code></td>
<td align="center"><code>release this</code></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://ready-to-release.github.io/eac/how-to-guides/"><strong>Browse all commands →</strong></a>
</p>

---

<h2 align="center">Why It Works</h2>

<p align="center">
<a href="https://dora.dev/research/">DORA research</a> proves regulated industries achieve elite performance:
</p>

<div align="center">
<table>
<tr>
<td align="center"><a href="https://dora.dev/research/"><img src="https://img.shields.io/badge/208×-Faster_Deploys-22c55e?style=for-the-badge" alt="208x Faster Deploys"></a></td>
<td align="center"><a href="https://dora.dev/research/"><img src="https://img.shields.io/badge/106×-Faster_Lead_Time-3b82f6?style=for-the-badge" alt="106x Faster Lead Time"></a></td>
<td align="center"><a href="https://dora.dev/research/"><img src="https://img.shields.io/badge/7×-Lower_Failure_Rate-8b5cf6?style=for-the-badge" alt="7x Lower Failure Rate"></a></td>
</tr>
<tr>
<td align="center">Deployment frequency</td>
<td align="center">Commit to production</td>
<td align="center">Change failure rate</td>
</tr>
</table>
</div>

<p align="center">
The difference isn't less compliance—it's <strong>automated compliance</strong>.<br>
<sub>Small validated changes · Continuous evidence capture · Traceability in every commit</sub>
</p>

<p align="center">
<a href="https://github.com/ready-to-release/eac"><img src="https://img.shields.io/badge/🐕-Dogfooding-f97316?style=flat-square" alt="Dogfooding"></a><br>
<sub>This repository uses R2R and EAC to build itself—living proof of the toolings efficacy.</sub>
</p>

---

<h2 align="center">Extend It</h2>

<p align="center">
Build once. Reuse everywhere.
</p>

<div align="center">
<table>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/how-to-guides/r2r/creating-extensions/"><img src="https://img.shields.io/badge/R2R_Extensions-3b82f6?style=for-the-badge" alt="R2R Extensions"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/how-to-guides/eac/modules/creating-modules"><img src="https://img.shields.io/badge/EAC_Modules-8b5cf6?style=for-the-badge" alt="EAC Modules"></a></td>
</tr>
<tr>
<td align="center"><code>extensions:</code></td>
<td align="center"><code>modules:</code></td>
</tr>
<tr>
<td align="center"><code>  - name: 'my-ext'</code></td>
<td align="center"><code>  - moniker: my-svc</code></td>
</tr>
<tr>
<td align="center"><code>    image: 'ghcr.io/...'</code></td>
<td align="center"><code>    type: go</code></td>
</tr>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/how-to-guides/r2r/creating-extensions/">Create extensions →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/how-to-guides/eac/modules/creating-modules">Add modules →</a></td>
</tr>
</table>
</div>

<p align="center">
<strong>Your standards. Your tooling. One CLI.</strong>
</p>

---

<h2 align="center">Documentation</h2>

<p align="center">
Four paths to understanding—pick your learning style.
</p>

<div align="center">
<table>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/tutorials/"><img src="https://img.shields.io/badge/Tutorials-22c55e?style=for-the-badge" alt="Tutorials"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/how-to-guides/"><img src="https://img.shields.io/badge/How--to-3b82f6?style=for-the-badge" alt="How-to"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/"><img src="https://img.shields.io/badge/Reference-8b5cf6?style=for-the-badge" alt="Reference"></a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/explanation/"><img src="https://img.shields.io/badge/Explanation-f97316?style=for-the-badge" alt="Explanation"></a></td>
</tr>
<tr>
<td align="center">Learn step-by-step</td>
<td align="center">Accomplish tasks</td>
<td align="center">Technical details</td>
<td align="center">Understand why</td>
</tr>
<tr>
<td align="center"><a href="https://ready-to-release.github.io/eac/tutorials/">Start learning →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/how-to-guides/">Find guides →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/reference/">Look it up →</a></td>
<td align="center"><a href="https://ready-to-release.github.io/eac/explanation/">Deep dive →</a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://ready-to-release.github.io/eac/"><strong>Full documentation →</strong></a>
</p>

---

<h2 align="center">The Crew</h2>

<p align="center">
Built by practitioners who ship to production daily.
</p>

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/casperease"><img src="https://img.shields.io/badge/casperease-3b82f6?style=for-the-badge" alt="casperease"></a></td>
<td align="center"><a href="https://github.com/miohansen"><img src="https://img.shields.io/badge/miohansen-8b5cf6?style=for-the-badge" alt="miohansen"></a></td>
<td align="center"><a href="https://claude.ai/code"><img src="https://img.shields.io/badge/claude_code-f97316?style=for-the-badge" alt="claude code"></a></td>
<td align="center"><a href="https://www.linkedin.com/in/tomasmalmsten/"><img src="https://img.shields.io/badge/tomasmalmsten-22c55e?style=for-the-badge" alt="tomasmalmsten"></a></td>
</tr>
<tr>
<td align="center">Core maintainer</td>
<td align="center">Core maintainer</td>
<td align="center">Assistant developer</td>
<td align="center">Contributor</td>
</tr>
<tr>
<td align="center"><a href="https://www.linkedin.com/in/casper-nielsen-1a9a208/">linkedin</a> · <a href="https://github.com/casperease">github</a></td>
<td align="center"><a href="https://www.linkedin.com/in/mikaelottesenhansen/">linkedin</a> · <a href="https://github.com/miohansen">github</a></td>
<td align="center"><a href="https://claude.ai/code">claude.ai</a> · <a href="https://docs.anthropic.com/en/docs/claude-code">docs</a></td>
<td align="center"><a href="https://www.linkedin.com/in/tomasmalmsten/">linkedin</a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://github.com/ready-to-release/eac/issues/new"><img src="https://img.shields.io/badge/💬-Questions%3F_Open_an_issue-64748b?style=flat-square" alt="Questions"></a><br>
<sub><a href="LICENSE">MIT</a> (code) · <a href="docs/LICENSE">CC BY-SA 4.0</a> (docs)</sub>
</p>
