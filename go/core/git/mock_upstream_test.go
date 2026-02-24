//go:build L1 && ov
// +build L1,ov

package git

import (
	"fmt"
	"testing"
)

// TestMockRepository_UpstreamBranch verifies the UpstreamBranch method on MockRepository.
func TestMockRepository_UpstreamBranch(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *MockRepository
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "returns configured upstream branch",
			setupMock: func() *MockRepository {
				return NewMockRepository("/tmp").
					WithUpstreamBranch("develop")
			},
			want:    "develop",
			wantErr: false,
		},
		{
			name: "returns error when UpstreamBranchError is set",
			setupMock: func() *MockRepository {
				m := NewMockRepository("/tmp").
					WithUpstreamBranch("develop")
				m.UpstreamBranchError = fmt.Errorf("no upstream configured")
				return m
			},
			want:        "",
			wantErr:     true,
			errContains: "no upstream configured",
		},
		{
			name: "returns error in default empty state when no upstream is configured",
			setupMock: func() *MockRepository {
				return NewMockRepository("/tmp")
			},
			want:        "",
			wantErr:     true,
			errContains: "no upstream configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupMock()

			got, err := mock.UpstreamBranch()

			if (err != nil) != tt.wantErr {
				t.Errorf("UpstreamBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if err.Error() != tt.errContains {
					t.Errorf("UpstreamBranch() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			}
			if got != tt.want {
				t.Errorf("UpstreamBranch() = %q, want %q", got, tt.want)
			}
		})
	}
}
