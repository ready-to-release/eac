package cache

import "fmt"

// CIRunQuerier queries CI workflow run status from an external system.
// This is a port interface — the cache package never imports github or
// other infrastructure. Callers inject a concrete implementation.
type CIRunQuerier interface {
	// LastSuccessfulRunSHA returns the HEAD SHA of the most recent successful
	// run for the given workflow file name (e.g. "ci-core.yaml").
	// Returns ("", nil) when no successful run exists.
	LastSuccessfulRunSHA(workflowName string) (sha string, err error)
}

// CICacheResult holds the result of a CI cache check for a single module.
type CICacheResult struct {
	Module     string // Module moniker
	Cached     bool   // True if module has valid CI at HEAD
	Reason     string // Human-readable reason for the decision
	LastRunSHA string // SHA of the last successful CI run (may be empty)
}

// CICacheChecker checks whether a module's CI build can be skipped
// because a successful build already exists at the current HEAD SHA.
type CICacheChecker struct {
	querier CIRunQuerier
	config  *Config // nil means caching enabled
}

// NewCICacheChecker creates a new CI cache checker.
// Pass nil for config to enable caching (default behavior).
func NewCICacheChecker(querier CIRunQuerier, config *Config) *CICacheChecker {
	return &CICacheChecker{
		querier: querier,
		config:  config,
	}
}

// Check determines whether the given module has a valid CI build at headSHA.
// The workflow name is derived as "ci-{module}.yaml".
func (c *CICacheChecker) Check(module, headSHA string) CICacheResult {
	// If config says skip CI cache, bypass entirely
	if c.config != nil && c.config.ShouldSkipCI() {
		return CICacheResult{
			Module: module,
			Cached: false,
			Reason: "cache_bypassed",
		}
	}

	workflowName := fmt.Sprintf("ci-%s.yaml", module)
	lastSHA, err := c.querier.LastSuccessfulRunSHA(workflowName)
	if err != nil {
		return CICacheResult{
			Module: module,
			Cached: false,
			Reason: fmt.Sprintf("query_failed: %v", err),
		}
	}

	if lastSHA == "" {
		return CICacheResult{
			Module: module,
			Cached: false,
			Reason: "no_ci_run",
		}
	}

	if lastSHA == headSHA {
		return CICacheResult{
			Module:     module,
			Cached:     true,
			Reason:     "valid_ci_at_head",
			LastRunSHA: lastSHA,
		}
	}

	return CICacheResult{
		Module:     module,
		Cached:     false,
		Reason:     fmt.Sprintf("ci_at_different_sha:%s", lastSHA[:min(7, len(lastSHA))]),
		LastRunSHA: lastSHA,
	}
}

// MockCIRunQuerier implements CIRunQuerier for testing without GitHub.
type MockCIRunQuerier struct {
	// Runs maps workflow name to the HEAD SHA of the last successful run.
	// An empty string means no successful run exists.
	Runs map[string]string
	// Err is returned by LastSuccessfulRunSHA when non-nil.
	Err error
}

// NewMockCIRunQuerier creates a mock querier with the given workflow->SHA mappings.
func NewMockCIRunQuerier(runs map[string]string) *MockCIRunQuerier {
	if runs == nil {
		runs = make(map[string]string)
	}
	return &MockCIRunQuerier{Runs: runs}
}

// LastSuccessfulRunSHA returns the configured SHA for the workflow, or Err.
func (m *MockCIRunQuerier) LastSuccessfulRunSHA(workflowName string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return m.Runs[workflowName], nil
}
