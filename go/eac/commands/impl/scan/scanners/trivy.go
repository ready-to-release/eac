// Package scanners provides Trivy security scanner adapters.
package scanners

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// trivySBOMAdapter adapts the standard ScanFunc signature to the Trivy SBOM implementation.
func trivySBOMAdapter(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts tool.ScanOptions) (interface{}, error) {
	ctx := GlobalScanContext
	if ctx == nil {
		return nil, fmt.Errorf("scan context not initialized: trivy sbom requires GlobalScanContext to be set")
	}

	if ctx.TrivyImage == "" {
		return nil, fmt.Errorf("trivy image not configured in scan context")
	}

	format := ctx.SBOMFormat
	if format == "" {
		format = "cyclonedx-json"
	}

	// Delegate to existing implementation
	return internal.RunTrivySBOM(workspaceRoot, moduleRoot, format, ctx.TrivyImage)
}

// trivyVulnAdapter adapts the standard ScanFunc signature to the Trivy vulnerability scanner.
func trivyVulnAdapter(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts tool.ScanOptions) (interface{}, error) {
	ctx := GlobalScanContext
	if ctx == nil {
		return nil, fmt.Errorf("scan context not initialized: trivy vuln requires GlobalScanContext to be set")
	}

	if ctx.TrivyImage == "" {
		return nil, fmt.Errorf("trivy image not configured in scan context")
	}

	// Resolve module root to absolute path for RunTrivyVuln which expects absolute path
	absModuleRoot := filepath.Join(workspaceRoot, moduleRoot)

	// Delegate to existing implementation
	return internal.RunTrivyVuln(absModuleRoot, ctx.VulnSeverities, ctx.TrivyImage)
}

// trivySecretsAdapter adapts the standard ScanFunc signature to the Trivy secrets scanner.
func trivySecretsAdapter(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts tool.ScanOptions) (interface{}, error) {
	ctx := GlobalScanContext
	if ctx == nil {
		return nil, fmt.Errorf("scan context not initialized: trivy secrets requires GlobalScanContext to be set")
	}

	if ctx.TrivyImage == "" {
		return nil, fmt.Errorf("trivy image not configured in scan context")
	}

	// Resolve module root to absolute path
	absModuleRoot := filepath.Join(workspaceRoot, moduleRoot)

	// Delegate to existing implementation
	return internal.RunTrivySecrets(absModuleRoot, ctx.TrivyImage)
}

// trivyIaCAdapter adapts the standard ScanFunc signature to the Trivy IaC scanner.
func trivyIaCAdapter(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts tool.ScanOptions) (interface{}, error) {
	ctx := GlobalScanContext
	if ctx == nil {
		return nil, fmt.Errorf("scan context not initialized: trivy iac requires GlobalScanContext to be set")
	}

	if ctx.TrivyImage == "" {
		return nil, fmt.Errorf("trivy image not configured in scan context")
	}

	// Resolve module root to absolute path
	absModuleRoot := filepath.Join(workspaceRoot, moduleRoot)

	// Delegate to existing implementation
	return internal.RunTrivyIaC(absModuleRoot, ctx.TrivyImage)
}

// trivyComplianceAdapter adapts the standard ScanFunc signature to the Trivy compliance scanner.
func trivyComplianceAdapter(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts tool.ScanOptions) (interface{}, error) {
	ctx := GlobalScanContext
	if ctx == nil {
		return nil, fmt.Errorf("scan context not initialized: trivy compliance requires GlobalScanContext to be set")
	}

	if ctx.TrivyImage == "" {
		return nil, fmt.Errorf("trivy image not configured in scan context")
	}

	compliance := ctx.ComplianceStandard
	if compliance == "" {
		compliance = "docker-cis"
	}

	// Resolve module root to absolute path
	absModuleRoot := filepath.Join(workspaceRoot, moduleRoot)

	// Delegate to existing implementation
	return internal.RunTrivyCompliance(absModuleRoot, compliance, ctx.TrivyImage)
}
