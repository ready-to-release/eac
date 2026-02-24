//go:build L0

package aiproviders_test

import (
	"sort"
	"sync"
	"testing"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
)

// mockFactory returns a ProviderFactory that always returns nil, nil.
// Each call to mockFactory produces a distinct function value so we can
// verify identity with function-pointer comparison tricks.
func mockFactory(name string) ai.ProviderFactory {
	return func(config *ai.ProviderConfig) (ai.Provider, error) {
		return nil, nil
	}
}

func TestRegisterAndGet(t *testing.T) {
	r := aiproviders.NewRegistry()
	factory := mockFactory("test")

	r.Register("test-provider", factory)

	got, ok := r.Get("test-provider")
	if !ok {
		t.Fatal("expected Get to return true for registered provider")
	}
	if got == nil {
		t.Fatal("expected Get to return a non-nil factory")
	}

	// Verify the factory is functional by calling it.
	provider, err := got(&ai.ProviderConfig{})
	if err != nil {
		t.Fatalf("unexpected error from factory: %v", err)
	}
	// Our mock returns nil provider, which is expected.
	if provider != nil {
		t.Fatal("expected nil provider from mock factory")
	}
}

func TestGetUnregisteredProvider(t *testing.T) {
	r := aiproviders.NewRegistry()

	got, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected Get to return false for unregistered provider")
	}
	if got != nil {
		t.Fatal("expected Get to return nil factory for unregistered provider")
	}
}

func TestSupportedProviders(t *testing.T) {
	r := aiproviders.NewRegistry()
	r.Register("alpha", mockFactory("alpha"))
	r.Register("bravo", mockFactory("bravo"))
	r.Register("charlie", mockFactory("charlie"))

	names := r.SupportedProviders()

	if len(names) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(names))
	}

	sort.Strings(names)
	expected := []string{"alpha", "bravo", "charlie"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("SupportedProviders()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestResetForTesting(t *testing.T) {
	r := aiproviders.NewRegistry()
	r.Register("provider-a", mockFactory("a"))
	r.Register("provider-b", mockFactory("b"))

	// Confirm registrations exist before reset.
	if len(r.SupportedProviders()) != 2 {
		t.Fatal("expected 2 providers before reset")
	}

	r.ResetForTesting()

	names := r.SupportedProviders()
	if len(names) != 0 {
		t.Fatalf("expected 0 providers after ResetForTesting, got %d: %v", len(names), names)
	}

	_, ok := r.Get("provider-a")
	if ok {
		t.Fatal("expected Get to return false after ResetForTesting")
	}
}

func TestNewRegistryCreatesIsolatedInstance(t *testing.T) {
	r1 := aiproviders.NewRegistry()
	r2 := aiproviders.NewRegistry()

	r1.Register("only-in-r1", mockFactory("r1"))
	r2.Register("only-in-r2", mockFactory("r2"))

	// r1 should not see r2's registration.
	if _, ok := r1.Get("only-in-r2"); ok {
		t.Fatal("r1 should not contain provider registered in r2")
	}

	// r2 should not see r1's registration.
	if _, ok := r2.Get("only-in-r1"); ok {
		t.Fatal("r2 should not contain provider registered in r1")
	}

	// Each registry should have exactly one provider.
	if len(r1.SupportedProviders()) != 1 {
		t.Fatalf("r1 expected 1 provider, got %d", len(r1.SupportedProviders()))
	}
	if len(r2.SupportedProviders()) != 1 {
		t.Fatalf("r2 expected 1 provider, got %d", len(r2.SupportedProviders()))
	}
}

func TestConcurrentRegisterAndGet(t *testing.T) {
	r := aiproviders.NewRegistry()

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Half the goroutines register providers.
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			name := "provider-" + itoa(i)
			r.Register(name, mockFactory(name))
		}(i)
	}

	// The other half read concurrently.
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			name := "provider-" + itoa(i)
			r.Get(name)
			r.SupportedProviders()
		}(i)
	}

	wg.Wait()

	// After all goroutines finish, every provider should be registered.
	names := r.SupportedProviders()
	if len(names) != numGoroutines {
		t.Fatalf("expected %d providers after concurrent registration, got %d", numGoroutines, len(names))
	}
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
