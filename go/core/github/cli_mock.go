package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Compile-time interface check
var _ CLIExecutor = (*CLIMock)(nil)

// CLIMock mocks GitHub CLI command execution for testing.
// It supports fixture-based responses and error injection.
type CLIMock struct {
	// Command responses: normalized command -> JSON output
	responses map[string]string

	// Error injection: normalized command -> error
	errors map[string]error

	// Call tracking for assertions
	calls []MockCLICall

	// Workflow runs by workflow name
	workflowRuns map[string][]WorkflowRun

	// Releases
	releases []Release

	// Individual runs by ID (for run view)
	runsById map[int]WorkflowRun

	// Workflows by name (for workflow view)
	workflows map[string]string

	// Tree files by SHA
	treeFiles map[string][]string
}

// MockCLICall records a CLI command invocation.
type MockCLICall struct {
	Command string   // Normalized command (e.g., "run list")
	Args    []string // All arguments
}

// NewCLIMock creates a new GitHub CLI mock with empty collections.
func NewCLIMock() *CLIMock {
	return &CLIMock{
		responses:    make(map[string]string),
		errors:       make(map[string]error),
		calls:        []MockCLICall{},
		workflowRuns: make(map[string][]WorkflowRun),
		releases:     []Release{},
		runsById:     make(map[int]WorkflowRun),
		workflows:    make(map[string]string),
		treeFiles:    make(map[string][]string),
	}
}

// WithWorkflowRuns configures workflow runs for a specific workflow.
// Enables fluent builder pattern: mock.WithWorkflowRuns(...).WithRelease(...)
func (m *CLIMock) WithWorkflowRuns(workflow string, runs []WorkflowRun) *CLIMock {
	m.workflowRuns[workflow] = runs
	// Also index by ID for run view commands
	for _, run := range runs {
		m.runsById[run.ID] = run
	}
	return m
}

// WithWorkflowRun adds a single workflow run (for run view commands).
func (m *CLIMock) WithWorkflowRun(run WorkflowRun) *CLIMock {
	m.runsById[run.ID] = run
	return m
}

// WithRelease adds a release to the mock.
func (m *CLIMock) WithRelease(tag, name string) *CLIMock {
	m.releases = append(m.releases, Release{
		TagName: tag,
		Name:    name,
		Draft:   false,
	})
	return m
}

// WithWorkflow configures a workflow name for workflow view commands.
func (m *CLIMock) WithWorkflow(workflow, name string) *CLIMock {
	m.workflows[workflow] = name
	return m
}

// WithTreeFiles configures tree files for a specific SHA.
func (m *CLIMock) WithTreeFiles(sha string, files []string) *CLIMock {
	m.treeFiles[sha] = files
	return m
}

// WithError configures an error to return for a specific command pattern.
// The pattern is a normalized command string (e.g., "run list -w missing.yml").
func (m *CLIMock) WithError(pattern string, err error) *CLIMock {
	m.errors[pattern] = err
	return m
}

// LoadWorkflowRunsFixture loads workflow runs from a JSON fixture file.
// The fixture path is relative to the cli_fixtures directory.
func (m *CLIMock) LoadWorkflowRunsFixture(workflow, fixturePath string) error {
	data, err := os.ReadFile(filepath.Join("cli_fixtures", fixturePath))
	if err != nil {
		return fmt.Errorf("failed to load fixture: %w", err)
	}

	var runs []WorkflowRun
	if err := json.Unmarshal(data, &runs); err != nil {
		return fmt.Errorf("failed to parse workflow runs fixture: %w", err)
	}

	m.WithWorkflowRuns(workflow, runs)
	return nil
}

// LoadReleasesFixture loads releases from a JSON fixture file.
// The fixture path is relative to the cli_fixtures directory.
func (m *CLIMock) LoadReleasesFixture(fixturePath string) error {
	data, err := os.ReadFile(filepath.Join("cli_fixtures", fixturePath))
	if err != nil {
		return fmt.Errorf("failed to load fixture: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(data, &releases); err != nil {
		return fmt.Errorf("failed to parse releases fixture: %w", err)
	}

	m.releases = releases
	return nil
}

// Exec executes a GitHub CLI command and returns JSON output.
// Implements command parsing and response matching.
func (m *CLIMock) Exec(args ...string) ([]byte, error) {
	return m.ExecContext(context.Background(), args...)
}

// ExecContext executes a GitHub CLI command with context support.
func (m *CLIMock) ExecContext(ctx context.Context, args ...string) ([]byte, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Record the call
	command := normalizeCommand(args)
	m.calls = append(m.calls, MockCLICall{
		Command: command,
		Args:    args,
	})

	// Check for error injection (match against command patterns)
	if err := m.checkErrorInjection(args); err != nil {
		return nil, err
	}

	// Route to appropriate handler based on command
	if len(args) == 0 {
		return nil, errors.New("no command specified")
	}

	switch args[0] {
	case "run":
		return m.handleRunCommand(args[1:])
	case "release":
		return m.handleReleaseCommand(args[1:])
	case "workflow":
		return m.handleWorkflowCommand(args[1:])
	case "api":
		return m.handleAPICommand(args[1:])
	default:
		return nil, fmt.Errorf("unsupported command: %s", args[0])
	}
}

// GetCalls returns all recorded CLI calls.
func (m *CLIMock) GetCalls() []MockCLICall {
	return m.calls
}

// ResetCalls clears the call history.
func (m *CLIMock) ResetCalls() {
	m.calls = []MockCLICall{}
}

// --- Internal command handlers ---

func (m *CLIMock) handleRunCommand(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("run subcommand required")
	}

	switch args[0] {
	case "list":
		return m.handleRunList(args[1:])
	case "view":
		return m.handleRunView(args[1:])
	default:
		return nil, fmt.Errorf("unsupported run subcommand: %s", args[0])
	}
}

func (m *CLIMock) handleRunList(args []string) ([]byte, error) {
	// Parse workflow name from -w flag
	workflow, ok := parseFlagValue(args, "-w")
	if !ok {
		return nil, errors.New("workflow (-w) flag required")
	}

	// Get runs for this workflow
	runs, ok := m.workflowRuns[workflow]
	if !ok {
		runs = []WorkflowRun{}
	}

	// Marshal to JSON
	return json.Marshal(runs)
}

func (m *CLIMock) handleRunView(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("run ID required")
	}

	runID, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid run ID: %s", args[0])
	}

	run, ok := m.runsById[runID]
	if !ok {
		return nil, fmt.Errorf("run %d not found", runID)
	}

	return json.Marshal(run)
}

func (m *CLIMock) handleReleaseCommand(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("release subcommand required")
	}

	switch args[0] {
	case "list":
		return m.handleReleaseList(args[1:])
	case "view":
		return m.handleReleaseView(args[1:])
	default:
		return nil, fmt.Errorf("unsupported release subcommand: %s", args[0])
	}
}

func (m *CLIMock) handleReleaseList(args []string) ([]byte, error) {
	// Return all releases (limit handling can be added if needed)
	return json.Marshal(m.releases)
}

func (m *CLIMock) handleReleaseView(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("release tag required")
	}

	tag := args[0]
	for _, release := range m.releases {
		if release.TagName == tag {
			return json.Marshal(release)
		}
	}

	return nil, fmt.Errorf("release %s not found", tag)
}

func (m *CLIMock) handleWorkflowCommand(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("workflow subcommand required")
	}

	switch args[0] {
	case "view":
		return m.handleWorkflowView(args[1:])
	default:
		return nil, fmt.Errorf("unsupported workflow subcommand: %s", args[0])
	}
}

func (m *CLIMock) handleWorkflowView(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("workflow name required")
	}

	workflow := args[0]
	name, ok := m.workflows[workflow]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", workflow)
	}

	// Return simple JSON with name
	result := map[string]string{"name": name}
	return json.Marshal(result)
}

func (m *CLIMock) handleAPICommand(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("API endpoint required")
	}

	endpoint := args[0]

	// Handle tree API: repos/{owner}/{repo}/git/trees/{sha}
	if strings.Contains(endpoint, "/git/trees/") {
		parts := strings.Split(endpoint, "/")
		if len(parts) < 5 {
			return nil, errors.New("invalid tree API endpoint")
		}
		sha := parts[len(parts)-1]

		files, ok := m.treeFiles[sha]
		if !ok {
			files = []string{}
		}

		// Return as newline-separated list (matching jq output format)
		return []byte(strings.Join(files, "\n")), nil
	}

	return nil, fmt.Errorf("unsupported API endpoint: %s", endpoint)
}

func (m *CLIMock) checkErrorInjection(args []string) error {
	// Build a full command string for matching
	fullCommand := strings.Join(args, " ")

	// Check exact match first
	if err, ok := m.errors[fullCommand]; ok {
		return err
	}

	// Check partial command matches (for flexibility)
	// This allows patterns like "run list -w missing.yml" to match
	// "run list -w missing.yml --json ..."
	for pattern, err := range m.errors {
		if strings.Contains(fullCommand, pattern) {
			return err
		}
	}

	return nil
}

// normalizeCommand creates a normalized command string for matching.
// Example: ["run", "list", "-w", "ci.yml"] -> "run list -w ci.yml"
func normalizeCommand(args []string) string {
	if len(args) == 0 {
		return ""
	}

	// Build base command (first 1-2 words)
	var parts []string
	if len(args) >= 2 {
		parts = []string{args[0], args[1]}
	} else {
		parts = []string{args[0]}
	}

	return strings.Join(parts, " ")
}
