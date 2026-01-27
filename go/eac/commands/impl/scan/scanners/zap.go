// Package scanners provides OWASP ZAP security scanner adapter.
package scanners

import (
	"fmt"
	"io"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/core/tool"
)

// zapDASTAdapter adapts the standard ScanFunc signature to the ZAP DAST implementation.
func zapDASTAdapter(workspaceRoot, moduleRoot, outputDir string, logWriter io.Writer, opts tool.ScanOptions) (interface{}, error) {
	ctx := GlobalScanContext
	if ctx == nil {
		return nil, fmt.Errorf("scan context not initialized: zap dast requires GlobalScanContext to be set")
	}

	if ctx.ZAPImage == "" {
		return nil, fmt.Errorf("zap image not configured in scan context")
	}

	if ctx.ZAPTargetURL == "" {
		return nil, fmt.Errorf("zap target url not configured in scan context")
	}

	scanType := ctx.ZAPScanType
	if scanType == "" {
		scanType = internal.ZAPBaseline
	}

	// Delegate to existing implementation
	return internal.RunZAPScan(ctx.ZAPTargetURL, scanType, workspaceRoot, ctx.ZAPImage)
}
