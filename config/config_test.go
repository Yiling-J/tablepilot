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
}
