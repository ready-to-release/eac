package test

import "github.com/ready-to-release/eac/go/clibase/registry"

func init() {
	registry.RegisterAll(
		ExportManual,
		ImportManual,
		ListSuites,
		MergeResults,
		Test,
		TestDebug,
	)
}
