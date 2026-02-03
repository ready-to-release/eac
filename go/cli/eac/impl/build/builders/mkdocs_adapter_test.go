package builders

import (
	"testing"
)

func TestMkDocsPreprocessHandler_Registered(t *testing.T) {
	// The handler should be registered via init()
	h := GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("mkdocs-preprocess handler not registered")
	}

	if got := h.Name(); got != "mkdocs-preprocess" {
		t.Errorf("handler Name() = %q, want %q", got, "mkdocs-preprocess")
	}
}

func TestMkDocsPreprocessHandler_HasHandler(t *testing.T) {
	if !HasHandler("mkdocs-preprocess") {
		t.Error("HasHandler(\"mkdocs-preprocess\") = false, want true")
	}
}

func TestMkDocsPreprocessHandler_Requirements(t *testing.T) {
	h := GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("handler not found")
	}

	reqs := h.Requirements()
	if reqs != nil && len(reqs) > 0 {
		t.Errorf("Requirements() = %v, want nil or empty (no Docker required)", reqs)
	}
}

func TestMkDocsPreprocessHandler_IsContainer(t *testing.T) {
	h := GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("handler not found")
	}

	if h.IsContainer() {
		t.Error("IsContainer() = true, want false")
	}
}

func TestMkDocsPreprocessHandler_IsHostInstalled(t *testing.T) {
	h := GetHandler("mkdocs-preprocess")
	if h == nil {
		t.Fatal("handler not found")
	}

	if !h.IsHostInstalled() {
		t.Error("IsHostInstalled() = false, want true")
	}
}

func TestMkDocsPreprocessHandler_InAllHandlers(t *testing.T) {
	handlers := GetAllHandlers()
	if _, exists := handlers["mkdocs-preprocess"]; !exists {
		t.Error("mkdocs-preprocess not found in GetAllHandlers()")
	}
}

// ============================================================================
// site-render-tool handler tests
// ============================================================================

func TestMkDocsSiteHandler_Registered(t *testing.T) {
	h := GetHandler("site-render-tool")
	if h == nil {
		t.Fatal("site-render-tool handler not registered")
	}

	if got := h.Name(); got != "site-render-tool" {
		t.Errorf("handler Name() = %q, want %q", got, "site-render-tool")
	}
}

func TestMkDocsSiteHandler_HasHandler(t *testing.T) {
	if !HasHandler("site-render-tool") {
		t.Error("HasHandler(\"site-render-tool\") = false, want true")
	}
}

func TestMkDocsSiteHandler_Requirements(t *testing.T) {
	h := GetHandler("site-render-tool")
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
	h := GetHandler("site-render-tool")
	if h == nil {
		t.Fatal("handler not found")
	}

	if !h.IsContainer() {
		t.Error("IsContainer() = false, want true")
	}
}

func TestMkDocsSiteHandler_IsHostInstalled(t *testing.T) {
	h := GetHandler("site-render-tool")
	if h == nil {
		t.Fatal("handler not found")
	}

	if h.IsHostInstalled() {
		t.Error("IsHostInstalled() = true, want false")
	}
}

func TestMkDocsSiteHandler_InAllHandlers(t *testing.T) {
	handlers := GetAllHandlers()
	if _, exists := handlers["site-render-tool"]; !exists {
		t.Error("site-render-tool not found in GetAllHandlers()")
	}
}

// ============================================================================
// pdf-tool handler tests
// ============================================================================

func TestMkDocsPDFHandler_Registered(t *testing.T) {
	h := GetHandler("pdf-tool")
	if h == nil {
		t.Fatal("pdf-tool handler not registered")
	}

	if got := h.Name(); got != "pdf-tool" {
		t.Errorf("handler Name() = %q, want %q", got, "pdf-tool")
	}
}

func TestMkDocsPDFHandler_HasHandler(t *testing.T) {
	if !HasHandler("pdf-tool") {
		t.Error("HasHandler(\"pdf-tool\") = false, want true")
	}
}

func TestMkDocsPDFHandler_Requirements(t *testing.T) {
	h := GetHandler("pdf-tool")
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
	h := GetHandler("pdf-tool")
	if h == nil {
		t.Fatal("handler not found")
	}

	if !h.IsContainer() {
		t.Error("IsContainer() = false, want true")
	}
}

func TestMkDocsPDFHandler_IsHostInstalled(t *testing.T) {
	h := GetHandler("pdf-tool")
	if h == nil {
		t.Fatal("handler not found")
	}

	if h.IsHostInstalled() {
		t.Error("IsHostInstalled() = true, want false")
	}
}

func TestMkDocsPDFHandler_InAllHandlers(t *testing.T) {
	handlers := GetAllHandlers()
	if _, exists := handlers["pdf-tool"]; !exists {
		t.Error("pdf-tool not found in GetAllHandlers()")
	}
}
