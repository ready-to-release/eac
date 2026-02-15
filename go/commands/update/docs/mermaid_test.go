//go:build L0 && ov

package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMermaidResult_ZeroValue(t *testing.T) {
	result := MermaidResult{}
	assert.Equal(t, 0, result.FilesScanned)
	assert.Equal(t, 0, result.DiagramsFound)
	assert.Equal(t, 0, result.CacheHits)
	assert.Equal(t, 0, result.CacheMisses)
	assert.Equal(t, 0, result.Rendered)
	assert.Equal(t, 0, result.Failed)
}
