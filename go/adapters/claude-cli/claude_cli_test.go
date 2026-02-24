//go:build L0

package claudecli

import (
	"testing"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_DefaultModel(t *testing.T) {
	p := New()
	if p.defaultModel != DefaultModel {
		t.Errorf("p.defaultModel = %q, want %q", p.defaultModel, DefaultModel)
	}
}

func TestName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "claude-cli" {
		t.Errorf("Name() = %q, want %q", got, "claude-cli")
	}
}

func TestDefaultModelConstant(t *testing.T) {
	if DefaultModel != ModelHaiku {
		t.Errorf("DefaultModel = %q, want ModelHaiku (%q)", DefaultModel, ModelHaiku)
	}
}

func TestModelConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"ModelHaiku", ModelHaiku, "claude-3-haiku-20240307"},
		{"ModelSonnet", ModelSonnet, "claude-3-5-sonnet-20240620"},
		{"ModelOpus", ModelOpus, "claude-3-opus-20240229"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	p := New()
	var iface ai.Provider = p
	if iface.Name() != "claude-cli" {
		t.Errorf("interface Name() = %q, want %q", iface.Name(), "claude-cli")
	}
}

func TestMapModelToCLIName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"claude-3-haiku-20240307", "haiku"},
		{"claude-3-5-sonnet-20240620", "sonnet"},
		{"claude-3-5-sonnet-20241022", "sonnet"},
		{"claude-3-opus-20240229", "opus"},
		{"some-unknown-model", "some-unknown-model"},
		{"", ""},
		{"haiku", "haiku"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapModelToCLIName(tt.input)
			if got != tt.expected {
				t.Errorf("mapModelToCLIName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRemoveAPIKeyFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no key present",
			input:    []string{"HOME=/home", "PATH=/bin"},
			expected: []string{"HOME=/home", "PATH=/bin"},
		},
		{
			name:     "key removed",
			input:    []string{"ANTHROPIC_API_KEY=secret", "PATH=/bin"},
			expected: []string{"PATH=/bin"},
		},
		{
			name:     "only key present",
			input:    []string{"ANTHROPIC_API_KEY=secret"},
			expected: nil,
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "similar prefix not removed",
			input:    []string{"ANTHROPIC_API_KEY_OTHER=ok", "PATH=/bin"},
			expected: []string{"ANTHROPIC_API_KEY_OTHER=ok", "PATH=/bin"},
		},
		{
			name:     "multiple keys removed",
			input:    []string{"ANTHROPIC_API_KEY=a", "FOO=bar", "ANTHROPIC_API_KEY=b"},
			expected: []string{"FOO=bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeAPIKeyFromEnv(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("removeAPIKeyFromEnv() returned %d items, want %d\ngot:  %v\nwant: %v",
					len(got), len(tt.expected), got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("result[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestInitRegistersFactory(t *testing.T) {
	factory, ok := aiproviders.Get("claude-cli")
	if !ok {
		t.Fatal("init() did not register 'claude-cli' in default registry")
	}
	if factory == nil {
		t.Fatal("registered factory is nil")
	}
}

func TestInitFactory(t *testing.T) {
	factory, _ := aiproviders.Get("claude-cli")
	// claude-cli ignores config (no API key needed).
	provider, err := factory(&ai.ProviderConfig{})
	if err != nil {
		t.Fatalf("factory error = %v", err)
	}
	if provider == nil {
		t.Fatal("factory returned nil provider")
	}
	if provider.Name() != "claude-cli" {
		t.Errorf("provider.Name() = %q, want %q", provider.Name(), "claude-cli")
	}
}
