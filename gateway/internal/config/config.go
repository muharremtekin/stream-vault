package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the complete gateway configuration.
type Config struct {
	Server    ServerConfig            `mapstructure:"server"`
	JWT       JWTConfig               `mapstructure:"jwt"`
	RateLimit RateLimitConfig         `mapstructure:"ratelimit"`
	Consul    ConsulConfig            `mapstructure:"consul"`
	Services  map[string]ServiceEntry `mapstructure:"services"`
	Logging   LoggingConfig           `mapstructure:"logging"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// JWTConfig holds JWT authentication parameters.
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
}

// RateLimitConfig holds token bucket rate limiter settings.
type RateLimitConfig struct {
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
	RedisURL          string  `mapstructure:"redis_url"`
}

// ConsulConfig holds Consul service discovery settings.
type ConsulConfig struct {
	Address             string        `mapstructure:"address"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
}

// ServiceEntry represents a single upstream service configuration.
type ServiceEntry struct {
	URL         string `mapstructure:"url"`
	HealthPath  string `mapstructure:"health_path"`
	StripPrefix bool   `mapstructure:"strip_prefix"`
}

// LoggingConfig holds structured logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from config.yaml and environment variables.
// Environment variables are prefixed with GATEWAY_ and use underscores as
// separators (e.g., GATEWAY_SERVER_PORT=8080).
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults.
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.issuer", "streamvault")
	v.SetDefault("ratelimit.requests_per_second", 100.0)
	v.SetDefault("ratelimit.burst", 200)
	v.SetDefault("ratelimit.redis_url", "redis://localhost:6379/0")
	v.SetDefault("consul.address", "localhost:8500")
	v.SetDefault("consul.health_check_interval", 10*time.Second)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Read from config file.
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/gateway")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		// Config file not found is acceptable; we rely on defaults and env vars.
	}

	// Read from environment variables.
	v.SetEnvPrefix("GATEWAY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func validate(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", cfg.Server.Port)
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret must not be empty")
	}
	return nil
}
