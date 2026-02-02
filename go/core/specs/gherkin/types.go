package gherkin

// ScenarioDetail holds complete scenario metadata including tags and steps.
type ScenarioDetail struct {
	Name        string   // Scenario name
	Tags        []string // Scenario-level tags (includes feature tags)
	Steps       []string // Given/When/Then/And/But steps
	FeatureName string   // Feature title
	FeatureTags []string // Feature-level tags
	FilePath    string   // Source file path
	Description string   // Scenario description
	LineNumber  int      // Line number where scenario starts
}
