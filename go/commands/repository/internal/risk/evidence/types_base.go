package evidence

import (
	"time"

	"github.com/ready-to-release/eac/go/core/config"
)

// EvidenceType represents the type of evidence.
type EvidenceType string

const (
	EvidenceTypeUnitTest          EvidenceType = "unit-test"
	EvidenceTypeAcceptanceTest    EvidenceType = "acceptance-test"
	EvidenceTypeVulnerabilityScan EvidenceType = "vulnerability-scan"
	EvidenceTypeSBOM              EvidenceType = "sbom"
	EvidenceTypeSecretsScan       EvidenceType = "secrets-scan"
	EvidenceTypeSASTScan          EvidenceType = "sast-scan"
	EvidenceTypeComplianceScan    EvidenceType = "compliance-scan"
	EvidenceTypeIaCScan           EvidenceType = "iac-scan"
	EvidenceTypeDAST              EvidenceType = "dast-scan"
)

// Evidence represents a piece of evidence for risk assessment.
type Evidence struct {
	Type       EvidenceType `json:"type"`
	Path       string       `json:"path"`
	ModTime    time.Time    `json:"mod_time"`
	Module     string       `json:"module,omitempty"`
	ControlIDs []string     `json:"control_ids,omitempty"` // Controls this evidence relates to
}

// EvidenceAgePolicy controls evidence age validation.
type EvidenceAgePolicy struct {
	MaxAge time.Duration `json:"max_age"`
}

// DefaultEvidenceAgePolicy returns the default evidence age policy.
func DefaultEvidenceAgePolicy() EvidenceAgePolicy {
	return EvidenceAgePolicy{
		MaxAge: config.EvidenceMaxAge(),
	}
}
