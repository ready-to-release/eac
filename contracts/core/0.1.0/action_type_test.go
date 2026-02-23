package core

import (
	"testing"
)

func TestGetActionDescriptor_BuiltIn(t *testing.T) {
	cases := []struct {
		action  ActionType
		verb    string
		logFile string
	}{
		{ActionBuild, "Building", "build.log"},
		{ActionTest, "Testing", "test.log"},
		{ActionScan, "Scanning", "scan.log"},
		{ActionLint, "Linting", "lint.log"},
		{ActionAISummary, "Summarizing", "ai-summary.log"},
		{ActionServe, "Serving", "serve.log"},
	}
	for _, c := range cases {
		d, ok := GetActionDescriptor(c.action)
		if !ok {
			t.Errorf("GetActionDescriptor(%q): not found", c.action)
			continue
		}
		if d.Verb != c.verb {
			t.Errorf("GetActionDescriptor(%q).Verb = %q, want %q", c.action, d.Verb, c.verb)
		}
		if d.LogFile != c.logFile {
			t.Errorf("GetActionDescriptor(%q).LogFile = %q, want %q", c.action, d.LogFile, c.logFile)
		}
	}
}

func TestGetActionDescriptor_Unknown(t *testing.T) {
	_, ok := GetActionDescriptor("unknown-action")
	if ok {
		t.Error("GetActionDescriptor(unknown): expected false, got true")
	}
}

func TestRegisterActionType_Custom(t *testing.T) {
	custom := ActionType("deploy")
	RegisterActionType(ActionDescriptor{
		Type:      custom,
		Verb:      "Deploying",
		PastVerb:  "deployed",
		OutputDir: "out/deploy",
		LogFile:   "deploy.log",
	})
	d, ok := GetActionDescriptor(custom)
	if !ok {
		t.Fatal("registered action type not found")
	}
	if d.Verb != "Deploying" {
		t.Errorf("Verb = %q, want Deploying", d.Verb)
	}
}

func TestRegisterActionType_Overwrite(t *testing.T) {
	// Registering an existing type replaces the entry
	RegisterActionType(ActionDescriptor{
		Type:     ActionBuild,
		Verb:     "Constructing",
		PastVerb: "constructed",
	})
	d, _ := GetActionDescriptor(ActionBuild)
	if d.Verb != "Constructing" {
		t.Errorf("overwrite did not take effect: Verb = %q", d.Verb)
	}

	// Restore original so other tests in this package are not affected
	RegisterActionType(ActionDescriptor{ActionBuild, "Building", "built", "out/build", "build.log"})
}
