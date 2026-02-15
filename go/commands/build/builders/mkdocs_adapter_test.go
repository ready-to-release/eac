package builders

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/tool"
)

func TestMkDocsPreprocessHandler_Registered(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("mkdocs-preprocess handler not registered")
	}

	if got := h.Name(); got != "mkdocs-preprocess" {
		t.Errorf("handler Name() = %q, want %q", got, "mkdocs-preprocess")
	}
}

func TestMkDocsPreprocessHandler_Requirements(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("handler not found")
	}

	reqs := h.Requirements()
	if reqs != nil && len(reqs) > 0 {
		t.Errorf("Requirements() = %v, want nil or empty (no Docker required)", reqs)
	}
}

func TestMkDocsPreprocessHandler_IsContainer(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("handler not found")
	}

	if h.IsContainer() {
		t.Error("IsContainer() = true, want false")
	}
}

func TestMkDocsPreprocessHandler_IsHostInstalled(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("handler not found")
	}

	if !h.IsHostInstalled() {
		t.Error("IsHostInstalled() = false, want true")
	}
}

// ============================================================================
// mkdocs-render-oci handler tests
// ============================================================================

func TestMkDocsSiteHandler_Registered(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-render-oci")
	if h == nil {
		t.Fatal("mkdocs-render-oci handler not registered")
	}

	if got := h.Name(); got != "mkdocs-render-oci" {
		t.Errorf("handler Name() = %q, want %q", got, "mkdocs-render-oci")
	}
}

func TestMkDocsSiteHandler_Requirements(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-render-oci")
	if h == nil {
		t.Fatal("handler not found")
	}

	reqs := h.Requirements()
	if len(reqs) == 0 {
		t.Error("Requirements() should return [docker]")
	}
	if reqs[0] != "docker" {
		t.Errorf("Requirements()[0] = %q, want %q", reqs[0], "docker")
	}
}

func TestMkDocsSiteHandler_IsContainer(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-render-oci")
	if h == nil {
		t.Fatal("handler not found")
	}

	if !h.IsContainer() {
		t.Error("IsContainer() = false, want true")
	}
}

func TestMkDocsSiteHandler_IsHostInstalled(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("mkdocs-render-oci")
	if h == nil {
		t.Fatal("handler not found")
	}

	if h.IsHostInstalled() {
		t.Error("IsHostInstalled() = true, want false")
	}
}

// ============================================================================
// pdf-oci handler tests
// ============================================================================

func TestMkDocsPDFHandler_Registered(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("pdf-oci")
	if h == nil {
		t.Fatal("pdf-oci handler not registered")
	}

	if got := h.Name(); got != "pdf-oci" {
		t.Errorf("handler Name() = %q, want %q", got, "pdf-oci")
	}
}

func TestMkDocsPDFHandler_Requirements(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("pdf-oci")
	if h == nil {
		t.Fatal("handler not found")
	}

	reqs := h.Requirements()
	if len(reqs) == 0 {
		t.Error("Requirements() should return [docker]")
	}
	if reqs[0] != "docker" {
		t.Errorf("Requirements()[0] = %q, want %q", reqs[0], "docker")
	}
}

func TestMkDocsPDFHandler_IsContainer(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("pdf-oci")
	if h == nil {
		t.Fatal("handler not found")
	}

	if !h.IsContainer() {
		t.Error("IsContainer() = false, want true")
	}
}

func TestMkDocsPDFHandler_IsHostInstalled(t *testing.T) {
	h := tool.GlobalBuildBridge().GetHandler("pdf-oci")
	if h == nil {
		t.Fatal("handler not found")
	}

	if h.IsHostInstalled() {
		t.Error("IsHostInstalled() = true, want false")
	}
}
