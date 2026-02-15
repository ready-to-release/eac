package scan

import (
	"bytes"
	"sync"
	"testing"

	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/core/evidence"
	"github.com/stretchr/testify/assert"
)

func TestRecordScanResult(t *testing.T) {
	t.Run("nil context is no-op", func(t *testing.T) {
		recordScanResult(nil, "mod", true) // must not panic
	})

	t.Run("first result sets value", func(t *testing.T) {
		sctx := &scanContext{scanResults: make(map[string]bool)}
		recordScanResult(sctx, "mod-a", true)
		assert.True(t, sctx.scanResults["mod-a"])
	})

	t.Run("failure overrides previous success", func(t *testing.T) {
		sctx := &scanContext{scanResults: make(map[string]bool)}
		recordScanResult(sctx, "mod-a", true)
		recordScanResult(sctx, "mod-a", false)
		assert.False(t, sctx.scanResults["mod-a"])
	})

	t.Run("success does not override previous failure", func(t *testing.T) {
		sctx := &scanContext{scanResults: make(map[string]bool)}
		recordScanResult(sctx, "mod-a", false)
		recordScanResult(sctx, "mod-a", true)
		assert.False(t, sctx.scanResults["mod-a"])
	})

	t.Run("independent modules tracked separately", func(t *testing.T) {
		sctx := &scanContext{scanResults: make(map[string]bool)}
		recordScanResult(sctx, "mod-a", true)
		recordScanResult(sctx, "mod-b", false)
		assert.True(t, sctx.scanResults["mod-a"])
		assert.False(t, sctx.scanResults["mod-b"])
	})

	t.Run("concurrent access is safe", func(t *testing.T) {
		sctx := &scanContext{scanResults: make(map[string]bool)}
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				recordScanResult(sctx, "mod", n%2 == 0)
			}(i)
		}
		wg.Wait()
		// No panic and result is deterministic: at least one false was written
		// so the final value must be false (failure sticky)
		assert.False(t, sctx.scanResults["mod"])
	})
}

func TestGetScannerEmoji(t *testing.T) {
	tests := []struct {
		scanner evidence.ScannerType
		want    string
	}{
		{evidence.ScannerSBOM, "📦"},
		{evidence.ScannerVuln, "🔍"},
		{evidence.ScannerSecrets, "🔐"},
		{evidence.ScannerIaC, "🏗️"},
		{evidence.ScannerCompliance, "✅"},
		{evidence.ScannerSAST, "🔬"},
		{evidence.ScannerDAST, "🌐"},
		{evidence.ScannerType("unknown"), "🔒"},
	}
	for _, tt := range tests {
		t.Run(string(tt.scanner), func(t *testing.T) {
			assert.Equal(t, tt.want, getScannerEmoji(tt.scanner))
		})
	}
}

func TestLogScannerConfig(t *testing.T) {
	// logScannerConfig calls GetDockerImage(ctx, ...) which needs a non-nil ctx.
	// With an empty ExecutionContext (EACConfig==nil), the image helper returns ""
	// so we verify output keywords without requiring full EACConfig wiring.
	emptyCtx := &cmdframework.ExecutionContext{}
	multiCfg := &MultiScanConfig{
		SBOMFormat:         "cyclonedx-json",
		SemgrepConfig:      "auto",
		ComplianceStandard: "nist-800-53",
	}

	tests := []struct {
		scanner  evidence.ScannerType
		contains string
	}{
		{evidence.ScannerSBOM, "Format"},
		{evidence.ScannerVuln, "Trivy"},
		{evidence.ScannerSecrets, "Trivy"},
		{evidence.ScannerIaC, "Trivy"},
		{evidence.ScannerCompliance, "Compliance"},
		{evidence.ScannerSAST, "Semgrep"},
		{evidence.ScannerDAST, "ZAP"},
	}
	for _, tt := range tests {
		t.Run(string(tt.scanner), func(t *testing.T) {
			var buf bytes.Buffer
			logScannerConfig(tt.scanner, &buf, multiCfg, emptyCtx)
			assert.Contains(t, buf.String(), tt.contains)
		})
	}
}
