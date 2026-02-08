package docprep

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext(t *testing.T) (*PreprocessContext, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	pctx := NewPreprocessContext(
		context.Background(),
		&config.Book{Name: "test"},
		"/workspace",
		"/staging",
		&buf,
		SiteMode{},
	)
	pctx.WarnAsError = false
	return pctx, &buf
}

func TestPipelineExecutesInOrder(t *testing.T) {
	var order []string

	p := NewPipeline().
		AddFunc("first", func(_ *PreprocessContext) error {
			order = append(order, "first")
			return nil
		}).
		AddFunc("second", func(_ *PreprocessContext) error {
			order = append(order, "second")
			return nil
		}).
		AddFunc("third", func(_ *PreprocessContext) error {
			order = append(order, "third")
			return nil
		})

	pctx, _ := newTestContext(t)
	err := p.Execute(pctx)

	require.NoError(t, err)
	assert.Equal(t, []string{"first", "second", "third"}, order)
}

func TestPipelineStopsOnFirstError(t *testing.T) {
	var order []string
	errBoom := errors.New("boom")

	p := NewPipeline().
		AddFunc("first", func(_ *PreprocessContext) error {
			order = append(order, "first")
			return nil
		}).
		AddFunc("failing", func(_ *PreprocessContext) error {
			order = append(order, "failing")
			return errBoom
		}).
		AddFunc("third", func(_ *PreprocessContext) error {
			order = append(order, "third")
			return nil
		})

	pctx, _ := newTestContext(t)
	err := p.Execute(pctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, errBoom)
	assert.Contains(t, err.Error(), "phase failing")
	assert.Equal(t, []string{"first", "failing"}, order)
}

func TestPipelineRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var buf bytes.Buffer
	pctx := NewPreprocessContext(
		ctx,
		&config.Book{Name: "test"},
		"/workspace",
		"/staging",
		&buf,
		SiteMode{},
	)
	pctx.WarnAsError = false

	var order []string

	p := NewPipeline().
		AddFunc("first", func(_ *PreprocessContext) error {
			order = append(order, "first")
			cancel() // Cancel context during first phase
			return nil
		}).
		AddFunc("second", func(_ *PreprocessContext) error {
			order = append(order, "second")
			return nil
		})

	err := p.Execute(pctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled before second")
	assert.Equal(t, []string{"first"}, order)
}

func TestPipelineWarnAsError(t *testing.T) {
	pctx, _ := newTestContext(t)
	pctx.WarnAsError = true

	p := NewPipeline().
		AddFunc("warn-phase", func(pctx *PreprocessContext) error {
			pctx.Warn("something bad")
			return nil
		})

	err := p.Execute(pctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "1 warning(s)")
}

func TestPipelineNoWarningsPass(t *testing.T) {
	pctx, _ := newTestContext(t)
	pctx.WarnAsError = true

	p := NewPipeline().
		AddFunc("ok-phase", func(_ *PreprocessContext) error {
			return nil
		})

	err := p.Execute(pctx)
	require.NoError(t, err)
}

func TestPipelineEmpty(t *testing.T) {
	pctx, _ := newTestContext(t)
	p := NewPipeline()

	err := p.Execute(pctx)
	require.NoError(t, err)
}

func TestPipelinePhases(t *testing.T) {
	p := NewPipeline().
		AddFunc("a", func(_ *PreprocessContext) error { return nil }).
		AddFunc("b", func(_ *PreprocessContext) error { return nil })

	phases := p.Phases()
	require.Len(t, phases, 2)
	assert.Equal(t, "a", phases[0].Name())
	assert.Equal(t, "b", phases[1].Name())
}

func TestPipelineAddPhaseInterface(t *testing.T) {
	phase := &mockPhase{name: "mock"}
	p := NewPipeline().Add(phase)

	pctx, _ := newTestContext(t)
	err := p.Execute(pctx)

	require.NoError(t, err)
	assert.True(t, phase.executed)
}

type mockPhase struct {
	name     string
	executed bool
}

func (m *mockPhase) Name() string { return m.name }
func (m *mockPhase) Execute(_ *PreprocessContext) error {
	m.executed = true
	return nil
}

func TestPipelineLogging(t *testing.T) {
	pctx, buf := newTestContext(t)

	p := NewPipeline().
		AddFunc("test-phase", func(_ *PreprocessContext) error {
			return nil
		})

	err := p.Execute(pctx)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Book preprocessing: test (site mode)")
	assert.Contains(t, output, "Phase: test-phase")
	assert.Contains(t, output, "Preprocessing complete: test")
}
