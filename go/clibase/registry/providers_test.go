package registry

import (
	"sync"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterProvider(t *testing.T) {
	ResetProviders()

	called := false
	RegisterProvider(func() []core.CommandPort {
		called = true
		return []core.CommandPort{
			newMockCommand("get modules", "get-modules", "Get modules"),
		}
	})

	providers := Providers()
	require.Len(t, providers, 1)

	cmds := providers[0]()
	assert.True(t, called)
	assert.Len(t, cmds, 1)
	assert.Equal(t, "get modules", cmds[0].Name())
}

func TestRegisterMultipleProviders(t *testing.T) {
	ResetProviders()

	RegisterProvider(func() []core.CommandPort {
		return []core.CommandPort{
			newMockCommand("build", "build", "Build"),
		}
	})
	RegisterProvider(func() []core.CommandPort {
		return []core.CommandPort{
			newMockCommand("test", "test", "Test"),
		}
	})
	RegisterProvider(func() []core.CommandPort {
		return []core.CommandPort{
			newMockCommand("scan", "scan", "Scan"),
		}
	})

	providers := Providers()
	require.Len(t, providers, 3)

	// Verify ordering is preserved (FIFO).
	names := make([]string, len(providers))
	for i, p := range providers {
		cmds := p()
		require.Len(t, cmds, 1)
		names[i] = cmds[0].Name()
	}
	assert.Equal(t, []string{"build", "test", "scan"}, names)
}

func TestProvidersReturnsCopy(t *testing.T) {
	ResetProviders()

	RegisterProvider(func() []core.CommandPort {
		return []core.CommandPort{
			newMockCommand("build", "build", "Build"),
		}
	})

	// Get the slice and mutate it.
	first := Providers()
	require.Len(t, first, 1)
	first[0] = nil

	// The internal slice must be unaffected.
	second := Providers()
	require.Len(t, second, 1)
	assert.NotNil(t, second[0], "mutating returned slice must not affect internal state")
}

func TestResetProviders(t *testing.T) {
	ResetProviders()

	RegisterProvider(func() []core.CommandPort {
		return []core.CommandPort{
			newMockCommand("build", "build", "Build"),
		}
	})
	require.Len(t, Providers(), 1)

	ResetProviders()
	assert.Empty(t, Providers(), "ResetProviders must clear all providers")
}

func TestConcurrentRegistration(t *testing.T) {
	ResetProviders()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(n int) {
			defer wg.Done()
			RegisterProvider(func() []core.CommandPort {
				return []core.CommandPort{
					newMockCommand("cmd", "cmd", "Cmd"),
				}
			})
			_ = n
		}(i)
	}
	wg.Wait()

	providers := Providers()
	assert.Len(t, providers, goroutines,
		"all concurrent registrations must succeed")
}
