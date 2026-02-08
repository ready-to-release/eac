package pipeline

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		PipelineAwaitCI,
		PipelineAwaitRelease,
		PipelineCheckRecentRun,
		PipelineDownloadEvidenceArtifacts,
		PipelineFindRunID,
		PipelineGetArtifactID,
		PipelineGetTreeFiles,
		PipelineRun,
		PipelineStatus,
		PipelineWait,
	)
}
