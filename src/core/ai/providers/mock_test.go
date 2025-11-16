// File: src/core/ai/providers/mock_test.go
package providers

import (
	"context"
	"testing"

	"github.com/ready-to-release/eac/src/core/ai"
)

func TestMockProvider_Execute(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		input          string
		wantContains   string
		wantErr        bool
	}{
		{
			name:         "returns configured response",
			mockResponse: "test response",
			input:        "any input",
			wantContains: "test response",
			wantErr:      false,
		},
		{
			name:         "empty response is valid",
			mockResponse: "",
			input:        "input",
			wantContains: "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewMockProvider(tt.mockResponse)
			ctx := context.Background()

			output, err := provider.Execute(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if output != tt.wantContains {
				t.Errorf("Execute() output = %v, want %v", output, tt.wantContains)
			}
		})
	}
}

func TestMockProvider_Name(t *testing.T) {
	provider := NewMockProvider("response")
	if provider.Name() != "mock" {
		t.Errorf("Name() = %v, want %v", provider.Name(), "mock")
	}
}

func TestMockProvider_WithOptions(t *testing.T) {
	provider := NewMockProvider("response")
	ctx := context.Background()

	// Test that options are accepted without error
	_, err := provider.Execute(ctx, "input",
		ai.WithModel("test-model"),
		ai.WithTemperature(0.5),
		ai.WithMaxTokens(100),
	)

	if err != nil {
		t.Errorf("Execute() with options error = %v, want nil", err)
	}
}
