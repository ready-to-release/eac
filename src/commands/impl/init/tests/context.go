// Package tests provides BDD step definitions for the init command.
package tests

import (
	"github.com/ready-to-release/eac/src/core/git"
	coretesting "github.com/ready-to-release/eac/src/core/testing"
)

// Shared test context variables
var (
	TestMockRepo *git.MockRepository
	SharedCtx    *coretesting.SharedTestContext
)
