package cucumber

import (
	"testing"
)

func TestScenario_GetFailedStepError(t *testing.T) {
	tests := []struct {
		name     string
		scenario Scenario
		want     string
	}{
		{
			name: "returns error from failed step",
			scenario: Scenario{
				Steps: []Step{
					{Name: "step 1", Result: StepResult{Status: "passed"}},
					{Name: "step 2", Result: StepResult{Status: "failed", Error: "expected 5 but got 3"}},
					{Name: "step 3", Result: StepResult{Status: "skipped"}},
				},
			},
			want: "expected 5 but got 3",
		},
		{
			name: "returns empty for all passed steps",
			scenario: Scenario{
				Steps: []Step{
					{Name: "step 1", Result: StepResult{Status: "passed"}},
					{Name: "step 2", Result: StepResult{Status: "passed"}},
				},
			},
			want: "",
		},
		{
			name: "returns empty for failed step without error message",
			scenario: Scenario{
				Steps: []Step{
					{Name: "step 1", Result: StepResult{Status: "failed", Error: ""}},
				},
			},
			want: "",
		},
		{
			name: "returns first failed step error when multiple failures",
			scenario: Scenario{
				Steps: []Step{
					{Name: "step 1", Result: StepResult{Status: "failed", Error: "first error"}},
					{Name: "step 2", Result: StepResult{Status: "failed", Error: "second error"}},
				},
			},
			want: "first error",
		},
		{
			name: "handles empty steps slice",
			scenario: Scenario{
				Steps: []Step{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.scenario.GetFailedStepError()
			if got != tt.want {
				t.Errorf("GetFailedStepError() = %q, want %q", got, tt.want)
			}
		})
	}
}
