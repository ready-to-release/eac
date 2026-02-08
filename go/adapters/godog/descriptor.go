package godog

import (
	"github.com/ready-to-release/eac/go/clibase/testrunners"
)

func init() {
	testrunners.RegisterDescriptor(&testrunners.TestTypeDescriptor{
		TestType:             "godog",
		IsBDD:                true,
		ComponentType:        "gherkin",
		MonikerStyle:         "feature",
		RunnerFileConvention: "godog_test.go",
		FeatureTestTypeResolver: func(info testrunners.FeatureModuleInfo) bool {
			// Godog owns features for Go modules (or modules without TypeScript)
			return !info.HasTypeScript
		},
	})
}
