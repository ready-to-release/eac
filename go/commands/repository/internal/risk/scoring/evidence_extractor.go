// Package scoring provides evidence extraction for AI risk analysis.
package scoring

import (
	"encoding/json"
	"strings"

	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/evidence"
	coreConfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// ExtractVulnerabilityFindings extracts vulnerability data from security scan evidence
// for AI risk analysis. Parses multiple scanner formats (Trivy, GoSec, Nancy, ZAP).
func ExtractVulnerabilityFindings(evidenceCollection *evidence.EvidenceCollection) []VulnerabilityInput {
	if evidenceCollection == nil || evidenceCollection.SecurityResults == nil {
		return []VulnerabilityInput{}
	}

	var findings []VulnerabilityInput

	secResults := evidenceCollection.SecurityResults

	// Extract from vulnerability scanner (Trivy)
	if secResults.VulnFile != "" {
		vulnFindings := extractFromVulnScan(secResults.VulnFile)
		findings = append(findings, vulnFindings...)
	}

	// Extract from SAST scanner (GoSec)
	if secResults.SASTFile != "" {
		sastFindings := extractFromSASTScan(secResults.SASTFile)
		findings = append(findings, sastFindings...)
	}

	// Extract from DAST scanner (ZAP)
	if secResults.ZAPFile != "" {
		zapFindings := extractFromZAPScan(secResults.ZAPFile)
		findings = append(findings, zapFindings...)
	}

	return findings
}

// extractFromVulnScan extracts vulnerability findings from Trivy scan results.
func extractFromVulnScan(vulnFile string) []VulnerabilityInput {
	var findings []VulnerabilityInput

	// Load security evidence file
	evidenceFile, err := evidence.LoadSecurityEvidence(vulnFile)
	if err != nil {
		return findings
	}

	// Parse Trivy findings format
	var trivyData map[string]interface{}
	if err := json.Unmarshal(evidenceFile.Findings, &trivyData); err != nil {
		return findings
	}

	// Handle Trivy Results format
	if results, ok := trivyData["Results"].([]interface{}); ok {
		for _, result := range results {
			if r, ok := result.(map[string]interface{}); ok {
				if vulns, ok := r["Vulnerabilities"].([]interface{}); ok {
					for _, vuln := range vulns {
						if v, ok := vuln.(map[string]interface{}); ok {
							finding := VulnerabilityInput{
								Scanner:       evidenceFile.Scanner,
								Severity:      getString(v, "Severity"),
								Vulnerability: getString(v, "VulnerabilityID"),
								Description:   getString(v, "Description"),
								Package:       getString(v, "PkgName"),
							}

							// Extract CVSS score if available
							if cvss, ok := v["CVSS"].(map[string]interface{}); ok {
								if score, ok := cvss["Score"].(float64); ok {
									finding.CVSSScore = score
								}
							}

							findings = append(findings, finding)
						}
					}
				}
			}
		}
	}

	return findings
}

// extractFromSASTScan extracts findings from GoSec SAST scan results.
func extractFromSASTScan(sastFile string) []VulnerabilityInput {
	var findings []VulnerabilityInput

	evidenceFile, err := evidence.LoadSecurityEvidence(sastFile)
	if err != nil {
		return findings
	}

	// Parse GoSec findings format
	var gosecData map[string]interface{}
	if err := json.Unmarshal(evidenceFile.Findings, &gosecData); err != nil {
		return findings
	}

	// GoSec format: {"Issues": [{...}]}
	if issues, ok := gosecData["Issues"].([]interface{}); ok {
		for _, issue := range issues {
			if i, ok := issue.(map[string]interface{}); ok {
				// Map GoSec severity to standard severity
				severity := mapGoSecSeverity(getString(i, "severity"))

				finding := VulnerabilityInput{
					Scanner:       evidenceFile.Scanner,
					Severity:      severity,
					Vulnerability: getString(i, "rule_id"),
					Description:   getString(i, "details"),
				}

				findings = append(findings, finding)
			}
		}
	}

	return findings
}

// extractFromZAPScan extracts findings from OWASP ZAP DAST scan results.
func extractFromZAPScan(zapFile string) []VulnerabilityInput {
	var findings []VulnerabilityInput

	evidenceFile, err := evidence.LoadSecurityEvidence(zapFile)
	if err != nil {
		return findings
	}

	// Parse ZAP findings format
	var zapData map[string]interface{}
	if err := json.Unmarshal(evidenceFile.Findings, &zapData); err != nil {
		return findings
	}

	// ZAP format: site -> alerts
	if site, ok := zapData["site"].([]interface{}); ok {
		for _, s := range site {
			if siteData, ok := s.(map[string]interface{}); ok {
				if alerts, ok := siteData["alerts"].([]interface{}); ok {
					for _, alert := range alerts {
						if a, ok := alert.(map[string]interface{}); ok {
							// Map ZAP risk to severity
							severity := mapZAPRisk(getString(a, "risk"))

							finding := VulnerabilityInput{
								Scanner:       evidenceFile.Scanner,
								Severity:      severity,
								Vulnerability: getString(a, "name"),
								Description:   getString(a, "desc"),
							}

							findings = append(findings, finding)
						}
					}
				}
			}
		}
	}

	return findings
}

// BuildModuleContext builds context about the module for AI analysis.
func BuildModuleContext(moduleName string, registry *modules.Registry, satisfiedControls []string) ModuleContext {
	// Determine criticality by module moniker (from risk-config.yml, falls back to _default)
	criticality := coreConfig.DefaultRiskScoringConfig().GetCriticality(moduleName)

	if registry == nil {
		// Return default context if registry not available
		return ModuleContext{
			ComponentType:    "service",
			Criticality:      criticality,
			ExistingControls: satisfiedControls,
		}
	}

	// Get module from registry
	module, exists := registry.Get(moduleName)
	if !exists || module == nil {
		return ModuleContext{
			ComponentType:    "service",
			Criticality:      criticality,
			ExistingControls: satisfiedControls,
		}
	}

	// Extract primary component type for context (informational only)
	componentType := "service" // default
	if module.HasComponent("go") {
		componentType = "go"
	} else if module.HasComponent("typescript") {
		componentType = "typescript"
	} else if module.HasComponent("container") {
		componentType = "container"
	}

	return ModuleContext{
		ComponentType:    componentType,
		Criticality:      criticality,
		ExistingControls: satisfiedControls,
	}
}

// Helper functions

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func mapGoSecSeverity(gosecSeverity string) string {
	// GoSec uses: "HIGH", "MEDIUM", "LOW"
	upper := strings.ToUpper(gosecSeverity)
	switch upper {
	case "HIGH":
		return "HIGH"
	case "MEDIUM":
		return "MEDIUM"
	case "LOW":
		return "LOW"
	default:
		return "MEDIUM"
	}
}

func mapZAPRisk(zapRisk string) string {
	// ZAP uses: "High", "Medium", "Low", "Informational"
	upper := strings.ToUpper(zapRisk)
	switch upper {
	case "HIGH":
		return "HIGH"
	case "MEDIUM":
		return "MEDIUM"
	case "LOW":
		return "LOW"
	case "INFORMATIONAL":
		return "LOW"
	default:
		return "MEDIUM"
	}
}
