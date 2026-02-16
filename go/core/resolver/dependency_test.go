package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyGraph(t *testing.T) {
	tests := []struct {
		name              string
		module            string
		enabledComponents map[string]string
		buildAfter        map[string][]string
		expectedDeps      map[string][]string
	}{
		{
			name:   "no dependencies",
			module: "test-module",
			enabledComponents: map[string]string{
				"go":     "go",
				"assets": "assets",
			},
			buildAfter:   map[string][]string{},
			expectedDeps: map[string][]string{},
		},
		{
			name:   "container depends on go",
			module: "test-module",
			enabledComponents: map[string]string{
				"go":        "go",
				"container": "container",
			},
			buildAfter: map[string][]string{
				"container": {"go"},
			},
			expectedDeps: map[string][]string{
				"container": {"go"},
			},
		},
		{
			name:   "book depends on structurizr and assets",
			module: "test-module",
			enabledComponents: map[string]string{
				"structurizr": "structurizr",
				"assets":      "assets",
				"book":        "book",
			},
			buildAfter: map[string][]string{
				"book": {"structurizr", "assets"},
			},
			expectedDeps: map[string][]string{
				"book": {"structurizr", "assets"},
			},
		},
		{
			name:   "dependency type not enabled - no edge",
			module: "test-module",
			enabledComponents: map[string]string{
				"container": "container",
				// go is not enabled
			},
			buildAfter: map[string][]string{
				"container": {"go"},
			},
			expectedDeps: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getBuildAfter := func(compType string) []string {
				return tt.buildAfter[compType]
			}

			g := NewDependencyGraph(tt.module, tt.enabledComponents, getBuildAfter)

			require.NotNil(t, g)
			assert.Equal(t, tt.module, g.Module())

			for comp, expectedDeps := range tt.expectedDeps {
				actualDeps := g.DependsOn(comp)
				assert.ElementsMatch(t, expectedDeps, actualDeps, "deps for %s", comp)
			}
		})
	}
}

func TestDependencyGraph_HasCycle(t *testing.T) {
	tests := []struct {
		name              string
		enabledComponents map[string]string
		buildAfter        map[string][]string
		hasCycle          bool
	}{
		{
			name: "no cycle",
			enabledComponents: map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
			},
			buildAfter: map[string][]string{
				"b": {"a"},
				"c": {"b"},
			},
			hasCycle: false,
		},
		{
			name: "direct cycle",
			enabledComponents: map[string]string{
				"a": "a",
				"b": "b",
			},
			buildAfter: map[string][]string{
				"a": {"b"},
				"b": {"a"},
			},
			hasCycle: true,
		},
		{
			name: "indirect cycle",
			enabledComponents: map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
			},
			buildAfter: map[string][]string{
				"a": {"c"},
				"b": {"a"},
				"c": {"b"},
			},
			hasCycle: true,
		},
		{
			name:              "empty graph",
			enabledComponents: map[string]string{},
			buildAfter:        map[string][]string{},
			hasCycle:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getBuildAfter := func(compType string) []string {
				return tt.buildAfter[compType]
			}

			g := NewDependencyGraph("test", tt.enabledComponents, getBuildAfter)
			assert.Equal(t, tt.hasCycle, g.HasCycle())
		})
	}
}

func TestDependencyGraph_TopologicalOrder(t *testing.T) {
	tests := []struct {
		name              string
		enabledComponents map[string]string
		buildAfter        map[string][]string
		wantErr           bool
		validateOrder     func(t *testing.T, order []string)
	}{
		{
			name: "linear chain",
			enabledComponents: map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
			},
			buildAfter: map[string][]string{
				"b": {"a"},
				"c": {"b"},
			},
			wantErr: false,
			validateOrder: func(t *testing.T, order []string) {
				require.Len(t, order, 3)
				// a must come before b, b must come before c
				aIdx := indexOf(order, "a")
				bIdx := indexOf(order, "b")
				cIdx := indexOf(order, "c")
				assert.Less(t, aIdx, bIdx, "a should come before b")
				assert.Less(t, bIdx, cIdx, "b should come before c")
			},
		},
		{
			name: "diamond dependency",
			enabledComponents: map[string]string{
				"a": "a",
				"b": "b",
				"c": "c",
				"d": "d",
			},
			buildAfter: map[string][]string{
				"b": {"a"},
				"c": {"a"},
				"d": {"b", "c"},
			},
			wantErr: false,
			validateOrder: func(t *testing.T, order []string) {
				require.Len(t, order, 4)
				// a must come before b and c, both must come before d
				aIdx := indexOf(order, "a")
				bIdx := indexOf(order, "b")
				cIdx := indexOf(order, "c")
				dIdx := indexOf(order, "d")
				assert.Less(t, aIdx, bIdx, "a should come before b")
				assert.Less(t, aIdx, cIdx, "a should come before c")
				assert.Less(t, bIdx, dIdx, "b should come before d")
				assert.Less(t, cIdx, dIdx, "c should come before d")
			},
		},
		{
			name: "cycle returns error",
			enabledComponents: map[string]string{
				"a": "a",
				"b": "b",
			},
			buildAfter: map[string][]string{
				"a": {"b"},
				"b": {"a"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getBuildAfter := func(compType string) []string {
				return tt.buildAfter[compType]
			}

			g := NewDependencyGraph("test", tt.enabledComponents, getBuildAfter)
			order, err := g.TopologicalOrder()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validateOrder != nil {
				tt.validateOrder(t, order)
			}
		})
	}
}

func TestDependencyGraph_Nodes(t *testing.T) {
	enabledComponents := map[string]string{
		"go":        "go",
		"container": "container",
		"assets":    "assets",
	}

	g := NewDependencyGraph("test", enabledComponents, func(string) []string { return nil })

	nodes := g.Nodes()
	assert.Len(t, nodes, 3)
	assert.ElementsMatch(t, []string{"go", "container", "assets"}, nodes)
}

func TestDependencyGraph_NilGraph(t *testing.T) {
	var g *DependencyGraph

	assert.Nil(t, g.DependsOn("foo"))
	assert.False(t, g.HasCycle())

	order, err := g.TopologicalOrder()
	assert.NoError(t, err)
	assert.Nil(t, order)

	assert.Nil(t, g.Nodes())
	assert.Empty(t, g.Module())
}

// indexOf returns the index of s in slice, or -1 if not found.
func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}
