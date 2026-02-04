package tui

import (
	"context"
	"io"
	"testing"
	"time"
)

// mockConsole is a test implementation of Console.
type mockConsole struct {
	name string
}

func (m *mockConsole) Start(ctx context.Context) error                { return nil }
func (m *mockConsole) StartAsync(ctx context.Context)                 {}
func (m *mockConsole) Wait()                                          {}
func (m *mockConsole) Stop()                                          {}
func (m *mockConsole) NewWriter(source string, logWriter io.Writer) io.Writer {
	return logWriter
}
func (m *mockConsole) SendLine(line Line)                                        {}
func (m *mockConsole) WriteResult(text string)                                   {}
func (m *mockConsole) UpdateStatus(status Status)                                {}
func (m *mockConsole) StatusRefreshInterval() time.Duration                      { return 0 }
func (m *mockConsole) SetPhase(phase Phase)                                      {}
func (m *mockConsole) CompletePhase(phase Phase, success bool, summary string)   {}
func (m *mockConsole) WriteToPhase(phase Phase, text string)                     {}
func (m *mockConsole) SetPhaseSummary(phase Phase, summary string)               {}
func (m *mockConsole) StartUoW(moniker, displayName string, weight int)       {}
func (m *mockConsole) MarkUoWRunning(moniker string)                          {}
func (m *mockConsole) MarkUoWComplete(moniker string, exitCode int)           {}
func (m *mockConsole) MarkUoWCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
}
func (m *mockConsole) SendSummary(data *SummaryData)       {}
func (m *mockConsole) SetInitSummary(summary *InitSummary) {}

// mockFactory creates a factory that returns a mockConsole with the given name.
func mockFactory(name string) ConsoleFactory {
	return func(config Config) Console {
		return &mockConsole{name: name}
	}
}

// getName extracts the name from a mockConsole for testing.
func getName(c Console) string {
	if m, ok := c.(*mockConsole); ok {
		return m.name
	}
	return ""
}

func TestRegistry_ExactMatch(t *testing.T) {
	r := NewRegistry()
	r.Register("build", mockFactory("build-tui"))
	r.Register("test", mockFactory("test-tui"))
	r.Register("*", mockFactory("default-tui"))

	tests := []struct {
		command string
		want    string
	}{
		{"build", "build-tui"},
		{"test", "test-tui"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			c := r.NewForCommand(tt.command, Config{})
			if got := getName(c); got != tt.want {
				t.Errorf("NewForCommand(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestRegistry_PrefixMatch(t *testing.T) {
	r := NewRegistry()
	r.Register("work *", mockFactory("work-tui"))
	r.Register("*", mockFactory("default-tui"))

	tests := []struct {
		command string
		want    string
	}{
		{"work create", "work-tui"},
		{"work merge", "work-tui"},
		{"work pull", "work-tui"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			c := r.NewForCommand(tt.command, Config{})
			if got := getName(c); got != tt.want {
				t.Errorf("NewForCommand(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestRegistry_HierarchicalFallback(t *testing.T) {
	r := NewRegistry()
	r.Register("work", mockFactory("work-tui"))
	r.Register("*", mockFactory("default-tui"))

	// "work create" should fall back to "work" binding since no exact match
	c := r.NewForCommand("work create", Config{})
	if got := getName(c); got != "work-tui" {
		t.Errorf("NewForCommand(\"work create\") = %q, want \"work-tui\"", got)
	}
}

func TestRegistry_DefaultFallback(t *testing.T) {
	r := NewRegistry()
	r.Register("*", mockFactory("default-tui"))

	tests := []struct {
		command string
	}{
		{"unknown"},
		{"update"},
		{"get config"},
		{"show modules"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			c := r.NewForCommand(tt.command, Config{})
			if got := getName(c); got != "default-tui" {
				t.Errorf("NewForCommand(%q) = %q, want \"default-tui\"", tt.command, got)
			}
		})
	}
}

func TestRegistry_OverrideSubcommand(t *testing.T) {
	r := NewRegistry()
	r.Register("work", mockFactory("work-general-tui"))
	r.Register("work create", mockFactory("work-create-tui"))
	r.Register("*", mockFactory("default-tui"))

	tests := []struct {
		command string
		want    string
	}{
		// "work create" has specific binding, should use it
		{"work create", "work-create-tui"},
		// "work merge" falls back to "work" binding
		{"work merge", "work-general-tui"},
		// bare "work" uses "work" binding
		{"work", "work-general-tui"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			c := r.NewForCommand(tt.command, Config{})
			if got := getName(c); got != tt.want {
				t.Errorf("NewForCommand(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestRegistry_PrefixOverExact(t *testing.T) {
	// When both "work *" prefix and "work" exact exist,
	// subcommands should match prefix, bare command matches exact
	r := NewRegistry()
	r.Register("work", mockFactory("work-bare-tui"))
	r.Register("work *", mockFactory("work-prefix-tui"))
	r.Register("*", mockFactory("default-tui"))

	tests := []struct {
		command string
		want    string
	}{
		// bare "work" matches exact
		{"work", "work-bare-tui"},
		// "work create" matches prefix
		{"work create", "work-prefix-tui"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			c := r.NewForCommand(tt.command, Config{})
			if got := getName(c); got != tt.want {
				t.Errorf("NewForCommand(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestRegistry_NoDefault_Panics(t *testing.T) {
	r := NewRegistry()
	// No default registered

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when no default and no match, but did not panic")
		}
	}()

	r.NewForCommand("unknown", Config{})
}

func TestRegistry_ListBindings(t *testing.T) {
	r := NewRegistry()
	r.Register("build", mockFactory("build-tui"))
	r.Register("work *", mockFactory("work-tui"))
	r.Register("*", mockFactory("default-tui"))

	bindings := r.ListBindings()
	if len(bindings) != 3 {
		t.Errorf("ListBindings() returned %d bindings, want 3", len(bindings))
	}

	// Check all expected patterns are present
	found := make(map[string]bool)
	for _, b := range bindings {
		found[b] = true
	}

	for _, want := range []string{"build", "work *", "*"} {
		if !found[want] {
			t.Errorf("ListBindings() missing %q", want)
		}
	}
}

func TestRegistry_MustHaveDefault_OK(t *testing.T) {
	r := NewRegistry()
	r.Register("*", mockFactory("default-tui"))

	// Should not panic
	r.MustHaveDefault()
}

func TestRegistry_MustHaveDefault_Panics(t *testing.T) {
	r := NewRegistry()
	// No default registered

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when no default, but did not panic")
		}
	}()

	r.MustHaveDefault()
}

func TestRegistry_ConfigPassthrough(t *testing.T) {
	var capturedConfig Config
	factory := func(config Config) Console {
		capturedConfig = config
		return &mockConsole{name: "test"}
	}

	r := NewRegistry()
	r.Register("*", factory)

	want := Config{
		Height:       50,
		BufferSize:   500,
		ASCIIMode:    true,
		RunPhaseName: "testing",
		CommandName:  "test",
		CommandType:  "test",
	}

	r.NewForCommand("test", want)

	if capturedConfig != want {
		t.Errorf("Config not passed through correctly: got %+v, want %+v", capturedConfig, want)
	}
}

func TestRegistry_DeepSubcommand(t *testing.T) {
	r := NewRegistry()
	r.Register("release", mockFactory("release-tui"))
	r.Register("release check", mockFactory("release-check-tui"))
	r.Register("*", mockFactory("default-tui"))

	tests := []struct {
		command string
		want    string
	}{
		{"release", "release-tui"},
		{"release check", "release-check-tui"},
		{"release check ci", "release-check-tui"}, // Falls back to "release check"
		{"release other", "release-tui"},          // Falls back to "release"
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			c := r.NewForCommand(tt.command, Config{})
			if got := getName(c); got != tt.want {
				t.Errorf("NewForCommand(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}
