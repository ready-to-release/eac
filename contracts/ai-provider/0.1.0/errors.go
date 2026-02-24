package ai

import "errors"

// ErrProviderNotConfigured is returned when AI provider configuration is missing.
// Commands that depend on AI should check for this error and display the message.
var ErrProviderNotConfigured = errors.New("ai provider not configured, please run eac init --ai-provider to initialize it")

// ErrProviderUnknown is returned when the configured provider is not registered.
var ErrProviderUnknown = errors.New("unknown AI provider")
