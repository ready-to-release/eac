package output

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Ensure passthroughBuffer implements OutputBufferPort.
var _ core.OutputBufferPort = (*passthroughBuffer)(nil)

// passthroughBuffer is a no-op implementation for console mode.
// It allows writes to pass through to stdout/stderr without buffering.
type passthroughBuffer struct{}

// NewPassthrough creates a passthrough buffer that does no capturing.
// Use this for console mode where TUI is disabled.
func NewPassthrough() core.OutputBufferPort {
	return &passthroughBuffer{}
}

// Start is a no-op for passthrough mode.
func (p *passthroughBuffer) Start() error {
	return nil
}

// Stop is a no-op for passthrough mode.
func (p *passthroughBuffer) Stop() {}

// Flush is a no-op for passthrough mode.
func (p *passthroughBuffer) Flush() {}

// IsActive always returns false for passthrough mode.
func (p *passthroughBuffer) IsActive() bool {
	return false
}

// GetBuffered returns nil for passthrough mode since nothing is buffered.
func (p *passthroughBuffer) GetBuffered() (stdout, stderr []byte) {
	return nil, nil
}
