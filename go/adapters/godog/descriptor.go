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
		Language:             "go",
		ParentComponentType:  "go",
		DefaultLevel:         "@L1",
		DefaultDepTag:        "@deps:go",
		OutputArtifacts: []testrunners.ArtifactPattern{
			{ID: "cucumber-report", Pattern: "*.cucumber.json", Type: "cucumber-report"},
		},
		FeatureTestTypeResolver: func(info testrunners.FeatureModuleInfo) bool {
			// Godog owns features for Go modules (or modules without TypeScript)
			return !info.HasTypeScript
		},
	})
}
