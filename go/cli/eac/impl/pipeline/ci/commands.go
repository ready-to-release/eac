package ci

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		PipelineCI,
		PipelineCIDispatchAndWait,
		PipelineCIGetRunID,
		PipelineCISchedule,
		PipelineCISummaryLink,
	)
}
