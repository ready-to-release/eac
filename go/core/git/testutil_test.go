package git

import (
	"testing"
)

func TestLazyRepo_SetAndGet(t *testing.T) {
	lr := &LazyRepo{}

	// Initially nil - Get without a valid path would fail,
	// but Set should override
	mock := NewMockRepository("/test/root")
	lr.Set(mock)

	got, err := lr.Get("/any/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != mock {
		t.Fatalf("expected mock repository, got different instance")
	}
}

func TestLazyRepo_Reset(t *testing.T) {
	lr := &LazyRepo{}

	mock := NewMockRepository("/test/root")
	lr.Set(mock)
	lr.Reset()

	// After reset, repo should be nil so Get would try to open a real repo
	// We cannot test that without a real git repo, but we can verify Reset clears it
	if lr.repo != nil {
		t.Fatal("expected repo to be nil after Reset")
	}
}

func TestLazyRepo_SetNil(t *testing.T) {
	lr := &LazyRepo{}

	mock := NewMockRepository("/test/root")
	lr.Set(mock)
	lr.Set(nil)

	// After setting nil, repo should be nil
	if lr.repo != nil {
		t.Fatal("expected repo to be nil after Set(nil)")
	}
}

func TestLazyRepo_CachesRepo(t *testing.T) {
	lr := &LazyRepo{}

	mock := NewMockRepository("/test/root")
	lr.Set(mock)

	// Get twice - should return same instance
	got1, _ := lr.Get("/path1")
	got2, _ := lr.Get("/path2")

	if got1 != got2 {
		t.Fatal("expected same repository instance on subsequent Get calls")
	}
}
