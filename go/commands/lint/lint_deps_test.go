package lint

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/cmdframework"
)

func TestLintDepsVerifier_NilModuleRegistry(t *testing.T) {
	// When ModuleRegistry is nil (not yet populated), deps verifier should
	// return verified status with no requirements.
	ctx := &cmdframework.ExecutionContext{
		Config:         &cmdframework.CommandConfig{},
		ModuleRegistry: nil,
	}

	status := lintDepsVerifier(ctx)

	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if !status.Verified {
		t.Error("expected Verified=true for nil ModuleRegistry")
	}
	if len(status.Required) != 0 {
		t.Errorf("expected no Required deps, got %v", status.Required)
	}
	if len(status.Available) != 0 {
		t.Errorf("expected no Available deps, got %v", status.Available)
	}
	if len(status.Missing) != 0 {
		t.Errorf("expected no Missing deps, got %v", status.Missing)
	}
}
