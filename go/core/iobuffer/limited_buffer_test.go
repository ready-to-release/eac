package iobuffer

import (
	"strings"
	"testing"
)

// TestLimitedBuffer_Write tests writing to a limited buffer.
func TestLimitedBuffer_Write(t *testing.T) {
	tests := []struct {
		name      string
		limit     int64
		writes    []string
		wantTotal int64
		wantError bool
	}{
		{
			name:      "within limit",
			limit:     100,
			writes:    []string{"hello", "world"},
			wantTotal: 10,
			wantError: false,
		},
		{
			name:      "at limit",
			limit:     10,
			writes:    []string{"12345", "67890"},
			wantTotal: 10,
			wantError: false,
		},
		{
			name:      "exceed limit",
			limit:     5,
			writes:    []string{"123456"},
			wantTotal: 5,
			wantError: true,
		},
		{
			name:      "multiple writes exceed limit",
			limit:     10,
			writes:    []string{"123", "456", "789", "abc"},
			wantTotal: 10,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := NewLimitedBuffer(tt.limit)
			var err error

			for _, write := range tt.writes {
				_, err = lb.Write([]byte(write))
			}

			if lb.Total() != tt.wantTotal {
				t.Errorf("LimitedBuffer.Total() = %d, want %d", lb.Total(), tt.wantTotal)
			}

			if (err != nil) != tt.wantError {
				t.Errorf("LimitedBuffer.Write() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestLimitedBuffer_String tests String() method.
func TestLimitedBuffer_String(t *testing.T) {
	lb := NewLimitedBuffer(100)

	writes := []string{"hello", " ", "world"}
	for _, write := range writes {
		lb.Write([]byte(write))
	}

	got := lb.String()
	want := "hello world"

	if got != want {
		t.Errorf("LimitedBuffer.String() = %q, want %q", got, want)
	}
}

// TestLimitedBuffer_ExactLimit tests writing exactly at the limit.
func TestLimitedBuffer_ExactLimit(t *testing.T) {
	lb := NewLimitedBuffer(5)

	n, err := lb.Write([]byte("12345"))
	if err != nil {
		t.Errorf("LimitedBuffer.Write() at exact limit should not error, got: %v", err)
	}

	if n != 5 {
		t.Errorf("LimitedBuffer.Write() wrote %d bytes, want 5", n)
	}

	if lb.Total() != 5 {
		t.Errorf("LimitedBuffer.Total() = %d, want 5", lb.Total())
	}

	// Next write should fail
	_, err = lb.Write([]byte("x"))
	if err == nil {
		t.Error("LimitedBuffer.Write() after limit should error, got nil")
	}
}

// TestLimitedBuffer_PartialWrite tests that partial writes work correctly.
func TestLimitedBuffer_PartialWrite(t *testing.T) {
	lb := NewLimitedBuffer(5)

	// Write 3 bytes
	n, err := lb.Write([]byte("abc"))
	if err != nil || n != 3 {
		t.Fatalf("First write failed: n=%d, err=%v", n, err)
	}

	// Write 5 more bytes, but only 2 should fit
	n, err = lb.Write([]byte("12345"))
	if err == nil {
		t.Error("LimitedBuffer.Write() should error when exceeding limit")
	}

	if n != 2 {
		t.Errorf("LimitedBuffer.Write() partial write n = %d, want 2", n)
	}

	if lb.Total() != 5 {
		t.Errorf("LimitedBuffer.Total() = %d, want 5", lb.Total())
	}

	got := lb.String()
	want := "abc12"
	if got != want {
		t.Errorf("LimitedBuffer.String() = %q, want %q", got, want)
	}
}

// TestLimitedBuffer_LargeData tests with a large data limit (10MB).
func TestLimitedBuffer_LargeData(t *testing.T) {
	const maxSize int64 = 10 * 1024 * 1024 // 10MB
	lb := NewLimitedBuffer(maxSize)

	// Write 1MB of data
	data := strings.Repeat("a", 1024*1024)
	n, err := lb.Write([]byte(data))

	if err != nil {
		t.Errorf("LimitedBuffer.Write() 1MB should not error: %v", err)
	}

	if n != 1024*1024 {
		t.Errorf("LimitedBuffer.Write() wrote %d bytes, want %d", n, 1024*1024)
	}

	// Write more data up to 10MB
	for i := 0; i < 9; i++ {
		lb.Write([]byte(data))
	}

	if lb.Total() != maxSize {
		t.Errorf("LimitedBuffer.Total() = %d, want %d", lb.Total(), maxSize)
	}

	// Next write should fail
	_, err = lb.Write([]byte("x"))
	if err == nil {
		t.Error("LimitedBuffer.Write() beyond max should error")
	}
}

// TestNewLimitedBuffer tests the constructor function.
func TestNewLimitedBuffer(t *testing.T) {
	lb := NewLimitedBuffer(42)
	if lb == nil {
		t.Fatal("NewLimitedBuffer returned nil")
	}
	if lb.Total() != 0 {
		t.Errorf("NewLimitedBuffer Total() = %d, want 0", lb.Total())
	}
	if lb.String() != "" {
		t.Errorf("NewLimitedBuffer String() = %q, want empty string", lb.String())
	}
}
