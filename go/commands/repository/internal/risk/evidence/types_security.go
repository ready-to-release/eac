package evidence

import (
	"encoding/json"
	"time"
)

// SecurityResults contains paths to security evidence files for a module.
type SecurityResults struct {
	ModuleName     string    `json:"module_name"`
	SBOMFile       string    `json:"sbom_file,omitempty"`
	VulnFile       string    `json:"vuln_file,omitempty"`
	SecretsFile    string    `json:"secrets_file,omitempty"`
	SASTFile       string    `json:"sast_file,omitempty"`
	ComplianceFile string    `json:"compliance_file,omitempty"`
	IaCFile        string    `json:"iac_file,omitempty"`
	ZAPFile        string    `json:"zap_file,omitempty"`
	LastModified   time.Time `json:"last_modified"`
}

// SecurityEvidenceFile represents the standardized security evidence format.
// This matches the format used by the security command.
type SecurityEvidenceFile struct {
	Module    string          `json:"module"`
	Scanner   string          `json:"scanner"`
	Timestamp string          `json:"timestamp"`
	SHA256    string          `json:"sha256"`
	Findings  json.RawMessage `json:"findings"`
}

// VulnerabilitySummary provides aggregate vulnerability information.
type VulnerabilitySummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Total    int `json:"total"`
}

// SBOMSummary provides aggregate SBOM information.
type SBOMSummary struct {
	TotalComponents int `json:"total_components"`
}
