package config

import (
	"time"

	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/spf13/viper"
)

const (
	DefaultSchemaCacheTTL = 1 * time.Hour
)

type Config struct {
	Server    ServerConfig
	Agent     AgentConfig
	Providers map[domainAgent.Provider]string
}

type ServerConfig struct {
	Port      string
	UIBaseURL string
	Env       string
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

	config.Server.Port = viper.GetString("SERVER_PORT")
	config.Server.UIBaseURL = viper.GetString("SERVER_UIBASEURL")
	config.Server.Env = viper.GetString("SERVER_ENV")
	if config.Server.Env == "" {
		config.Server.Env = "production"
	}

	config.Agent.SchemaCacheTTL = parseDuration(
		viper.GetString("SCHEMA_CACHE_TTL"),
		DefaultSchemaCacheTTL,
	)

	config.Providers = loadProviders()

	return config, nil
}

func loadProviders() map[domainAgent.Provider]string {
	providers := make(map[domainAgent.Provider]string)

	for provider, cfg := range domainAgent.ProviderRegistry {
		if key := viper.GetString(cfg.EnvKey); key != "" {
			providers[provider] = key
		}
	}

	return providers
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
