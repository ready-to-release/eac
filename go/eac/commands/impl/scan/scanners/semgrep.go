// Package scanners provides Semgrep security scanner adapter.
package scanners

import (
	"fmt"
	"io"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// semgrepSASTAdapter adapts the standard ScanFunc signature to the Semgrep SAST implementation.
func semgrepSASTAdapter(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts tool.ScanOptions) (interface{}, error) {
	ctx := GlobalScanContext
	if ctx == nil {
		return nil, fmt.Errorf("scan context not initialized: semgrep sast requires GlobalScanContext to be set")
	}

	if ctx.SemgrepImage == "" {
		return nil, fmt.Errorf("semgrep image not configured in scan context")
	}

	config := ctx.SemgrepConfig
	if config == "" {
		config = "auto"
	}

	// Delegate to existing implementation
	return internal.RunSemgrepSAST(workspaceRoot, moduleRoot, config, ctx.SemgrepImage)
}
