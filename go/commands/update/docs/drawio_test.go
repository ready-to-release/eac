//go:build L0 && ov

package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrawioResult_ZeroValue(t *testing.T) {
	result := DrawioResult{}
	assert.Equal(t, 0, result.ImagesFound)
	assert.Equal(t, 0, result.CacheHits)
	assert.Equal(t, 0, result.CacheMisses)
	assert.Equal(t, 0, result.Optimized)
	assert.Equal(t, 0, result.Failed)
}
