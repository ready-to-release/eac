// Package buildpython contains godog step implementations for specs/eac/build-python.
//
// Features:
// - specs/eac/build-python/specification.feature
//
// These tests verify Python module building against template test repositories.
package buildpython

import (
	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// buildPythonContext holds state between steps for build-python tests.
type buildPythonContext struct {
	sharedCtx *eacgodog.TestContext
}

var bpCtx *buildPythonContext

// RegisterSteps registers all build-python specific step definitions.
func RegisterSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	bpCtx = &buildPythonContext{sharedCtx: ctx}

	sc.Step(`^I use the "([^"]*)" test layout$`, iUseTheTestLayout)
}

func iUseTheTestLayout(layoutName string) error {
	bpCtx.sharedCtx.MustBeIsolated()
	return eacgodog.CopyTestLayout(bpCtx.sharedCtx, layoutName, true)
}
