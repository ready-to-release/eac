// Command: get modules
// Description: Get all module contracts in the repository
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --as-<name>: Output using custom renderer (e.g., --as-summary, --as-count)
package main

import (
	"github.com/ready-to-release/eac/src/commands/get"
	"github.com/ready-to-release/eac/src/internal/contracts/reports"
)

func init() {
	Register("get modules", GetModules)
}

func GetModules() int {
	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		report, err := reports.GetModuleContracts("../..", "0.1.0")
		if err != nil {
			return nil, err
		}
		return report.Modules, nil
	})
}
