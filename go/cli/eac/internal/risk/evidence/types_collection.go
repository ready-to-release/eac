package evidence

import "time"

// EvidenceCollection holds all collected evidence for a module.
type EvidenceCollection struct {
	Module           string                `json:"module"`
	TestManifestData *TestManifestData     `json:"test_manifest_data,omitempty"` // Simplified test manifest data
	SecurityResults  *SecurityResults      `json:"security_results,omitempty"`
	TestSummary      *TestSummary          `json:"test_summary,omitempty"`
	VulnSummary      *VulnerabilitySummary `json:"vuln_summary,omitempty"`
	SBOMSummary      *SBOMSummary          `json:"sbom_summary,omitempty"`
	CollectedAt      time.Time             `json:"collected_at"`
	Warnings         []string              `json:"warnings,omitempty"` // Warnings about missing or stale evidence
}

// HasAnyEvidence returns true if the collection has any evidence.
func (ec *EvidenceCollection) HasAnyEvidence() bool {
	if ec.TestManifestData != nil && len(ec.TestManifestData.Tests) > 0 {
		return true
	}
	if ec.SecurityResults != nil && (ec.SecurityResults.VulnFile != "" || ec.SecurityResults.SBOMFile != "") {
		return true
	}
	return false
}

// HasTestEvidence returns true if test evidence exists.
func (ec *EvidenceCollection) HasTestEvidence() bool {
	return ec.TestManifestData != nil && len(ec.TestManifestData.Tests) > 0
}

// HasSecurityEvidence returns true if security evidence exists.
func (ec *EvidenceCollection) HasSecurityEvidence() bool {
	return ec.SecurityResults != nil && (ec.SecurityResults.VulnFile != "" || ec.SecurityResults.SBOMFile != "")
}

// LatestModTime returns the most recent modification time of any evidence.
func (ec *EvidenceCollection) LatestModTime() time.Time {
	var latest time.Time

	if ec.TestManifestData != nil && ec.TestManifestData.TestTime.After(latest) {
		latest = ec.TestManifestData.TestTime
	}
	if ec.SecurityResults != nil && ec.SecurityResults.LastModified.After(latest) {
		latest = ec.SecurityResults.LastModified
	}

	return latest
}
