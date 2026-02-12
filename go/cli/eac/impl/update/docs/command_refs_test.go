//go:build L0 && ov

package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandRefsResult_ZeroValue(t *testing.T) {
	result := CommandRefsResult{}
	assert.Equal(t, 0, result.ValidCommands)
	assert.Equal(t, 0, result.DocumentedCount)
	assert.Equal(t, 0, result.MissingDocs)
	assert.Equal(t, 0, result.OrphanedDocs)
	assert.Equal(t, 0, result.Created)
	assert.Equal(t, 0, result.Removed)
}
