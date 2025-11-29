// Package testers contains test functions for different module types.
package testers

import (
	"io"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// TestFunc is the signature for module type test functions.
// Parameters: module contract, workspace root, output directory, log writer, report format, suite name
// Returns: exit code
type TestFunc func(*modules.ModuleContract, string, string, io.Writer, string, string) int
