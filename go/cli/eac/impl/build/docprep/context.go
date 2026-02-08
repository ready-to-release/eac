package docprep

import (
	"context"
	"fmt"
	"io"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/config"
)

// PreprocessContext holds shared state for all pipeline phases.
// It is created once and passed to every phase. Phases read shared
// state and may write to their designated fields.
type PreprocessContext struct {
	// Immutable inputs (set at construction, never modified)
	Book          *config.Book
	WorkspaceRoot string
	StagingDir    string
	Mode          OutputMode
	Log           Logger
	Ctx           context.Context

	// Mutable shared state (written by phases, read by later phases)
	FileIndex        *staging.FileIndex  // Built after copy, shared across phases
	FileMapper       staging.FileMapper  // Source->staging file mappings from copy
	ReferencedAssets map[string]bool     // Asset paths referenced by markdown (for lazy copying)
	CmdOutputs       map[string]string   // Command outputs for inline insertion
	Warnings         []string            // Collected warnings

	// Module context (optional, for fragment resolution)
	Moniker string // Module moniker, set by caller if available

	// Options
	WarnAsError bool
}

// NewPreprocessContext creates a new context from build inputs.
func NewPreprocessContext(
	ctx context.Context,
	book *config.Book,
	workspaceRoot, stagingDir string,
	logWriter io.Writer,
	mode OutputMode,
) *PreprocessContext {
	return &PreprocessContext{
		Book:          book,
		WorkspaceRoot: workspaceRoot,
		StagingDir:    stagingDir,
		Mode:          mode,
		Log:           NewLogger(logWriter),
		Ctx:           ctx,
		FileMapper:    staging.NewSimpleFileMap(),
		CmdOutputs:    make(map[string]string),
		WarnAsError:   true,
	}
}

// Warn logs a warning and collects it for WarnAsError mode.
func (pctx *PreprocessContext) Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	pctx.Log.Warnf("%s", msg)
	pctx.Warnings = append(pctx.Warnings, msg)
}
