//go:build L0

package claude

import (
	"testing"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
)

func TestNew_ValidArgs(t *testing.T) {
	p, err := New("test-key", "test-model")
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if p == nil {
		t.Fatal("New() returned nil provider")
	}
}

func TestNew_EmptyAPIKey(t *testing.T) {
	p, err := New("", "test-model")
	if err == nil {
		t.Fatal("New() error = nil, want error for empty API key")
	}
	if p != nil {
		t.Fatal("New() returned non-nil provider on error")
	}
}

func TestNew_EmptyModel(t *testing.T) {
	p, err := New("test-key", "")
	if err == nil {
		t.Fatal("New() error = nil, want error for empty model")
	}
	if p != nil {
		t.Fatal("New() returned non-nil provider on error")
	}
}

func TestNew_BothEmpty(t *testing.T) {
	_, err := New("", "")
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}
}

func TestName(t *testing.T) {
	p, _ := New("key", "model")
	if got := p.Name(); got != "claude-api" {
		t.Errorf("Name() = %q, want %q", got, "claude-api")
	}
}

func TestDefaultModel(t *testing.T) {
	if DefaultModel != "claude-3-haiku-20240307" {
		t.Errorf("DefaultModel = %q, want %q", DefaultModel, "claude-3-haiku-20240307")
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	p, _ := New("key", "model")
	var iface ai.Provider = p
	if iface.Name() != "claude-api" {
		t.Errorf("interface Name() = %q, want %q", iface.Name(), "claude-api")
	}
}

func TestNew_StoresModel(t *testing.T) {
	p, _ := New("key", "my-custom-model")
	if p.model != "my-custom-model" {
		t.Errorf("p.model = %q, want %q", p.model, "my-custom-model")
	}
}

func TestInitRegistersFactory(t *testing.T) {
	factory, ok := aiproviders.Get("claude-api")
	if !ok {
		t.Fatal("init() did not register 'claude-api' in default registry")
	}
	if factory == nil {
		t.Fatal("registered factory is nil")
	}
}

func TestInitFactory_ValidConfig(t *testing.T) {
	factory, _ := aiproviders.Get("claude-api")
	provider, err := factory(&ai.ProviderConfig{
		AI: ai.ConnectionConfig{APIKey: "test-key", Model: "test-model"},
	})
	if err != nil {
		t.Fatalf("factory error = %v", err)
	}
	if provider == nil {
		t.Fatal("factory returned nil provider")
	}
	if provider.Name() != "claude-api" {
		t.Errorf("provider.Name() = %q, want %q", provider.Name(), "claude-api")
	}
}

func TestInitFactory_InvalidConfig(t *testing.T) {
	factory, _ := aiproviders.Get("claude-api")
	_, err := factory(&ai.ProviderConfig{
		AI: ai.ConnectionConfig{APIKey: "", Model: "test-model"},
	})
	if err == nil {
		t.Fatal("factory error = nil, want error for empty API key")
	}
}
