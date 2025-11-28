package tests

import (
	"fmt"

	applytests "github.com/ready-to-release/eac/src/commands/impl/templates/apply/tests"
	installtests "github.com/ready-to-release/eac/src/commands/impl/templates/install/tests"
)

// templatesContext is an interface that all template test contexts must implement
type templatesContext interface {
	WorkDir() string
	TestDir() string
	CommandOutput() string
	ErrorOutput() string
	ExitCode() int
	SetCommandOutput(string)
	SetErrorOutput(string)
	SetExitCode(int)
	Isolation() interface{}
}

// getActiveContext returns the currently active template test context
func getActiveContext() (templatesContext, error) {
	if listCtx != nil {
		return listCtx, nil
	}
	if applyCtx := applytests.GetContext(); applyCtx != nil {
		return applyCtx, nil
	}
	if installCtx := installtests.GetContext(); installCtx != nil {
		return installCtx, nil
	}
	return nil, fmt.Errorf("no template test context initialized")
}

// Implement templatesContext interface for listTestContext
func (c *listTestContext) WorkDir() string {
	return c.workDir
}

func (c *listTestContext) TestDir() string {
	return c.testDir
}

func (c *listTestContext) CommandOutput() string {
	return c.commandOutput
}

func (c *listTestContext) ErrorOutput() string {
	return c.errorOutput
}

func (c *listTestContext) ExitCode() int {
	return c.exitCode
}

func (c *listTestContext) SetCommandOutput(output string) {
	c.commandOutput = output
}

func (c *listTestContext) SetErrorOutput(output string) {
	c.errorOutput = output
}

func (c *listTestContext) SetExitCode(code int) {
	c.exitCode = code
}

func (c *listTestContext) Isolation() interface{} {
	return c.isolation
}
