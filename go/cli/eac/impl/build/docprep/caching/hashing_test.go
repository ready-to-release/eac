package caching

import (
	"testing"
)

func TestHashMermaidKey_Deterministic(t *testing.T) {
	key := MermaidCacheKey{
		SourceFile: "test.md",
		BlockIndex: 0,
		Code:       "graph TD\n    A --> B",
		Width:      800,
		Height:     600,
		Theme:      "default",
	}

	hash1 := HashMermaidKey(key)
	hash2 := HashMermaidKey(key)

	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashMermaidKey_DifferentInputs(t *testing.T) {
	key1 := MermaidCacheKey{Code: "graph TD\n    A --> B", Theme: "default"}
	key2 := MermaidCacheKey{Code: "graph TD\n    A --> C", Theme: "default"}
	key3 := MermaidCacheKey{Code: "graph TD\n    A --> B", Theme: "dark"}

	hash1 := HashMermaidKey(key1)
	hash2 := HashMermaidKey(key2)
	hash3 := HashMermaidKey(key3)

	if hash1 == hash2 {
		t.Errorf("Different code should produce different hash")
	}
	if hash1 == hash3 {
		t.Errorf("Different theme should produce different hash")
	}
}

func TestHashDrawioKey_Deterministic(t *testing.T) {
	key := DrawioCacheKey{SourceHash: "abc123", MaxWidth: 1200}

	hash1 := HashDrawioKey(key)
	hash2 := HashDrawioKey(key)

	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashDrawioKey_DifferentInputs(t *testing.T) {
	key1 := DrawioCacheKey{SourceHash: "abc123", MaxWidth: 1200}
	key2 := DrawioCacheKey{SourceHash: "def456", MaxWidth: 1200}
	key3 := DrawioCacheKey{SourceHash: "abc123", MaxWidth: 800}

	hash1 := HashDrawioKey(key1)
	hash2 := HashDrawioKey(key2)
	hash3 := HashDrawioKey(key3)

	if hash1 == hash2 {
		t.Errorf("Different source hash should produce different key")
	}
	if hash1 == hash3 {
		t.Errorf("Different max width should produce different key")
	}
}
