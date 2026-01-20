package formats

import (
	"io"
)

// ExportFormatter defines the interface for different export formats.
type ExportFormatter interface {
	Write(export *ManualTestExport, w io.Writer) error
	FileExtension() string
}

// ManualTestExport is the root export structure.
type ManualTestExport struct {
	ExportMetadata ExportMetadata     `json:"export_metadata" yaml:"export_metadata"`
	Scenarios      []ExportedScenario `json:"scenarios" yaml:"scenarios"`
}

// ExportMetadata contains export context.
type ExportMetadata struct {
	ExportTime     string `json:"export_time" yaml:"export_time"`
	Module         string `json:"module" yaml:"module"`
	ReleaseVersion string `json:"release_version" yaml:"release_version"`
	GitCommit      string `json:"git_commit" yaml:"git_commit"`
	SchemaVersion  string `json:"schema_version" yaml:"schema_version"`
}

// ExportedScenario is a single manual test scenario.
type ExportedScenario struct {
	ScenarioID   string   `json:"scenario_id" yaml:"scenario_id"`
	FeatureName  string   `json:"feature_name" yaml:"feature_name"`
	ScenarioName string   `json:"scenario_name" yaml:"scenario_name"`
	Tags         []string `json:"tags" yaml:"tags"`
	Steps        []string `json:"steps" yaml:"steps"`
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	FilePath     string   `json:"file_path,omitempty" yaml:"file_path,omitempty"`
}
