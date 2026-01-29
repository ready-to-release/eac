// Package tool provides tool ID constants loaded from tool-config.yml.
// These constants provide type-safe, compile-time checked tool references.
// The canonical source of truth is: contracts/eac-core/0.1.0/defaults/tool-config.yml

package tool

// Scanner tool IDs - security scanning tools.
// Categories: sbom, vuln, secrets, iac, compliance, sast, zap
const (
	// ToolTrivySBOM is the Trivy SBOM generator.
	ToolTrivySBOM = "trivy-sbom"

	// ToolTrivyVuln is the Trivy vulnerability scanner.
	ToolTrivyVuln = "trivy-vuln"

	// ToolTrivySecrets is the Trivy secrets scanner.
	ToolTrivySecrets = "trivy-secrets"

	// ToolTrivyCompliance is the Trivy compliance scanner.
	ToolTrivyCompliance = "trivy-compliance"

	// ToolTrivyIaC is the Trivy IaC scanner.
	ToolTrivyIaC = "trivy-iac"

	// ToolSemgrep is the Semgrep SAST scanner.
	ToolSemgrep = "semgrep"

	// ToolZap is the OWASP ZAP DAST scanner.
	ToolZap = "zap"
)

// Server tool IDs - content servers.
const (
	// ToolStaticSite is the static site server (nginx).
	ToolStaticSite = "static-site"

	// ToolMkDocsLive is the MkDocs live server.
	ToolMkDocsLive = "mkdocs-live"

	// ToolStructurizrLite is the Structurizr diagram server.
	ToolStructurizrLite = "structurizr-lite"
)

// Linter tool IDs - code quality tools.
const (
	// ToolGolangciLint is the Go linter aggregator.
	ToolGolangciLint = "golangci-lint"

	// ToolMarkdownlint is the Markdown linter.
	ToolMarkdownlint = "markdownlint"

	// ToolEslint is the JavaScript/TypeScript linter.
	ToolEslint = "eslint"

	// ToolRuff is the Python linter.
	ToolRuff = "ruff"

	// ToolClippy is the Rust linter.
	ToolClippy = "clippy"
)

// Builder tool IDs - build tools.
const (
	// ToolGo is the Go compiler and toolchain.
	ToolGo = "go"

	// ToolNpmBuild is NPM build (runs 'npm run build').
	ToolNpmBuild = "npm-build"

	// ToolMkDocsBuild is MkDocs documentation builder.
	ToolMkDocsBuild = "mkdocs-build"

	// ToolBuildx is Docker buildx for multi-platform builds.
	ToolBuildx = "buildx"

	// ToolCargo is the Rust package manager.
	ToolCargo = "cargo"

	// ToolDotnet is the .NET SDK.
	ToolDotnet = "dotnet"

	// ToolPython is the Python interpreter.
	ToolPython = "python"

	// ToolNpm is the Node.js package manager.
	ToolNpm = "npm"

	// ToolTsc is the TypeScript compiler.
	ToolTsc = "tsc"
)

// Design/Documentation tool IDs.
const (
	// ToolStructurizrCLI is the Structurizr CLI for export/validation.
	ToolStructurizrCLI = "structurizr-cli"

	// ToolPlantuml is the PlantUML renderer.
	ToolPlantuml = "plantuml"

	// ToolDrawio is the DrawIO CLI.
	ToolDrawio = "drawio"

	// ToolMkdocs is the MkDocs documentation generator (system).
	ToolMkdocs = "mkdocs"
)

// Test runner tool IDs.
const (
	// ToolPytest is the Python test runner.
	ToolPytest = "pytest"

	// ToolPester is the PowerShell test runner.
	ToolPester = "pester"
)

// Shell tool IDs.
const (
	// ToolBash is the Bash shell.
	ToolBash = "bash"

	// ToolPwsh is PowerShell.
	ToolPwsh = "pwsh"
)

// Infrastructure tool IDs.
const (
	// ToolDocker is the container runtime.
	ToolDocker = "docker"

	// ToolGit is version control.
	ToolGit = "git"

	// ToolGh is the GitHub CLI.
	ToolGh = "gh"

	// ToolCurl is the HTTP client.
	ToolCurl = "curl"
)

// AllScannerToolIDs returns all known scanner tool IDs.
func AllScannerToolIDs() []string {
	return []string{
		ToolTrivySBOM,
		ToolTrivyVuln,
		ToolTrivySecrets,
		ToolTrivyCompliance,
		ToolTrivyIaC,
		ToolSemgrep,
		ToolZap,
	}
}

// AllServerToolIDs returns all known server tool IDs.
func AllServerToolIDs() []string {
	return []string{
		ToolStaticSite,
		ToolMkDocsLive,
		ToolStructurizrLite,
	}
}

// AllLinterToolIDs returns all known linter tool IDs.
func AllLinterToolIDs() []string {
	return []string{
		ToolGolangciLint,
		ToolMarkdownlint,
		ToolEslint,
		ToolRuff,
		ToolClippy,
	}
}

// AllBuilderToolIDs returns all known builder tool IDs.
func AllBuilderToolIDs() []string {
	return []string{
		ToolGo,
		ToolNpmBuild,
		ToolMkDocsBuild,
		ToolBuildx,
		ToolCargo,
		ToolDotnet,
		ToolPython,
		ToolNpm,
		ToolTsc,
	}
}
