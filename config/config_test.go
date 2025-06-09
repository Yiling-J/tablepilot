package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConfig_New(t *testing.T) {
	viper.Reset()
	cfg, err := NewConfig("test.toml")
	require.NoError(t, err)
	clientA, ok := cfg.Providers[0].(*OpenAI)
	require.True(t, ok)
	require.Equal(t, &OpenAI{
		Name:    "gemini",
		Type:    "openai",
		Key:     "a",
		BaseURL: "urla",
	}, clientA)
	clientB, ok := cfg.Providers[1].(*OpenAI)
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
		Provider:  "gemini",
		RPM:       10,
	}, cfg.Models[0])
	require.Equal(t, Model{
		Model:    "gpt-4o",
		Alias:    "gpt4o",
		Provider: "oai",
		RPM:      5,
	}, cfg.Models[1])
}

func TestConfig_NewDefault(t *testing.T) {
	viper.Reset()
	cfg, err := NewConfig("")
	require.NoError(t, err)
	require.Equal(t, &Config{
		Common: Common{SourceDataDir: "./"},
		Server: Server{Address: ":8083"},
		Database: &Database{
			Driver: "sqlite3",
			DSN:    "data.db?_pragma=foreign_keys(1)",
		},
	}, cfg)
}
