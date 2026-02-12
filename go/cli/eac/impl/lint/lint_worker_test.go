package lint

import (
	"sync"
	"testing"
	"time"
)

func TestRecordLintResult_NewEntry(t *testing.T) {
	tests := []struct {
		name         string
		module       string
		compName     string
		providerName string
		exitCode     int
		duration     time.Duration
		wantSuccess  bool
	}{
		{
			name:         "successful lint creates passing result",
			module:       "mod-a",
			compName:     "go",
			providerName: "golangci-lint",
			exitCode:     0,
			duration:     2 * time.Second,
			wantSuccess:  true,
		},
		{
			name:         "failed lint creates failing result",
			module:       "mod-b",
			compName:     "go",
			providerName: "golangci-lint",
			exitCode:     1,
			duration:     3 * time.Second,
			wantSuccess:  false,
		},
		{
			name:         "zero duration records correctly",
			module:       "mod-c",
			compName:     "yaml",
			providerName: "yamllint",
			exitCode:     0,
			duration:     0,
			wantSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lctx := &lintContext{
				results: make(map[string]*LintModuleResult),
			}

			params := &lintWorkerParams{
				lctx:         lctx,
				module:       tt.module,
				compName:     tt.compName,
				providerName: tt.providerName,
			}

			recordLintResult(params, tt.exitCode, tt.duration)

			resultKey := tt.module + ":" + tt.compName
			result, ok := lctx.results[resultKey]
			if !ok {
				t.Fatalf("expected result for key %q, got none", resultKey)
			}

			if result.Moniker != resultKey {
				t.Errorf("Moniker = %q, want %q", result.Moniker, resultKey)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if result.Duration != tt.duration {
				t.Errorf("Duration = %v, want %v", result.Duration, tt.duration)
			}
			if len(result.Providers) != 1 || result.Providers[0] != tt.providerName {
				t.Errorf("Providers = %v, want [%s]", result.Providers, tt.providerName)
			}
		})
	}
}

func TestRecordLintResult_AggregatesMultipleProviders(t *testing.T) {
	lctx := &lintContext{
		results: make(map[string]*LintModuleResult),
	}

	// First provider succeeds
	params1 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "go",
		providerName: "golangci-lint",
	}
	recordLintResult(params1, 0, 2*time.Second)

	// Second provider also succeeds
	params2 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "go",
		providerName: "staticcheck",
	}
	recordLintResult(params2, 0, 1*time.Second)

	resultKey := "mod-a:go"
	result := lctx.results[resultKey]

	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.Success {
		t.Error("expected Success=true when both providers pass")
	}
	if result.Duration != 3*time.Second {
		t.Errorf("Duration = %v, want %v (aggregated)", result.Duration, 3*time.Second)
	}
	if len(result.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(result.Providers))
	}
	if result.Providers[0] != "golangci-lint" || result.Providers[1] != "staticcheck" {
		t.Errorf("Providers = %v, want [golangci-lint, staticcheck]", result.Providers)
	}
}

func TestRecordLintResult_SecondProviderFailureOverridesSuccess(t *testing.T) {
	lctx := &lintContext{
		results: make(map[string]*LintModuleResult),
	}

	// First provider succeeds
	params1 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "go",
		providerName: "golangci-lint",
	}
	recordLintResult(params1, 0, 1*time.Second)

	// Second provider fails
	params2 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "go",
		providerName: "staticcheck",
	}
	recordLintResult(params2, 1, 2*time.Second)

	result := lctx.results["mod-a:go"]
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Success {
		t.Error("expected Success=false when second provider fails")
	}
}

func TestRecordLintResult_FirstProviderFailureStaysFailure(t *testing.T) {
	lctx := &lintContext{
		results: make(map[string]*LintModuleResult),
	}

	// First provider fails
	params1 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "go",
		providerName: "golangci-lint",
	}
	recordLintResult(params1, 1, 1*time.Second)

	// Second provider succeeds
	params2 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "go",
		providerName: "staticcheck",
	}
	recordLintResult(params2, 0, 1*time.Second)

	result := lctx.results["mod-a:go"]
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Success {
		t.Error("expected Success=false when first provider failed")
	}
}

func TestRecordLintResult_DifferentComponentsAreSeparate(t *testing.T) {
	lctx := &lintContext{
		results: make(map[string]*LintModuleResult),
	}

	params1 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "go",
		providerName: "golangci-lint",
	}
	recordLintResult(params1, 0, 1*time.Second)

	params2 := &lintWorkerParams{
		lctx:         lctx,
		module:       "mod-a",
		compName:     "docs",
		providerName: "markdownlint",
	}
	recordLintResult(params2, 1, 500*time.Millisecond)

	goResult := lctx.results["mod-a:go"]
	docsResult := lctx.results["mod-a:docs"]

	if goResult == nil || docsResult == nil {
		t.Fatal("expected both results to exist")
	}
	if !goResult.Success {
		t.Error("go result should be successful")
	}
	if docsResult.Success {
		t.Error("docs result should be failed")
	}
}

func TestRecordLintResult_ConcurrentAccess(t *testing.T) {
	lctx := &lintContext{
		results: make(map[string]*LintModuleResult),
	}

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			params := &lintWorkerParams{
				lctx:         lctx,
				module:       "mod-concurrent",
				compName:     "go",
				providerName: "golangci-lint",
			}
			recordLintResult(params, 0, time.Millisecond)
		}(i)
	}

	wg.Wait()

	result := lctx.results["mod-concurrent:go"]
	if result == nil {
		t.Fatal("expected result after concurrent writes")
	}
	// First goroutine creates the entry, subsequent ones aggregate.
	// The provider list should have numGoroutines entries total:
	// 1 from the initial creation + (numGoroutines-1) appended.
	if len(result.Providers) != numGoroutines {
		t.Errorf("expected %d providers (one per goroutine), got %d", numGoroutines, len(result.Providers))
	}
	// Duration should be the sum of all individual durations
	expectedDuration := time.Duration(numGoroutines) * time.Millisecond
	if result.Duration != expectedDuration {
		t.Errorf("Duration = %v, want %v", result.Duration, expectedDuration)
	}
}
