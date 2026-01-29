package cmdframework

import (
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
)

func TestNewSummaryBuilder(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 2,
		"module-b": 3,
	}

	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	if sb == nil {
		t.Fatal("NewSummaryBuilder returned nil")
	}

	if len(sb.moduleCaches) != 2 {
		t.Errorf("expected 2 module caches, got %d", len(sb.moduleCaches))
	}

	if sb.moduleCompCount["module-a"] != 2 {
		t.Errorf("expected module-a count 2, got %d", sb.moduleCompCount["module-a"])
	}

	if sb.moduleCompCount["module-b"] != 3 {
		t.Errorf("expected module-b count 3, got %d", sb.moduleCompCount["module-b"])
	}
}

func TestSummaryBuilder_AddResult(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 2,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	// Add first result
	result1 := orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp1",
		ExitCode:  0,
		Duration:  100 * time.Millisecond,
	}
	sb.AddResult(result1)

	// Check incremental state
	cache := sb.moduleCaches["module-a"]
	if len(cache.components) != 1 {
		t.Errorf("expected 1 component, got %d", len(cache.components))
	}
	if cache.moduleDuration != 100*time.Millisecond {
		t.Errorf("expected duration 100ms, got %v", cache.moduleDuration)
	}

	// Add second result with longer duration
	result2 := orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp2",
		ExitCode:  0,
		Duration:  200 * time.Millisecond,
	}
	sb.AddResult(result2)

	// Check max duration is updated
	if cache.moduleDuration != 200*time.Millisecond {
		t.Errorf("expected duration 200ms, got %v", cache.moduleDuration)
	}

	// Check success count (module should be complete now)
	if sb.successCount != 1 {
		t.Errorf("expected success count 1, got %d", sb.successCount)
	}
}

func TestSummaryBuilder_AddResult_Failure(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 2,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	// Add successful result
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp1",
		ExitCode:  0,
	})

	// Add failed result
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp2",
		ExitCode:  1,
		Errors:    []string{"build failed"},
	})

	cache := sb.moduleCaches["module-a"]
	if !cache.hasFailure {
		t.Error("expected hasFailure to be true")
	}

	if sb.failureCount != 1 {
		t.Errorf("expected failure count 1, got %d", sb.failureCount)
	}

	if sb.successCount != 0 {
		t.Errorf("expected success count 0, got %d", sb.successCount)
	}
}

func TestSummaryBuilder_AddResult_Skipped(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 2,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	// Add two skipped results (exit code -1)
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp1",
		ExitCode:  -1,
	})
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp2",
		ExitCode:  -1,
	})

	cache := sb.moduleCaches["module-a"]
	if !cache.allSkipped {
		t.Error("expected allSkipped to be true")
	}

	if sb.skippedCount != 1 {
		t.Errorf("expected skipped count 1, got %d", sb.skippedCount)
	}
}

func TestSummaryBuilder_Finalize(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 1,
		"module-b": 1,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	// Add results
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "go",
		ExitCode:  0,
		Duration:  100 * time.Millisecond,
	})
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-b",
		Component: "go",
		ExitCode:  0,
		Duration:  200 * time.Millisecond,
	})

	totalTime := 500 * time.Millisecond
	data := sb.Finalize(totalTime)

	if data == nil {
		t.Fatal("Finalize returned nil")
	}

	if !data.Success {
		t.Error("expected Success to be true")
	}

	if data.TotalTime != totalTime {
		t.Errorf("expected TotalTime %v, got %v", totalTime, data.TotalTime)
	}

	if data.RunSummary != "2 built" {
		t.Errorf("unexpected RunSummary: %s", data.RunSummary)
	}

	if len(data.Details) == 0 {
		t.Error("expected non-empty Details")
	}
}

func TestSummaryBuilder_Finalize_WithFailure(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 1,
		"module-b": 1,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "go",
		ExitCode:  0,
	})
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-b",
		Component: "go",
		ExitCode:  1,
		Errors:    []string{"compilation error"},
	})

	data := sb.Finalize(time.Second)

	if data.Success {
		t.Error("expected Success to be false")
	}

	if data.RunSummary != "1 built, 1 failed" {
		t.Errorf("unexpected RunSummary: %s", data.RunSummary)
	}
}

func TestSummaryBuilder_Finalize_MixedStatus(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 1,
		"module-b": 1,
		"module-c": 1,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	// Cached
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "go",
		ExitCode:  -1,
	})
	// Success
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-b",
		Component: "go",
		ExitCode:  0,
	})
	// Failed
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-c",
		Component: "go",
		ExitCode:  1,
	})

	data := sb.Finalize(time.Second)

	if data.RunSummary != "1 cached, 1 built, 1 failed" {
		t.Errorf("unexpected RunSummary: %s", data.RunSummary)
	}
}

func TestSummaryBuilder_GetResultSets(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 2,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "typescript",
		ExitCode:  0,
	})
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "go",
		ExitCode:  0,
	})

	resultSets := sb.GetResultSets()

	if len(resultSets) != 1 {
		t.Fatalf("expected 1 result set, got %d", len(resultSets))
	}

	rs := resultSets[0]
	if rs.Module != "module-a" {
		t.Errorf("expected module-a, got %s", rs.Module)
	}

	if len(rs.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(rs.Components))
	}

	// Components should be sorted alphabetically
	if rs.Components[0].Component != "go" {
		t.Errorf("expected first component 'go', got '%s'", rs.Components[0].Component)
	}
	if rs.Components[1].Component != "typescript" {
		t.Errorf("expected second component 'typescript', got '%s'", rs.Components[1].Component)
	}
}

func TestSummaryBuilder_TestCommand(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 2,
	}
	sb := NewSummaryBuilder(CommandTypeTest, componentCounts)

	sb.AddResult(orchestrator.ComponentResult{
		Module:      "module-a",
		Component:   "unit:gotest",
		ExitCode:    0,
		TestsTotal:  10,
		TestsPassed: 10,
	})
	sb.AddResult(orchestrator.ComponentResult{
		Module:      "module-a",
		Component:   "e2e:godog",
		ExitCode:    0,
		TestsTotal:  5,
		TestsPassed: 5,
	})

	cache := sb.moduleCaches["module-a"]
	if cache.testsTotal != 15 {
		t.Errorf("expected testsTotal 15, got %d", cache.testsTotal)
	}
	if cache.testsPassed != 15 {
		t.Errorf("expected testsPassed 15, got %d", cache.testsPassed)
	}

	data := sb.Finalize(time.Second)
	if !data.Success {
		t.Error("expected Success to be true")
	}
}

func TestSummaryBuilder_ConcurrentAddResult(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 100,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	// Add results concurrently
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func(idx int) {
			sb.AddResult(orchestrator.ComponentResult{
				Module:    "module-a",
				Component: "comp" + string(rune('0'+idx%10)),
				ExitCode:  0,
			})
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	cache := sb.moduleCaches["module-a"]
	if len(cache.components) != 100 {
		t.Errorf("expected 100 components, got %d", len(cache.components))
	}

	if sb.successCount != 1 {
		t.Errorf("expected success count 1, got %d", sb.successCount)
	}
}

func TestSummaryBuilder_CompletionCallback(t *testing.T) {
	componentCounts := map[string]int{
		"module-a": 2,
		"module-b": 1,
	}
	sb := NewSummaryBuilder(CommandTypeBuild, componentCounts)

	// Track callback invocations
	callbackCalled := false
	var callbackBuilder *SummaryBuilder

	sb.SetOnComplete(func(builder *SummaryBuilder) {
		callbackCalled = true
		callbackBuilder = builder
	})

	// Add results - callback should NOT be called yet
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp1",
		ExitCode:  0,
	})
	if callbackCalled {
		t.Error("callback should not be called after first result")
	}

	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-a",
		Component: "comp2",
		ExitCode:  0,
	})
	if callbackCalled {
		t.Error("callback should not be called until all modules complete")
	}

	// Add final result - callback should be called now
	sb.AddResult(orchestrator.ComponentResult{
		Module:    "module-b",
		Component: "comp1",
		ExitCode:  0,
	})

	if !callbackCalled {
		t.Error("callback should have been called when all modules completed")
	}

	if callbackBuilder != sb {
		t.Error("callback should receive the same builder instance")
	}
}
