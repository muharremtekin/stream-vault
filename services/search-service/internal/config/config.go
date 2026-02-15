package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Consul        ConsulConfig        `mapstructure:"consul"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Redis         RedisConfig         `mapstructure:"redis"`
	RabbitMQ      RabbitMQConfig      `mapstructure:"rabbitmq"`
	Logging       LoggingConfig       `mapstructure:"logging"`
	OTel          OTelConfig          `mapstructure:"otel"`
}

type ServerConfig struct {
	HTTPPort     int           `mapstructure:"http_port"`
	GRPCPort     int           `mapstructure:"grpc_port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type ConsulConfig struct {
	Address             string        `mapstructure:"address"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
}

type ElasticsearchConfig struct {
	URL       string `mapstructure:"url"`
	IndexName string `mapstructure:"index_name"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type RabbitMQConfig struct {
	URL          string `mapstructure:"url"`
	CatalogQueue string `mapstructure:"catalog_queue"`
	WatchQueue   string `mapstructure:"watch_queue"`
	Prefetch     int    `mapstructure:"prefetch"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type OTelConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
	Insecure    bool   `mapstructure:"insecure"`
}

// Load reads configuration from config.yaml and environment variables.
// Environment variables are prefixed with SEARCH_ and use underscores as
// separators (e.g., SEARCH_SERVER_HTTP_PORT=5005).
func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.http_port", 5005)
	v.SetDefault("server.grpc_port", 50053)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("consul.address", "localhost:8500")
	v.SetDefault("consul.health_check_interval", 10*time.Second)
	v.SetDefault("elasticsearch.url", "http://localhost:9200")
	v.SetDefault("elasticsearch.index_name", "streamvault-content")
	v.SetDefault("redis.url", "redis://localhost:6379/3")
	v.SetDefault("rabbitmq.url", "amqp://localhost:5672/")
	v.SetDefault("rabbitmq.catalog_queue", "search.catalog-sync")
	v.SetDefault("rabbitmq.watch_queue", "search.watch-count")
	v.SetDefault("rabbitmq.prefetch", 10)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("otel.enabled", true)
	v.SetDefault("otel.endpoint", "jaeger:4317")
	v.SetDefault("otel.service_name", "search-service")
	v.SetDefault("otel.insecure", true)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/search")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("SEARCH")
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

func validate(cfg *Config) error {
	if cfg.Server.HTTPPort < 1 || cfg.Server.HTTPPort > 65535 {
		return fmt.Errorf("server.http_port must be between 1 and 65535, got %d", cfg.Server.HTTPPort)
	}
	if cfg.Server.GRPCPort < 1 || cfg.Server.GRPCPort > 65535 {
		return fmt.Errorf("server.grpc_port must be between 1 and 65535, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Elasticsearch.URL == "" {
		return fmt.Errorf("elasticsearch.url must not be empty")
	}
	if cfg.Elasticsearch.IndexName == "" {
		return fmt.Errorf("elasticsearch.index_name must not be empty")
	}
	if cfg.Redis.URL == "" {
		return fmt.Errorf("redis.url must not be empty")
	}
	return nil
}
