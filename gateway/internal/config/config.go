package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the complete gateway configuration.
type Config struct {
	Server     ServerConfig            `mapstructure:"server"`
	JWT        JWTConfig               `mapstructure:"jwt"`
	RateLimit  RateLimitConfig         `mapstructure:"ratelimit"`
	Consul     ConsulConfig            `mapstructure:"consul"`
	Services   map[string]ServiceEntry `mapstructure:"services"`
	Logging    LoggingConfig           `mapstructure:"logging"`
	OTel       OTelConfig              `mapstructure:"otel"`
	Resilience ResilienceConfig        `mapstructure:"resilience"`
}

// OTelConfig holds OpenTelemetry tracing settings.
type OTelConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
	Insecure    bool   `mapstructure:"insecure"`
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

// ResilienceConfig holds circuit breaker, bulkhead, timeout, and retry settings.
type ResilienceConfig struct {
	CircuitBreaker       CBConfig                           `mapstructure:"circuit_breaker"`
	Retry                RetryConfig                        `mapstructure:"retry"`
	DefaultTimeout       time.Duration                      `mapstructure:"default_timeout"`
	DefaultMaxConcurrent int                                `mapstructure:"default_max_concurrent"`
	Services             map[string]ServiceResilienceConfig `mapstructure:"services"`
}

// RetryConfig holds upstream HTTP retry policy settings.
type RetryConfig struct {
	MaxRetries     int           `mapstructure:"max_retries"`
	InitialBackoff time.Duration `mapstructure:"initial_backoff"`
	MaxBackoff     time.Duration `mapstructure:"max_backoff"`
	Multiplier     float64       `mapstructure:"multiplier"`
	JitterFraction float64       `mapstructure:"jitter_fraction"`
}

// CBConfig holds circuit breaker settings shared across all services.
type CBConfig struct {
	FailureThreshold uint32        `mapstructure:"failure_threshold"`
	SuccessThreshold uint32        `mapstructure:"success_threshold"`
	Timeout          time.Duration `mapstructure:"timeout"`
}

// ServiceResilienceConfig holds per-service resilience overrides.
type ServiceResilienceConfig struct {
	Timeout       time.Duration `mapstructure:"timeout"`
	MaxConcurrent int           `mapstructure:"max_concurrent"`
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
	v.SetDefault("otel.enabled", true)
	v.SetDefault("otel.endpoint", "jaeger:4317")
	v.SetDefault("otel.service_name", "gateway")
	v.SetDefault("otel.insecure", true)
	v.SetDefault("resilience.circuit_breaker.failure_threshold", 5)
	v.SetDefault("resilience.circuit_breaker.success_threshold", 3)
	v.SetDefault("resilience.circuit_breaker.timeout", 30*time.Second)
	v.SetDefault("resilience.retry.max_retries", 3)
	v.SetDefault("resilience.retry.initial_backoff", 100*time.Millisecond)
	v.SetDefault("resilience.retry.max_backoff", 2*time.Second)
	v.SetDefault("resilience.retry.multiplier", 2.0)
	v.SetDefault("resilience.retry.jitter_fraction", 0.2)
	v.SetDefault("resilience.default_timeout", 5*time.Second)
	v.SetDefault("resilience.default_max_concurrent", 30)

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
	serviceURLBindings := map[string]string{
		"services.user-service.url":         "GATEWAY_USER_SERVICE_URL",
		"services.catalog-service.url":      "GATEWAY_CATALOG_SERVICE_URL",
		"services.subscription-service.url": "GATEWAY_SUBSCRIPTION_SERVICE_URL",
	}
	for key, environmentVariable := range serviceURLBindings {
		if err := v.BindEnv(key, environmentVariable); err != nil {
			return nil, fmt.Errorf("binding %s: %w", environmentVariable, err)
		}
	}
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
