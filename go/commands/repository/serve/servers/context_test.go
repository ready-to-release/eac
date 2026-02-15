package servers

import "testing"

func TestDefaultServeContext(t *testing.T) {
	ctx := DefaultServeContext()

	if ctx == nil {
		t.Fatal("DefaultServeContext returned nil")
	}

	// Verify default values
	if ctx.StaticSiteImage != "cli-nginx-oci:latest" {
		t.Errorf("StaticSiteImage = %q, want %q", ctx.StaticSiteImage, "cli-nginx-oci:latest")
	}
	if ctx.MkDocsLiveImage != "squidfunk/mkdocs-material:latest" {
		t.Errorf("MkDocsLiveImage = %q, want %q", ctx.MkDocsLiveImage, "squidfunk/mkdocs-material:latest")
	}
	if ctx.StructurizrImage != "structurizr/lite:latest" {
		t.Errorf("StructurizrImage = %q, want %q", ctx.StructurizrImage, "structurizr/lite:latest")
	}
	if ctx.DefaultPort != 9000 {
		t.Errorf("DefaultPort = %d, want %d", ctx.DefaultPort, 9000)
	}
	if ctx.PortRangeMin != 9000 {
		t.Errorf("PortRangeMin = %d, want %d", ctx.PortRangeMin, 9000)
	}
	if ctx.PortRangeMax != 9999 {
		t.Errorf("PortRangeMax = %d, want %d", ctx.PortRangeMax, 9999)
	}
	if ctx.ContainerPort != 8000 {
		t.Errorf("ContainerPort = %d, want %d", ctx.ContainerPort, 8000)
	}
	if !ctx.OpenBrowser {
		t.Error("OpenBrowser should be true by default")
	}
	if ctx.BrowserPath != "/" {
		t.Errorf("BrowserPath = %q, want %q", ctx.BrowserPath, "/")
	}
	if ctx.WatchEnabled {
		t.Error("WatchEnabled should be false by default")
	}
}

func TestServeContext_NilFallback(t *testing.T) {
	// Adapter functions accept *ServeContext parameter.
	// When nil is passed, they should use DefaultServeContext().
	// Verify that DefaultServeContext returns non-nil with expected defaults.
	ctx := DefaultServeContext()
	if ctx == nil {
		t.Fatal("DefaultServeContext should not return nil")
	}
	if ctx.DefaultPort != 9000 {
		t.Errorf("DefaultPort = %d, want 9000", ctx.DefaultPort)
	}
}

func TestServeContext_AllFieldsSettable(t *testing.T) {
	ctx := &ServeContext{
		StaticSiteImage:      "custom-static:v1",
		MkDocsLiveImage:      "custom-mkdocs:v1",
		StructurizrImage:     "custom-structurizr:v1",
		StaticSiteDockerfile: "/path/to/Dockerfile",
		StaticSiteContext:    "/path/to/context",
		DefaultPort:          8080,
		PortRangeMin:         8000,
		PortRangeMax:         8999,
		ContainerPort:        80,
		OpenBrowser:          false,
		BrowserPath:          "/docs/",
		WatchEnabled:         true,
		WatchPaths:           []string{"docs/", "mkdocs.yml"},
		WatchExclude:         []string{"*.pyc", "__pycache__"},
		WorkspaceRoot:        "/workspace",
		ModuleMoniker:        "docs",
		ContentPath:          "/workspace/out/build/docs/site",
	}

	// Verify all fields are set correctly
	if ctx.StaticSiteImage != "custom-static:v1" {
		t.Errorf("StaticSiteImage not set correctly")
	}
	if ctx.MkDocsLiveImage != "custom-mkdocs:v1" {
		t.Errorf("MkDocsLiveImage not set correctly")
	}
	if ctx.StructurizrImage != "custom-structurizr:v1" {
		t.Errorf("StructurizrImage not set correctly")
	}
	if ctx.StaticSiteDockerfile != "/path/to/Dockerfile" {
		t.Errorf("StaticSiteDockerfile not set correctly")
	}
	if ctx.StaticSiteContext != "/path/to/context" {
		t.Errorf("StaticSiteContext not set correctly")
	}
	if ctx.DefaultPort != 8080 {
		t.Errorf("DefaultPort not set correctly")
	}
	if ctx.PortRangeMin != 8000 {
		t.Errorf("PortRangeMin not set correctly")
	}
	if ctx.PortRangeMax != 8999 {
		t.Errorf("PortRangeMax not set correctly")
	}
	if ctx.ContainerPort != 80 {
		t.Errorf("ContainerPort not set correctly")
	}
	if ctx.OpenBrowser {
		t.Errorf("OpenBrowser should be false")
	}
	if ctx.BrowserPath != "/docs/" {
		t.Errorf("BrowserPath not set correctly")
	}
	if !ctx.WatchEnabled {
		t.Errorf("WatchEnabled should be true")
	}
	if len(ctx.WatchPaths) != 2 {
		t.Errorf("WatchPaths length = %d, want 2", len(ctx.WatchPaths))
	}
	if len(ctx.WatchExclude) != 2 {
		t.Errorf("WatchExclude length = %d, want 2", len(ctx.WatchExclude))
	}
	if ctx.WorkspaceRoot != "/workspace" {
		t.Errorf("WorkspaceRoot not set correctly")
	}
	if ctx.ModuleMoniker != "docs" {
		t.Errorf("ModuleMoniker not set correctly")
	}
	if ctx.ContentPath != "/workspace/out/build/docs/site" {
		t.Errorf("ContentPath not set correctly")
	}
}
