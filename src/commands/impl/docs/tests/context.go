// Package tests provides shared test context for docs command.
//
// This file contains shared state and context types used across step definitions.
package tests

import (
	"github.com/docker/docker/client"
	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// TestContext holds shared state between steps for docs command tests.
type TestContext struct {
	// Shared fields
	CommandOutput string
	ExitCode      int
	CommandError  error

	// Docs-specific fields
	ContainerStarted bool
	ContainerURL     string
	ServerPort       int
	DockerAvailable  bool
	DockerClient     *client.Client
}

// Ctx is the shared test context - set by the main test runner
var Ctx *TestContext

// SharedCtx is the new shared context from core/testing.
// This is a pointer to the same context used by the main test runner,
// so changes made here are immediately visible everywhere.
var SharedCtx *coretesting.SharedTestContext

// OriginalRepoRoot stores the actual repository root - set by the main test runner
var OriginalRepoRoot string

// IsolatedTestProjectDir stores the temp directory for isolated test projects - set by the main test runner
var IsolatedTestProjectDir string

// RunCommand is a function to run commands - set by the main test runner
var RunCommand func(cmdLine string) error
