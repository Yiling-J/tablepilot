package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

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
	Client    string
	RPM       int
}

type Client interface{}

type BasicClient struct {
	Name string
	Type string
}

type OpenAI struct {
	Name    string
	Type    string
	Key     string
	BaseURL string `mapstructure:"base_url"`
}

type Config struct {
	Server   Server
	Database *Database
	Models   []Model
	Clients  []Client
}

func NewConfig(name string) (config *Config, err error) {
	viper.SetConfigFile(name)
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return config, err
	}
	var bc []BasicClient
	err = viper.UnmarshalKey("clients", &bc)
	if err != nil {
		return config, err
	}
	var clients []Client
	for i, client := range bc {
		key := fmt.Sprintf("clients.%d", i)
		switch client.Type {
		case "openai":
			var oai OpenAI
			err = viper.UnmarshalKey(key, &oai)
			if err != nil {
				return config, err
			}
			clients = append(clients, &oai)
		default:
			return nil, errors.New("unknown client")
		}
	}
	config = &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	config.Clients = clients
	if config.Server.Address == "" {
		config.Server.Address = ":8080"
	}
	return config, nil
}
