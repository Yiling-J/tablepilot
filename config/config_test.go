package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_New(t *testing.T) {
	cfg, err := NewConfig("test.toml")
	require.NoError(t, err)
	clientA, ok := cfg.Clients[0].(*OpenAI)
	require.True(t, ok)
	require.Equal(t, &OpenAI{
		Name:    "gemini",
		Type:    "openai",
		Key:     "a",
		BaseURL: "urla",
	}, clientA)
	clientB, ok := cfg.Clients[1].(*OpenAI)
	require.True(t, ok)
	require.Equal(t, &OpenAI{
		Name:    "oai",
		Type:    "openai",
		Key:     "b",
		BaseURL: "urlb",
	}, clientB)
	require.Equal(t, Model{
		Default:   true,
		Model:     "gemini-2.0-flash-001",
		MaxTokens: 1200,
		Client:    "gemini",
		RPM:       10,
	}, cfg.Models[0])
	require.Equal(t, Model{
		Model:  "gpt-4o",
		Alias:  "gpt4o",
		Client: "oai",
		RPM:    5,
	}, cfg.Models[1])
}
