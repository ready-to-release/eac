package builders

import (
	"encoding/json"
	"testing"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/core/repository"
)

func TestDependencyGraphHandlerInterface(t *testing.T) {
	h := &DependencyGraphHandler{}

	if h.Name() != "dependency-graph" {
		t.Errorf("Expected name 'dependency-graph', got '%s'", h.Name())
	}

	reqs := h.Requirements()
	if len(reqs) != 1 || reqs[0] != "docker" {
		t.Errorf("Expected requirements [docker], got %v", reqs)
	}

	if !h.IsContainer() {
		t.Error("Expected IsContainer() to return true")
	}

	if h.IsHostInstalled() {
		t.Error("Expected IsHostInstalled() to return false")
	}

	artifacts := h.ListArtifacts(nil, "/tmp")
	if len(artifacts) != 1 || artifacts[0] != "dependency-graph/" {
		t.Errorf("Expected artifacts ['dependency-graph/'], got %v", artifacts)
	}
}

func TestGetPlantUMLDiagramProducesContent(t *testing.T) {
	graph := &repository.ModuleDependencyGraph{
		Modules: []string{"eac", "eac-core", "eac-docker"},
		Edges: []repository.ModuleDependency{
			{From: "eac", To: "eac-core"},
			{From: "eac-docker", To: "eac-core"},
		},
	}

	puml := repository.GetPlantUMLDiagram(graph)

	if puml == "" {
		t.Fatal("Expected non-empty PlantUML output")
	}

	// Verify it contains expected markers
	if !containsStr(puml, "@startuml") {
		t.Error("Expected @startuml in output")
	}
	if !containsStr(puml, "@enduml") {
		t.Error("Expected @enduml in output")
	}
	if !containsStr(puml, "eac") {
		t.Error("Expected module name 'eac' in output")
	}
	if !containsStr(puml, "eac-core") {
		t.Error("Expected module name 'eac-core' in output")
	}

	t.Logf("✓ Generated PlantUML diagram (%d bytes)", len(puml))
}

func TestDependencyGraphCacheKey(t *testing.T) {
	graph1 := &repository.ModuleDependencyGraph{
		Modules: []string{"a", "b"},
		Edges:   []repository.ModuleDependency{{From: "a", To: "b"}},
	}

	graph2 := &repository.ModuleDependencyGraph{
		Modules: []string{"a", "b", "c"},
		Edges: []repository.ModuleDependency{
			{From: "a", To: "b"},
			{From: "b", To: "c"},
		},
	}

	puml1 := repository.GetPlantUMLDiagram(graph1)
	puml2 := repository.GetPlantUMLDiagram(graph2)

	hash1 := diagrams.HashPlantUMLContent(puml1)
	hash2 := diagrams.HashPlantUMLContent(puml2)

	if hash1 == hash2 {
		t.Error("Different graphs should produce different hashes")
	}

	// Same graph should produce same hash
	hash1b := diagrams.HashPlantUMLContent(puml1)
	if hash1 != hash1b {
		t.Error("Same graph should produce same hash")
	}

	t.Logf("✓ Graph 1 hash: %s, Graph 2 hash: %s", hash1, hash2)
}

func TestDependencyGraphIndexSerialization(t *testing.T) {
	index := DependencyGraphIndex{
		Entries: []DependencyGraphIndexEntry{
			{
				DiagramType: "dependency-graph",
				ContentHash: "abc12345",
				SVGFilename: "dependency-graph.svg",
			},
		},
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal index: %v", err)
	}

	// Verify it can be round-tripped
	var decoded DependencyGraphIndex
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal index: %v", err)
	}

	if len(decoded.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(decoded.Entries))
	}

	if decoded.Entries[0].DiagramType != "dependency-graph" {
		t.Errorf("Expected diagram_type 'dependency-graph', got '%s'", decoded.Entries[0].DiagramType)
	}

	t.Logf("✓ Index JSON:\n%s", string(data))
}

// containsStr is a simple substring check helper.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
