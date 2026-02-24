//go:build L0

package aitest

import (
	"context"
	"testing"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	"github.com/ready-to-release/eac/go/core/environments"
)

func TestNewTestProvider(t *testing.T) {
	p := NewTestProvider()
	if p == nil {
		t.Fatal("NewTestProvider() returned nil")
	}
}

func TestName(t *testing.T) {
	p := NewTestProvider()
	if got := p.Name(); got != "test" {
		t.Errorf("Name() = %q, want %q", got, "test")
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	p := NewTestProvider()
	var iface ai.Provider = p
	if iface.Name() != "test" {
		t.Errorf("interface Name() = %q, want %q", iface.Name(), "test")
	}
}

func TestExecute_EnvVar(t *testing.T) {
	expected := "mock AI response from env"
	t.Setenv(environments.EnvCLIETestAIResponse, expected)

	p := NewTestProvider()
	got, err := p.Execute(context.Background(), "any input")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got != expected {
		t.Errorf("Execute() = %q, want %q", got, expected)
	}
}

func TestExecute_NoMockConfigured(t *testing.T) {
	// Ensure env var is not set (t.Setenv restores after test).
	t.Setenv(environments.EnvCLIETestAIResponse, "")

	// NOTE: This test may still find a mock file if run from within the
	// actual eac repo that has .eac/test/ai-mock.txt. If so, the test
	// accepts success. In a clean CI environment, expect the error path.
	p := NewTestProvider()
	_, err := p.Execute(context.Background(), "any input")

	// We expect either success (file found in repo) or a specific error.
	if err != nil {
		expectedMsg := "test provider: no mock response configured"
		if got := err.Error(); len(got) < len(expectedMsg) || got[:len(expectedMsg)] != expectedMsg {
			t.Errorf("Execute() error = %q, want prefix %q", got, expectedMsg)
		}
	}
}

func TestInitRegistersFactory(t *testing.T) {
	factory, ok := aiproviders.Get("test")
	if !ok {
		t.Fatal("init() did not register 'test' in default registry")
	}
	if factory == nil {
		t.Fatal("registered factory is nil")
	}
}

func TestInitFactory(t *testing.T) {
	factory, _ := aiproviders.Get("test")
	provider, err := factory(&ai.ProviderConfig{})
	if err != nil {
		t.Fatalf("factory error = %v", err)
	}
	if provider == nil {
		t.Fatal("factory returned nil provider")
	}
	if provider.Name() != "test" {
		t.Errorf("provider.Name() = %q, want %q", provider.Name(), "test")
	}
}
