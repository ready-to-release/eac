package docprep

import (
	"bytes"
	"context"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPreprocessContext(t *testing.T) {
	book := &config.Book{Name: "test-book"}
	var buf bytes.Buffer

	pctx := NewPreprocessContext(
		context.Background(),
		book,
		"/workspace",
		"/staging",
		&buf,
		SiteMode{},
	)

	require.NotNil(t, pctx)
	assert.Equal(t, book, pctx.Book)
	assert.Equal(t, "/workspace", pctx.WorkspaceRoot)
	assert.Equal(t, "/staging", pctx.StagingDir)
	assert.Equal(t, "site", pctx.Mode.Name())
	assert.True(t, pctx.WarnAsError)
	assert.NotNil(t, pctx.CmdOutputs)
	assert.Empty(t, pctx.Warnings)
	assert.NotNil(t, pctx.Ctx)
}

func TestPreprocessContextWarn(t *testing.T) {
	book := &config.Book{Name: "test-book"}
	var buf bytes.Buffer

	pctx := NewPreprocessContext(
		context.Background(),
		book,
		"/workspace",
		"/staging",
		&buf,
		SiteMode{},
	)

	pctx.Warn("broken link: %s", "foo.md")

	require.Len(t, pctx.Warnings, 1)
	assert.Equal(t, "broken link: foo.md", pctx.Warnings[0])
	assert.Contains(t, buf.String(), "Warning: broken link: foo.md")
}

func TestPreprocessContextPDFMode(t *testing.T) {
	book := &config.Book{Name: "pdf-book"}
	var buf bytes.Buffer

	pctx := NewPreprocessContext(
		context.Background(),
		book,
		"/workspace",
		"/staging",
		&buf,
		PDFMode{ThemeName: "dark"},
	)

	assert.True(t, pctx.Mode.IsPDF())
	assert.Equal(t, "pdf-dark", pctx.Mode.Name())
}
