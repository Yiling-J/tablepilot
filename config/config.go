package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Database struct {
	Driver string
	DSN    string
}

type Model struct {
	Type    string
	Default bool
	Model   string
	Alias   string
	Client  string
	RPM     int
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
type Gemini struct {
	Name string
	Type string
	Key  string
}

type Config struct {
	Debug    bool
	Database *Database
	Models   []Model
	Clients  []Client
}

func NewConfig(name string) (config *Config, err error) {
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

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
		case "gemini":
			var gai Gemini
			err = viper.UnmarshalKey(key, &gai)
			if err != nil {
				return config, err
			}
			clients = append(clients, &gai)
		default:
			return nil, errors.New("unknow client")
		}
	}
	config = &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	config.Clients = clients
	return config, nil
}
