package registry

import (
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// CommandProvider supplies commands for registration.
// Each command module exposes a Commands() function matching this signature,
// registered via init() using RegisterProvider.
type CommandProvider func() []core.CommandPort

var (
	providerMu sync.Mutex
	providers  []CommandProvider
)

// RegisterProvider registers a command provider. Called from init().
func RegisterProvider(provider CommandProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	providers = append(providers, provider)
}

// Providers returns all registered providers.
// The returned slice is a copy; callers cannot mutate internal state.
func Providers() []CommandProvider {
	providerMu.Lock()
	defer providerMu.Unlock()
	result := make([]CommandProvider, len(providers))
	copy(result, providers)
	return result
}

// ResetProviders clears all providers. For testing only.
func ResetProviders() {
	providerMu.Lock()
	defer providerMu.Unlock()
	providers = nil
}
