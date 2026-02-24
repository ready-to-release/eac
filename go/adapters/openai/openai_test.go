//go:build L0

package openai

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
	if got := p.Name(); got != "openai" {
		t.Errorf("Name() = %q, want %q", got, "openai")
	}
}

func TestDefaultModel(t *testing.T) {
	if DefaultModel != "gpt-4-turbo" {
		t.Errorf("DefaultModel = %q, want %q", DefaultModel, "gpt-4-turbo")
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	p, _ := New("key", "model")
	var iface ai.Provider = p
	if iface.Name() != "openai" {
		t.Errorf("interface Name() = %q, want %q", iface.Name(), "openai")
	}
}

func TestNew_StoresModel(t *testing.T) {
	p, _ := New("key", "my-model")
	if p.model != "my-model" {
		t.Errorf("p.model = %q, want %q", p.model, "my-model")
	}
}

func TestNew_StoresClient(t *testing.T) {
	p, _ := New("key", "model")
	if p.client == nil {
		t.Fatal("p.client is nil after successful New()")
	}
}

func TestInitRegistersFactory(t *testing.T) {
	factory, ok := aiproviders.Get("openai")
	if !ok {
		t.Fatal("init() did not register 'openai' in default registry")
	}
	if factory == nil {
		t.Fatal("registered factory is nil")
	}
}

func TestInitFactory_ValidConfig(t *testing.T) {
	factory, _ := aiproviders.Get("openai")
	provider, err := factory(&ai.ProviderConfig{
		AI: ai.ConnectionConfig{APIKey: "test-key", Model: "test-model"},
	})
	if err != nil {
		t.Fatalf("factory error = %v", err)
	}
	if provider == nil {
		t.Fatal("factory returned nil provider")
	}
	if provider.Name() != "openai" {
		t.Errorf("provider.Name() = %q, want %q", provider.Name(), "openai")
	}
}

func TestInitFactory_InvalidConfig(t *testing.T) {
	factory, _ := aiproviders.Get("openai")
	_, err := factory(&ai.ProviderConfig{
		AI: ai.ConnectionConfig{APIKey: "", Model: "model"},
	})
	if err == nil {
		t.Fatal("factory error = nil, want error for empty API key")
	}
}
