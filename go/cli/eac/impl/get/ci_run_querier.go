package get

import (
	"fmt"

	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/github"
)

// Ensure ghCIRunQuerier implements cache.CIRunQuerier.
var _ cache.CIRunQuerier = (*ghCIRunQuerier)(nil)

// ghCIRunQuerier adapts github.API to the cache.CIRunQuerier port.
type ghCIRunQuerier struct {
	api github.API
}

// NewGHCIRunQuerier creates a CIRunQuerier backed by a GitHub API client.
func NewGHCIRunQuerier(api github.API) cache.CIRunQuerier {
	return &ghCIRunQuerier{api: api}
}

// LastSuccessfulRunSHA queries GitHub for the last successful workflow run.
func (q *ghCIRunQuerier) LastSuccessfulRunSHA(workflowName string) (string, error) {
	runs, err := q.api.ListRuns(workflowName, github.ListRunsOpts{
		Status: "success",
		Limit:  1,
	})
	if err != nil {
		return "", fmt.Errorf("ListRuns failed: %w", err)
	}
	if len(runs) == 0 {
		return "", nil
	}
	return runs[0].HeadSHA, nil
}
