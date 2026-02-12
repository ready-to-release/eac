//go:build L0 && ov

package generation

import (
	"context"
	"testing"

	"github.com/ready-to-release/eac/go/core/validation"
	"github.com/stretchr/testify/assert"
)

func TestStandardStrategy_Name(t *testing.T) {
	s := &StandardStrategy{}
	assert.Equal(t, "standard", s.Name())
}

func TestStandardStrategy_ShouldRetry(t *testing.T) {
	s := &StandardStrategy{}
	ctx := context.Background()

	t.Run("retries when under max attempts", func(t *testing.T) {
		errors := []validation.ValidationError{
			*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0),
		}
		assert.True(t, s.ShouldRetry(ctx, errors, 1, 3))
	})

	t.Run("does not retry at max attempts", func(t *testing.T) {
		errors := []validation.ValidationError{
			*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0),
		}
		assert.False(t, s.ShouldRetry(ctx, errors, 3, 3))
	})
}

func TestStandardStrategy_BuildRetryPrompt(t *testing.T) {
	s := &StandardStrategy{}
	ctx := context.Background()

	errors := []validation.ValidationError{
		*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0),
	}

	prompt := s.BuildRetryPrompt(ctx, "original prompt", "bad output", errors, 1)
	assert.Contains(t, prompt, "original prompt")
	assert.Contains(t, prompt, "bad json")
}

func TestFocusedStrategy_Name(t *testing.T) {
	s := &FocusedStrategy{}
	assert.Equal(t, "focused", s.Name())
}

func TestFocusedStrategy_ShouldRetry(t *testing.T) {
	s := &FocusedStrategy{}
	ctx := context.Background()

	errors := []validation.ValidationError{
		*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0),
	}
	assert.True(t, s.ShouldRetry(ctx, errors, 1, 3))
}

func TestEscalatingStrategy_Name(t *testing.T) {
	s := &EscalatingStrategy{}
	assert.Equal(t, "escalating", s.Name())
}

func TestEscalatingStrategy_ShouldRetry(t *testing.T) {
	s := &EscalatingStrategy{}
	ctx := context.Background()

	errors := []validation.ValidationError{
		*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0),
	}
	assert.True(t, s.ShouldRetry(ctx, errors, 1, 3))
}

func TestHelpers_HasRetriableErrors(t *testing.T) {
	t.Run("empty errors returns false", func(t *testing.T) {
		assert.False(t, hasRetriableErrors(nil))
	})

	t.Run("retriable error returns true", func(t *testing.T) {
		errors := []validation.ValidationError{
			*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0),
		}
		assert.True(t, hasRetriableErrors(errors))
	})
}

func TestHelpers_CountCriticalErrors(t *testing.T) {
	t.Run("empty errors returns 0", func(t *testing.T) {
		assert.Equal(t, 0, countCriticalErrors(nil))
	})

	t.Run("counts non-warning errors", func(t *testing.T) {
		errors := []validation.ValidationError{
			*validation.NewValidationError(validation.ErrInvalidJSON, "bad json", 0),
			*validation.NewValidationError(validation.ErrInvalidJSON, "more bad json", 0),
		}
		assert.Equal(t, 2, countCriticalErrors(errors))
	})
}

func TestHelpers_FormatCategories(t *testing.T) {
	t.Run("empty returns all", func(t *testing.T) {
		assert.Equal(t, "all", formatCategories(nil))
	})

	t.Run("single category", func(t *testing.T) {
		cats := []validation.ErrorCategory{validation.CategoryStructure}
		result := formatCategories(cats)
		assert.Contains(t, result, string(validation.CategoryStructure))
	})
}
