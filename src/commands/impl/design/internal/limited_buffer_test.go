package design

import (
	"strings"
	"testing"
)

// TestLimitedBuffer_Write tests writing to a limited buffer
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
			lb := &limitedBuffer{limit: tt.limit}
			var err error

			for _, write := range tt.writes {
				_, err = lb.Write([]byte(write))
			}

			if lb.total != tt.wantTotal {
				t.Errorf("limitedBuffer.total = %d, want %d", lb.total, tt.wantTotal)
			}

			if (err != nil) != tt.wantError {
				t.Errorf("limitedBuffer.Write() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestLimitedBuffer_String tests String() method
func TestLimitedBuffer_String(t *testing.T) {
	lb := &limitedBuffer{limit: 100}

	writes := []string{"hello", " ", "world"}
	for _, write := range writes {
		lb.Write([]byte(write))
	}

	got := lb.String()
	want := "hello world"

	if got != want {
		t.Errorf("limitedBuffer.String() = %q, want %q", got, want)
	}
}

// TestLimitedBuffer_ExactLimit tests writing exactly at the limit
func TestLimitedBuffer_ExactLimit(t *testing.T) {
	lb := &limitedBuffer{limit: 5}

	n, err := lb.Write([]byte("12345"))
	if err != nil {
		t.Errorf("limitedBuffer.Write() at exact limit should not error, got: %v", err)
	}

	if n != 5 {
		t.Errorf("limitedBuffer.Write() wrote %d bytes, want 5", n)
	}

	if lb.total != 5 {
		t.Errorf("limitedBuffer.total = %d, want 5", lb.total)
	}

	// Next write should fail
	_, err = lb.Write([]byte("x"))
	if err == nil {
		t.Error("limitedBuffer.Write() after limit should error, got nil")
	}
}

// TestLimitedBuffer_PartialWrite tests that partial writes work correctly
func TestLimitedBuffer_PartialWrite(t *testing.T) {
	lb := &limitedBuffer{limit: 5}

	// Write 3 bytes
	n, err := lb.Write([]byte("abc"))
	if err != nil || n != 3 {
		t.Fatalf("First write failed: n=%d, err=%v", n, err)
	}

	// Write 5 more bytes, but only 2 should fit
	n, err = lb.Write([]byte("12345"))
	if err == nil {
		t.Error("limitedBuffer.Write() should error when exceeding limit")
	}

	if n != 2 {
		t.Errorf("limitedBuffer.Write() partial write n = %d, want 2", n)
	}

	if lb.total != 5 {
		t.Errorf("limitedBuffer.total = %d, want 5", lb.total)
	}

	got := lb.String()
	want := "abc12"
	if got != want {
		t.Errorf("limitedBuffer.String() = %q, want %q", got, want)
	}
}

// TestLimitedBuffer_LargeData tests with max Docker output size
func TestLimitedBuffer_LargeData(t *testing.T) {
	lb := &limitedBuffer{limit: MaxDockerOutputSize}

	// Write 1MB of data
	data := strings.Repeat("a", 1024*1024)
	n, err := lb.Write([]byte(data))

	if err != nil {
		t.Errorf("limitedBuffer.Write() 1MB should not error: %v", err)
	}

	if n != 1024*1024 {
		t.Errorf("limitedBuffer.Write() wrote %d bytes, want %d", n, 1024*1024)
	}

	// Write more data up to 10MB
	for i := 0; i < 9; i++ {
		lb.Write([]byte(data))
	}

	if lb.total != MaxDockerOutputSize {
		t.Errorf("limitedBuffer.total = %d, want %d", lb.total, MaxDockerOutputSize)
	}

	// Next write should fail
	_, err = lb.Write([]byte("x"))
	if err == nil {
		t.Error("limitedBuffer.Write() beyond max should error")
	}
}
