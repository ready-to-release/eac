package caching

import (
	"crypto/sha256"
	"fmt"
)

// MermaidCacheKey contains all inputs that affect mermaid rendering.
type MermaidCacheKey struct {
	SourceFile string
	BlockIndex int
	Code       string
	Width      int
	Height     int
	Theme      string
}

// DrawioCacheKey contains all inputs that affect drawio image optimization.
type DrawioCacheKey struct {
	SourcePath string
	SourceHash string
	MaxWidth   int
}

// HashMermaidKey computes a deterministic hash for a mermaid cache key.
func HashMermaidKey(key MermaidCacheKey) string {
	h := sha256.New()
	fmt.Fprintf(h, "code:%s\n", key.Code)
	fmt.Fprintf(h, "width:%d\n", key.Width)
	fmt.Fprintf(h, "height:%d\n", key.Height)
	fmt.Fprintf(h, "theme:%s\n", key.Theme)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// HashDrawioKey computes a deterministic hash for a drawio cache key.
func HashDrawioKey(key DrawioCacheKey) string {
	h := sha256.New()
	fmt.Fprintf(h, "source:%s\n", key.SourceHash)
	fmt.Fprintf(h, "maxWidth:%d\n", key.MaxWidth)
	return fmt.Sprintf("%x", h.Sum(nil))
}
