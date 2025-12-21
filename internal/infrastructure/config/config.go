package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Default values
const (
	DefaultSchemaCacheTTL = 1 * time.Hour
)

type Config struct {
	Google GoogleConfig
	Server ServerConfig
	Agent  AgentConfig
}

type GoogleConfig struct {
	APIKey string
}

type ServerConfig struct {
	Port      string
	UIBaseURL string
}

type AgentConfig struct {
	SchemaCacheTTL time.Duration
}

func Load() (config Config, err error) {
	viper.AutomaticEnv()

	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, err
	}

	// Google
	config.Google.APIKey = viper.GetString("GOOGLE_API_KEY")

	// Server
	config.Server.Port = viper.GetString("SERVER_PORT")
	config.Server.UIBaseURL = viper.GetString("SERVER_UIBASEURL")

	// Agent
	config.Agent.SchemaCacheTTL = parseDuration(
		viper.GetString("SCHEMA_CACHE_TTL"),
		DefaultSchemaCacheTTL,
	)

	jss, _ := json.Marshal(config)
	fmt.Println(string(jss))

	return config, nil
}

// parseDuration parses a duration string, returns default if invalid or empty
func parseDuration(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}
