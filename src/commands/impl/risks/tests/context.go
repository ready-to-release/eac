// Package tests provides shared test context for risks commands.
package tests

import (
	"github.com/ready-to-release/eac/src/core/git"
	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// TestContext holds shared state between steps for risks commands.
type TestContext struct {
	// Command execution results
	CommandOutput string
	ExitCode      int
	CommandError  error

	// Assessment-specific fields
	StagedFiles     []string
	ChangedFiles    []string
	AllFiles        []string
	GeneratedReport string
	ReportPath      string

	// Create-specific fields
	AssessmentFile     string
	AssessmentContent  string
	RisksFound         int
	ControlsCreated    []string
	OrphanedTags       []string
	ExistingControl    string

	// List-specific fields
	ControlsList    []string
	LinkageReport   string
	OrphanedCount   int
	LinkedCount     int
	MissingLinksCount int

	// AI mocking
	MockAIResponse string

	// File tracking
	CreatedFiles []string
	TestFiles    map[string]string // path -> content
}

// SharedCtx is the shared context from core/testing
var SharedCtx *coretesting.SharedTestContext

// Ctx is the risks-specific test context
var Ctx *TestContext

// OriginalRepoRoot stores the actual repository root
var OriginalRepoRoot string

// IsolatedTestProjectDir stores the temp directory for isolated tests
var IsolatedTestProjectDir string

// RunCommand is a function to run commands
var RunCommand func(cmdLine string) error

// TestMockRepo holds the mock git repository
var TestMockRepo *git.MockRepository

// Reset clears the risks-specific context
func (c *TestContext) Reset() {
	c.CommandOutput = ""
	c.ExitCode = 0
	c.CommandError = nil
	c.StagedFiles = []string{}
	c.ChangedFiles = []string{}
	c.AllFiles = []string{}
	c.GeneratedReport = ""
	c.ReportPath = ""
	c.AssessmentFile = ""
	c.AssessmentContent = ""
	c.RisksFound = 0
	c.ControlsCreated = []string{}
	c.OrphanedTags = []string{}
	c.ExistingControl = ""
	c.ControlsList = []string{}
	c.LinkageReport = ""
	c.OrphanedCount = 0
	c.LinkedCount = 0
	c.MissingLinksCount = 0
	c.MockAIResponse = ""
	c.CreatedFiles = []string{}
	c.TestFiles = make(map[string]string)
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	return &TestContext{
		StagedFiles:      []string{},
		ChangedFiles:     []string{},
		AllFiles:         []string{},
		ControlsCreated:  []string{},
		OrphanedTags:     []string{},
		ControlsList:     []string{},
		CreatedFiles:     []string{},
		TestFiles:        make(map[string]string),
	}
}
