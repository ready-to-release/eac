package github

import (
	"fmt"
	"time"
)

// MockAPI implements API for testing.
// Configure the fields to control return values.
type MockAPI struct {
	// Configurable responses
	TreeFiles    map[string][]string      // sha -> files
	WorkflowRuns map[string][]WorkflowRun // workflow -> runs
	Releases     []Release

	// Error injection - set to force errors
	TreeError     error
	RunsError     error
	ReleasesError error

	// Call tracking for assertions
	Calls []MockCall
}

// MockCall records a method call for verification
type MockCall struct {
	Method string
	Args   []interface{}
}

// NewMockAPI creates a new mock with empty collections.
func NewMockAPI() *MockAPI {
	return &MockAPI{
		TreeFiles:    make(map[string][]string),
		WorkflowRuns: make(map[string][]WorkflowRun),
		Releases:     []Release{},
		Calls:        []MockCall{},
	}
}

// GetTreeFiles returns configured tree files or error.
func (m *MockAPI) GetTreeFiles(sha string) ([]string, error) {
	m.Calls = append(m.Calls, MockCall{"GetTreeFiles", []interface{}{sha}})

	if m.TreeError != nil {
		return nil, m.TreeError
	}

	files, ok := m.TreeFiles[sha]
	if !ok {
		return []string{}, nil
	}
	return files, nil
}

// ListRuns returns configured workflow runs or error.
func (m *MockAPI) ListRuns(workflow string, opts ListRunsOpts) ([]WorkflowRun, error) {
	m.Calls = append(m.Calls, MockCall{"ListRuns", []interface{}{workflow, opts}})

	if m.RunsError != nil {
		return nil, m.RunsError
	}

	runs, ok := m.WorkflowRuns[workflow]
	if !ok {
		return []WorkflowRun{}, nil
	}

	// Apply filters
	var filtered []WorkflowRun
	for _, run := range runs {
		if opts.Status != "" && run.Status != opts.Status && run.Conclusion != opts.Status {
			continue
		}
		filtered = append(filtered, run)
		if opts.Limit > 0 && len(filtered) >= opts.Limit {
			break
		}
	}

	return filtered, nil
}

// FindRunBySHA finds a run with matching SHA.
func (m *MockAPI) FindRunBySHA(workflow, sha string, limit int) (*WorkflowRun, error) {
	m.Calls = append(m.Calls, MockCall{"FindRunBySHA", []interface{}{workflow, sha, limit}})

	if m.RunsError != nil {
		return nil, m.RunsError
	}

	runs, ok := m.WorkflowRuns[workflow]
	if !ok {
		return nil, nil
	}

	for _, run := range runs {
		if run.HeadSHA == sha {
			return &run, nil
		}
	}

	return nil, nil
}

// HasRecentSuccess checks for recent successful run.
func (m *MockAPI) HasRecentSuccess(workflow, sha string, since time.Duration) (bool, error) {
	m.Calls = append(m.Calls, MockCall{"HasRecentSuccess", []interface{}{workflow, sha, since}})

	if m.RunsError != nil {
		return false, m.RunsError
	}

	runs, ok := m.WorkflowRuns[workflow]
	if !ok {
		return false, nil
	}

	cutoff := time.Now().Add(-since)
	for _, run := range runs {
		if run.HeadSHA == sha && run.Conclusion == "success" && run.CreatedAt.After(cutoff) {
			return true, nil
		}
	}

	return false, nil
}

// ListReleases returns configured releases or error.
func (m *MockAPI) ListReleases(limit int) ([]Release, error) {
	m.Calls = append(m.Calls, MockCall{"ListReleases", []interface{}{limit}})

	if m.ReleasesError != nil {
		return nil, m.ReleasesError
	}

	if limit <= 0 || limit > len(m.Releases) {
		return m.Releases, nil
	}
	return m.Releases[:limit], nil
}

// ReleaseExists checks if a release tag exists.
func (m *MockAPI) ReleaseExists(tag string) (bool, error) {
	m.Calls = append(m.Calls, MockCall{"ReleaseExists", []interface{}{tag}})

	if m.ReleasesError != nil {
		return false, m.ReleasesError
	}

	for _, r := range m.Releases {
		if r.TagName == tag {
			return true, nil
		}
	}
	return false, nil
}

// --- Helper methods for test setup ---

// AddTreeFiles adds files for a SHA.
func (m *MockAPI) AddTreeFiles(sha string, files []string) *MockAPI {
	m.TreeFiles[sha] = files
	return m
}

// AddRun adds a workflow run.
func (m *MockAPI) AddRun(workflow string, run WorkflowRun) *MockAPI {
	m.WorkflowRuns[workflow] = append(m.WorkflowRuns[workflow], run)
	return m
}

// AddSuccessRun adds a successful run for convenience.
func (m *MockAPI) AddSuccessRun(workflow, sha string) *MockAPI {
	return m.AddRun(workflow, WorkflowRun{
		ID:         len(m.WorkflowRuns[workflow]) + 1,
		HeadSHA:    sha,
		Status:     "completed",
		Conclusion: "success",
		CreatedAt:  time.Now(),
	})
}

// AddRelease adds a release.
func (m *MockAPI) AddRelease(tag, name string) *MockAPI {
	m.Releases = append(m.Releases, Release{
		TagName: tag,
		Name:    name,
	})
	return m
}

// CallCount returns how many times a method was called.
func (m *MockAPI) CallCount(method string) int {
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// LastCall returns the last call to a method, or nil if never called.
func (m *MockAPI) LastCall(method string) *MockCall {
	for i := len(m.Calls) - 1; i >= 0; i-- {
		if m.Calls[i].Method == method {
			return &m.Calls[i]
		}
	}
	return nil
}

// Reset clears all call history.
func (m *MockAPI) Reset() {
	m.Calls = []MockCall{}
}

// Verify panics if the method was not called.
func (m *MockAPI) Verify(method string) {
	if m.CallCount(method) == 0 {
		panic(fmt.Sprintf("expected %s to be called", method))
	}
}
