// Package cucumber provides types for parsing Cucumber JSON test reports
package cucumber

// CucumberReport represents the root structure of a Cucumber JSON report
type CucumberReport []Feature

// Feature represents a Gherkin feature with scenarios
type Feature struct {
	URI         string     `json:"uri"`
	ID          string     `json:"id"`
	Keyword     string     `json:"keyword"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Line        int        `json:"line"`
	Comments    []Comment  `json:"comments"`
	Tags        []Tag      `json:"tags"`
	Elements    []Scenario `json:"elements"`
}

// Comment represents a comment line in the feature file
type Comment struct {
	Value string `json:"value"`
	Line  int    `json:"line"`
}

// Tag represents a Gherkin tag
type Tag struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

// Scenario represents a test scenario
type Scenario struct {
	ID          string `json:"id"`
	Keyword     string `json:"keyword"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Line        int    `json:"line"`
	Type        string `json:"type"`
	Tags        []Tag  `json:"tags"`
	Steps       []Step `json:"steps"`
}

// Step represents a test step with its execution result
type Step struct {
	Keyword string      `json:"keyword"`
	Name    string      `json:"name"`
	Line    int         `json:"line"`
	Match   StepMatch   `json:"match"`
	Result  StepResult  `json:"result"`
}

// StepMatch contains information about the step definition
type StepMatch struct {
	Location string `json:"location"`
}

// StepResult contains the execution result of a step
type StepResult struct {
	Status   string `json:"status"` // "passed", "failed", "skipped", "undefined", "pending"
	Duration int64  `json:"duration,omitempty"`
	Error    string `json:"error_message,omitempty"`
}

// GetFeatureID extracts the Feature ID from comments
// Expected format: "# Feature ID: module_feature-name"
func (f *Feature) GetFeatureID() string {
	for _, comment := range f.Comments {
		if len(comment.Value) > 14 && comment.Value[:14] == "# Feature ID: " {
			return comment.Value[14:]
		}
	}
	return f.ID
}

// GetModule extracts the module name from comments
// Expected format: "# Module: module-name"
func (f *Feature) GetModule() string {
	for _, comment := range f.Comments {
		if len(comment.Value) > 10 && comment.Value[:10] == "# Module: " {
			return comment.Value[10:]
		}
	}
	return ""
}

// HasTag checks if a scenario has a specific tag
func (s *Scenario) HasTag(tagName string) bool {
	for _, tag := range s.Tags {
		if tag.Name == tagName {
			return true
		}
	}
	return false
}

// GetStatus returns the overall status of the scenario
// A scenario passes only if all steps pass
func (s *Scenario) GetStatus() string {
	for _, step := range s.Steps {
		if step.Result.Status != "passed" {
			return step.Result.Status
		}
	}
	return "passed"
}

// GetTagString returns a space-separated string of all tags
func (s *Scenario) GetTagString() string {
	if len(s.Tags) == 0 {
		return ""
	}

	result := s.Tags[0].Name
	for i := 1; i < len(s.Tags); i++ {
		result += " " + s.Tags[i].Name
	}
	return result
}

// GetVerificationType returns the verification type based on tags
// @IV = Installation Verification
// @PV = Performance Verification
// Default = Operational Verification (OV)
func (s *Scenario) GetVerificationType() string {
	if s.HasTag("@IV") {
		return "IV"
	}
	if s.HasTag("@PV") {
		return "PV"
	}
	return "OV"
}

// GetAcceptanceCriteria extracts the AC tag (e.g., "@ac1" -> "AC1")
func (s *Scenario) GetAcceptanceCriteria() string {
	for _, tag := range s.Tags {
		if len(tag.Name) >= 4 && tag.Name[:3] == "@ac" {
			// Convert "@ac1" -> "AC1"
			acNum := tag.Name[3:]
			return "AC" + acNum
		}
	}
	return ""
}
