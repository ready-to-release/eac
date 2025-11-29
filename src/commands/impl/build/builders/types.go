// Package builders contains build functions for different module types.
package builders

import (
	"io"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// BuildOptions contains flags for controlling the build process.
// This is passed through from the dispatch layer.
type BuildOptions struct {
	TidyFirst     bool   // Run go mod tidy before building
	Version       string // Version to inject via ldflags
	Compressed    bool   // Strip debug info with -ldflags "-s -w" (for releases)
	CompressedUPX bool   // Also apply UPX compression after build
}

// BuildFunc is the signature for module type build functions.
type BuildFunc func(*modules.ModuleContract, string, string, io.Writer, BuildOptions) int
