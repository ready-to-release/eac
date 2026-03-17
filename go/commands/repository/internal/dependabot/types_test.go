package dependabot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateEntry_IsConsolidated(t *testing.T) {
	t.Run("singular entry is not consolidated", func(t *testing.T) {
		u := UpdateEntry{PackageEcosystem: "gomod", Directory: "/go/core"}
		assert.False(t, u.IsConsolidated())
	})

	t.Run("entry with directories is consolidated", func(t *testing.T) {
		u := UpdateEntry{
			PackageEcosystem: "gomod",
			Directories:      []string{"/go/core", "/go/cli/eac"},
		}
		assert.True(t, u.IsConsolidated())
	})

	t.Run("empty directories is not consolidated", func(t *testing.T) {
		u := UpdateEntry{PackageEcosystem: "gomod", Directories: []string{}}
		assert.False(t, u.IsConsolidated())
	})
}

func TestUpdateEntry_Key(t *testing.T) {
	t.Run("singular returns ecosystem:directory", func(t *testing.T) {
		u := UpdateEntry{PackageEcosystem: "gomod", Directory: "/go/core"}
		assert.Equal(t, "gomod:/go/core", u.Key())
	})

	t.Run("consolidated returns ecosystem:*", func(t *testing.T) {
		u := UpdateEntry{
			PackageEcosystem: "gomod",
			Directories:      []string{"/go/core", "/go/cli/eac"},
		}
		assert.Equal(t, "gomod:*", u.Key())
	})
}
