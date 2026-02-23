package squashmessage

import (
	"strings"
	"testing"
)

// sampleDiff returns a synthetic unified diff that includes both signal sections
// (a Go source file) and noise sections (a go.sum and a package-lock.json).
func sampleDiff() string {
	return `diff --git a/go/core/tool/handler.go b/go/core/tool/handler.go
index abc1234..def5678 100644
--- a/go/core/tool/handler.go
+++ b/go/core/tool/handler.go
@@ -1,5 +1,6 @@
 package tool

+// NewHandler creates a handler.
 func NewHandler() *Handler {
     return &Handler{}
 }
diff --git a/go/adapters/ai/go.sum b/go/adapters/ai/go.sum
index 111..222 100644
--- a/go/adapters/ai/go.sum
+++ b/go/adapters/ai/go.sum
@@ -1,2 +1,3 @@
 golang.org/x/sys v0.40.0 h1:abc=
+golang.org/x/sys v0.41.0 h1:def=
diff --git a/frontend/package-lock.json b/frontend/package-lock.json
index 333..444 100644
--- a/frontend/package-lock.json
+++ b/frontend/package-lock.json
@@ -1,3 +1,4 @@
 {
+  "lockfileVersion": 3,
   "name": "frontend"
 }
diff --git a/tools/runner.lock b/tools/runner.lock
index 555..666 100644
--- a/tools/runner.lock
+++ b/tools/runner.lock
@@ -1 +1,2 @@
 version=1
+checksum=abc123
`
}

func TestFilterNoiseDiff_RemovesSumFiles(t *testing.T) {
	result := filterNoiseDiff(sampleDiff())

	if strings.Contains(result, "go/adapters/ai/go.sum") {
		t.Error("expected go.sum section to be removed")
	}
	if !strings.Contains(result, "go/core/tool/handler.go") {
		t.Error("expected handler.go section to be preserved")
	}
}

func TestFilterNoiseDiff_RemovesPackageLockJSON(t *testing.T) {
	result := filterNoiseDiff(sampleDiff())

	if strings.Contains(result, "package-lock.json") {
		t.Error("expected package-lock.json section to be removed")
	}
}

func TestFilterNoiseDiff_RemovesLockFiles(t *testing.T) {
	result := filterNoiseDiff(sampleDiff())

	if strings.Contains(result, "tools/runner.lock") {
		t.Error("expected .lock section to be removed")
	}
}

func TestFilterNoiseDiff_PreservesSignalContent(t *testing.T) {
	result := filterNoiseDiff(sampleDiff())

	if !strings.Contains(result, "NewHandler") {
		t.Error("expected handler.go content to be preserved")
	}
}

func TestFilterNoiseDiff_EmptyDiff(t *testing.T) {
	result := filterNoiseDiff("")
	if result != "" {
		t.Errorf("expected empty string for empty input, got %q", result)
	}
}

func TestFilterNoiseDiff_AllNoise(t *testing.T) {
	allNoise := `diff --git a/go.sum b/go.sum
index 000..111 100644
--- a/go.sum
+++ b/go.sum
@@ -1 +1,2 @@
 golang.org/x/sys v0.40.0 h1:abc=
+golang.org/x/sys v0.41.0 h1:def=
`
	result := filterNoiseDiff(allNoise)
	if strings.Contains(result, "diff --git") {
		t.Error("expected all sections removed when all are noise")
	}
}

func TestFilterNoiseDiff_NoNoise(t *testing.T) {
	signal := `diff --git a/go/core/tool/handler.go b/go/core/tool/handler.go
index abc..def 100644
--- a/go/core/tool/handler.go
+++ b/go/core/tool/handler.go
@@ -1 +1,2 @@
 package tool
+// changed
`
	result := filterNoiseDiff(signal)
	if !strings.Contains(result, "handler.go") {
		t.Error("expected signal content to be preserved when there is no noise")
	}
}
