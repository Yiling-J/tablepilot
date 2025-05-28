package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Common struct {
	SourceDataDir string `mapstructure:"source_data_dir"`
}

type Database struct {
	Driver string
	DSN    string
}

type Server struct {
	Address string
}

type Model struct {
	Default   bool
	Model     string
	MaxTokens int64 `mapstructure:"max_tokens"`
	Alias     string
	Provider  string
	RPM       int
	Image     bool
}

type Provider any

type BasicProvider struct {
	Name string
	Type string
}

type OpenAI struct {
	Name    string
	Type    string
	Key     string
	BaseURL string `mapstructure:"base_url"`
}

type Gemini struct {
	Name string
	Type string
	Key  string
}

type Config struct {
	Common    Common
	Server    Server
	Database  *Database
	Models    []Model
	Providers []Provider
	Sources   []map[string]any
}

func NewConfig(name string) (config *Config, err error) {
	viper.SetConfigFile(name)
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return config, err
	}
	var bc []BasicProvider
	err = viper.UnmarshalKey("providers", &bc)
	if err != nil {
		return config, err
	}
	var providers []Provider
	for i, client := range bc {
		key := fmt.Sprintf("providers.%d", i)
		switch client.Type {
		case "gemini", "Gemini":
			var genai Gemini
			err = viper.UnmarshalKey(key, &genai)
			if err != nil {
				return config, err
			}
			providers = append(providers, &genai)
		default:
			var oai OpenAI
			err = viper.UnmarshalKey(key, &oai)
			if err != nil {
				return config, err
			}
			providers = append(providers, &oai)
		}
	}
	config = &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	config.Providers = providers
	if config.Server.Address == "" {
		config.Server.Address = ":8080"
	}
	if config.Common.SourceDataDir == "" {
		config.Common.SourceDataDir = "./"
	}
	return config, nil
}
