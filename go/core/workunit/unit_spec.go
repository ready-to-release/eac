package workunit

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

// UnitSpec represents the input specification for a unit of work.
// Aliased from contracts/core/0.1.0 — the canonical definition.
type UnitSpec = core.UnitSpec

// NewBuildSpec creates a UnitSpec for a build operation.
var NewBuildSpec = core.NewBuildSpec

// NewTestSpec creates a UnitSpec for a test operation.
var NewTestSpec = core.NewTestSpec

// NewLintSpec creates a UnitSpec for a lint operation.
var NewLintSpec = core.NewLintSpec

// NewScanSpec creates a UnitSpec for a scan operation.
var NewScanSpec = core.NewScanSpec
