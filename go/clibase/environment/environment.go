// Package environment provides execution environment detection for commands.
// It determines whether code is running in CI, containers, tests, or local development.
package environment

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
)

// Env captures execution environment flags.
type Env struct {
	IsCI           bool // Running in CI (GitHub Actions, GitLab CI, etc.)
	IsContainer    bool // Running in clie-cli container
	IsTestContext  bool // Running inside a test harness
	IsLocalConsole bool // Interactive local development
}

// Detect determines the current execution environment.
// It checks for common test context environment variables.
func Detect() *Env {
	return DetectWithTestVars(
		environments.EnvCLIETestRunID,
		environments.EnvGodogFormat,
		environments.EnvCLIEMockSecurity,
		environments.EnvCLIEMockDocker,
	)
}

// DetectWithTestVars allows customizing which env vars indicate test context.
func DetectWithTestVars(testEnvVars ...string) *Env {
	env := &Env{
		IsCI: os.Getenv(environments.EnvCI) != "" ||
			os.Getenv(environments.EnvGitHubActions) != "" ||
			os.Getenv(environments.EnvGitLabCI) != "",
		IsContainer: logging.GetExecutionContext() == logging.ContextCLIECLI,
	}

	for _, v := range testEnvVars {
		if os.Getenv(v) != "" {
			env.IsTestContext = true
			break
		}
	}

	env.IsLocalConsole = !env.IsCI && !env.IsContainer && !env.IsTestContext
	return env
}

// ShouldUseTUI returns true if TUI mode is appropriate for this environment.
func (e *Env) ShouldUseTUI() bool {
	return e.IsLocalConsole
}

// ValidateTUI checks if TUI can be used and returns an error if not.
// Only returns an error if TUI was explicitly requested in an unsupported environment.
func (e *Env) ValidateTUI(tuiExplicitlySet, useTUI bool) error {
	if !tuiExplicitlySet || !useTUI {
		return nil
	}
	if e.IsCI {
		return fmt.Errorf("--tui cannot be used in CI environments")
	}
	if e.IsContainer {
		return fmt.Errorf("--tui cannot be used in container/extension mode (use local console instead)")
	}
	return nil
}

// ContextName returns a human-readable name for the execution context.
func (e *Env) ContextName() string {
	switch {
	case e.IsCI:
		return "CI"
	case e.IsContainer:
		return "container"
	case e.IsTestContext:
		return "test"
	default:
		return "local"
	}
}
