// Command: get files
// Description: Get repository files with their module ownership
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --as-<name>: Output using custom renderer (e.g., --as-summary)
package main

import (
	"sort"

	"github.com/ready-to-release/eac/src/commands/get"
	"github.com/ready-to-release/eac/src/repository/reports"
)

func init() {
	Register("get files", GetFiles)
}

func GetFiles() int {
	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		// Generate report for all tracked files (tracked only, no ignored, not staged only)
		report, err := reports.GetFilesModulesReport(true, false, false, "../..", "0.1.0")
		if err != nil {
			return nil, err
		}

		// Sort by last module in the list (if multiple modules)
		sort.Slice(report.AllFiles, func(i, j int) bool {
			// Get last module for each file (or empty string if no modules)
			lastModuleI := ""
			if len(report.AllFiles[i].Modules) > 0 {
				lastModuleI = report.AllFiles[i].Modules[len(report.AllFiles[i].Modules)-1]
			}

			lastModuleJ := ""
			if len(report.AllFiles[j].Modules) > 0 {
				lastModuleJ = report.AllFiles[j].Modules[len(report.AllFiles[j].Modules)-1]
			}

			// Sort by last module, then by file name if modules are equal
			if lastModuleI != lastModuleJ {
				return lastModuleI < lastModuleJ
			}
			return report.AllFiles[i].Name < report.AllFiles[j].Name
		})

		return report.AllFiles, nil
	})
}
