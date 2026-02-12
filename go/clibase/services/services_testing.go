package services

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
)

// testingTagsAdapter wraps config.TestingTagsConfig
type testingTagsAdapter struct {
	tags *config.TestingTagsConfig
}

func (a *testingTagsAdapter) GetSkipReasons() map[string]string {
	// Convert SkipReason slice to map
	result := make(map[string]string)
	for _, sr := range a.tags.SkipReasons {
		result[sr.Code] = sr.Description
	}
	return result
}
func (a *testingTagsAdapter) GetTags() []string {
	result := make([]string, 0, len(a.tags.Tags))
	for _, tag := range a.tags.Tags {
		result = append(result, tag.Tag)
	}
	return result
}

// testSuitesAdapter wraps config.TestSuitesConfig
type testSuitesAdapter struct {
	suites *config.TestSuitesConfig
}

func (a *testSuitesAdapter) ListDefault() []string { return a.suites.ListDefault() }
func (a *testSuitesAdapter) GetSuiteLTags(suiteName string) []string {
	return a.suites.GetSuiteLTags(suiteName)
}
func (a *testSuitesAdapter) GetSuites() []core.TestSuitePort {
	result := make([]core.TestSuitePort, len(a.suites.Suites))
	for i := range a.suites.Suites {
		result[i] = &testSuiteDefAdapter{s: &a.suites.Suites[i]}
	}
	return result
}

type testSuiteDefAdapter struct {
	s *config.TestSuiteDef
}

func (a *testSuiteDefAdapter) GetName() string   { return a.s.Name }
func (a *testSuiteDefAdapter) GetTags() []string { return nil }                  // Tags are stored differently in selectors
func (a *testSuiteDefAdapter) IsDefault() bool   { return !a.s.ExtendedSuite }   // ExtendedSuite=true means NOT default
