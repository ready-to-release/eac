//go:build L0 && ov

package adapters

import (
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestModule creates a ModuleContract with the given moniker and name for testing.
func newTestModule(moniker, description string) *modules.ModuleContract {
	return modules.NewModuleContract(domain.BaseContract{
		Moniker:     moniker,
		Description: description,
	}, ".")
}

// ---------------------------------------------------------------------------
// ModuleContractAdapter
// ---------------------------------------------------------------------------

func TestModuleContractAdapter_NilInput(t *testing.T) {
	result := AdaptModule(nil)
	assert.Nil(t, result, "AdaptModule(nil) should return nil")
}

func TestModuleContractAdapter_DelegatesMethods(t *testing.T) {
	mod := modules.NewModuleContract(domain.BaseContract{
		Moniker:     "test-mod",
		Description: "A test module",
		Group: "core",
		DependsOn:   []string{"dep-a", "dep-b"},
	}, ".")

	adapter := NewModuleContractAdapter(mod)

	tests := []struct {
		name   string
		method func() interface{}
		want   interface{}
	}{
		{"GetMoniker", func() interface{} { return adapter.GetMoniker() }, "test-mod"},
		{"GetName", func() interface{} { return adapter.GetName() }, "test-mod"},
		{"GetDescription", func() interface{} { return adapter.GetDescription() }, "A test module"},
		{"GetGroup", func() interface{} { return adapter.GetGroup() }, "core"},
		{"GetDependsOn", func() interface{} { return adapter.GetDependsOn() }, []string{"dep-a", "dep-b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModuleContractAdapter_Unwrap(t *testing.T) {
	mod := newTestModule("m", "M")
	adapter := NewModuleContractAdapter(mod)
	assert.Same(t, mod, adapter.Unwrap())
}

func TestUnwrapModule(t *testing.T) {
	t.Run("nil port returns nil", func(t *testing.T) {
		assert.Nil(t, UnwrapModule(nil))
	})

	t.Run("adapter returns concrete", func(t *testing.T) {
		mod := newTestModule("m", "M")
		port := AdaptModule(mod)
		assert.Same(t, mod, UnwrapModule(port))
	})
}

func TestAdaptModules(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, AdaptModules(nil))
	})

	t.Run("adapts all modules", func(t *testing.T) {
		mods := []*modules.ModuleContract{
			newTestModule("a", "A"),
			newTestModule("b", "B"),
		}
		ports := AdaptModules(mods)
		require.Len(t, ports, 2)
		assert.Equal(t, "a", ports[0].GetMoniker())
		assert.Equal(t, "b", ports[1].GetMoniker())
	})
}

// ---------------------------------------------------------------------------
// ModuleRegistryAdapter
// ---------------------------------------------------------------------------

func TestModuleRegistryAdapter(t *testing.T) {
	reg := modules.NewRegistry("1.0.0", ".")
	_ = reg.Add(newTestModule("alpha", "Alpha"))
	_ = reg.Add(newTestModule("beta", "Beta"))

	adapter := NewModuleRegistryAdapter(reg)

	t.Run("Count", func(t *testing.T) {
		assert.Equal(t, 2, adapter.Count())
	})

	t.Run("Has", func(t *testing.T) {
		assert.True(t, adapter.Has("alpha"))
		assert.False(t, adapter.Has("missing"))
	})

	t.Run("Get existing", func(t *testing.T) {
		port, ok := adapter.Get("alpha")
		assert.True(t, ok)
		assert.Equal(t, "alpha", port.GetMoniker())
	})

	t.Run("Get missing", func(t *testing.T) {
		_, ok := adapter.Get("missing")
		assert.False(t, ok)
	})

	t.Run("All returns all", func(t *testing.T) {
		all := adapter.All()
		assert.Len(t, all, 2)
	})

	t.Run("AllMonikers sorted", func(t *testing.T) {
		monikers := adapter.AllMonikers()
		assert.Equal(t, []string{"alpha", "beta"}, monikers)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Same(t, reg, adapter.Unwrap())
	})
}

func TestAdaptRegistry_Nil(t *testing.T) {
	assert.Nil(t, AdaptRegistry(nil))
}

// ---------------------------------------------------------------------------
// UnitIDAdapter
// ---------------------------------------------------------------------------

func TestUnitIDAdapter_DelegatesMethods(t *testing.T) {
	id := workunit.UnitID{
		Action:        workunit.ActionBuild,
		Module:        "my-mod",
		ComponentType: "go",
		ComponentName: "main",
		Tool:          "go-build",
		Spec:          "spec-1",
		Extra:         map[string]string{"key": "val"},
	}

	adapter := NewUnitIDAdapter(id)

	tests := []struct {
		name   string
		method func() interface{}
		want   interface{}
	}{
		{"GetAction", func() interface{} { return adapter.GetAction() }, "build"},
		{"GetModule", func() interface{} { return adapter.GetModule() }, "my-mod"},
		{"GetComponentType", func() interface{} { return adapter.GetComponentType() }, "go"},
		{"GetComponentName", func() interface{} { return adapter.GetComponentName() }, "main"},
		{"GetTool", func() interface{} { return adapter.GetTool() }, "go-build"},
		{"GetSpec", func() interface{} { return adapter.GetSpec() }, "spec-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method()
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("GetExtra", func(t *testing.T) {
		assert.Equal(t, "val", adapter.GetExtra()["key"])
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, id, adapter.Unwrap())
	})
}

func TestAdaptUnitID(t *testing.T) {
	id := workunit.UnitID{Module: "test"}
	port := AdaptUnitID(id)
	assert.Equal(t, "test", port.GetModule())
}

// ---------------------------------------------------------------------------
// UnitResultAdapter
// ---------------------------------------------------------------------------

func TestUnitResultAdapter_DelegatesMethods(t *testing.T) {
	result := workunit.UnitResult{
		ID:       workunit.UnitID{Module: "mod-1"},
		ExitCode: 0,
		Duration: 5 * time.Second,
		LogPath:  "/tmp/log.txt",
	}

	adapter := NewUnitResultAdapter(result)

	t.Run("GetExitCode", func(t *testing.T) {
		assert.Equal(t, 0, adapter.GetExitCode())
	})

	t.Run("GetDuration", func(t *testing.T) {
		assert.Equal(t, 5*time.Second, adapter.GetDuration())
	})

	t.Run("GetLogPath", func(t *testing.T) {
		assert.Equal(t, "/tmp/log.txt", adapter.GetLogPath())
	})

	t.Run("Success", func(t *testing.T) {
		assert.True(t, adapter.Success())
	})

	t.Run("Failed", func(t *testing.T) {
		assert.False(t, adapter.Failed())
	})

	t.Run("GetID", func(t *testing.T) {
		assert.Equal(t, "mod-1", adapter.GetID().GetModule())
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, result, adapter.Unwrap())
	})
}

func TestAdaptUnitResults(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, AdaptUnitResults(nil))
	})

	t.Run("adapts all results", func(t *testing.T) {
		results := []workunit.UnitResult{
			{ID: workunit.UnitID{Module: "a"}, ExitCode: 0},
			{ID: workunit.UnitID{Module: "b"}, ExitCode: 1},
		}
		ports := AdaptUnitResults(results)
		require.Len(t, ports, 2)
		assert.True(t, ports[0].Success())
		assert.True(t, ports[1].Failed())
	})
}

// ---------------------------------------------------------------------------
// AdaptSlice generic helper
// ---------------------------------------------------------------------------

func TestAdaptSlice_NilReturnsNil(t *testing.T) {
	var input []int
	result := AdaptSlice(input, func(i int) string { return "" })
	assert.Nil(t, result)
}

func TestAdaptSlice_TransformsAll(t *testing.T) {
	input := []int{1, 2, 3}
	result := AdaptSlice(input, func(i int) int { return i * 2 })
	assert.Equal(t, []int{2, 4, 6}, result)
}

// ---------------------------------------------------------------------------
// Compile-time interface checks (these are already in source but verify in tests)
// ---------------------------------------------------------------------------

func TestCompileTimeInterfaceChecks(t *testing.T) {
	// These assignments verify the adapter types satisfy their port interfaces.
	var _ core.ModuleContractPort = (*ModuleContractAdapter)(nil)
	var _ core.ModuleRegistryPort = (*ModuleRegistryAdapter)(nil)
	var _ core.UnitIDPort = (*UnitIDAdapter)(nil)
	var _ core.UnitSpecPort = (*UnitSpecAdapter)(nil)
	var _ core.UnitResultPort = (*UnitResultAdapter)(nil)
}
