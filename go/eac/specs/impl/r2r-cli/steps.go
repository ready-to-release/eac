// Package srccli contains godog step implementations for specs/r2r-cli.
//
// Features:
// - specs/r2r-cli/cli-invocation/
// - specs/r2r-cli/verify-configuration/
//
// All CLI tests use common step definitions from internal/steps.go.
package srccli

import (
	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// RegisterSteps registers all r2r-cli specific step definitions.
// Currently all CLI steps are handled by the common internal steps.
func RegisterSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// r2r-cli tests use common step definitions from internal/steps.go
	// No module-specific steps are currently needed.
}
