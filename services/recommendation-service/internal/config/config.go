package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Consul   ConsulConfig   `mapstructure:"consul"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Logging  LoggingConfig  `mapstructure:"logging"`
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

type PostgresConfig struct {
	URL           string `mapstructure:"url"`
	MaxConns      int32  `mapstructure:"max_conns"`
	MinConns      int32  `mapstructure:"min_conns"`
	RunMigrations bool   `mapstructure:"run_migrations"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type RabbitMQConfig struct {
	URL          string `mapstructure:"url"`
	CatalogQueue string `mapstructure:"catalog_queue"`
	WatchQueue   string `mapstructure:"watch_queue"`
	RatingsQueue string `mapstructure:"ratings_queue"`
	Prefetch     int    `mapstructure:"prefetch"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from config.yaml and environment variables.
// Environment variables are prefixed with RECOMMENDATION_ and use underscores
// as separators (e.g., RECOMMENDATION_SERVER_HTTP_PORT=5006).
func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.http_port", 5006)
	v.SetDefault("server.grpc_port", 50054)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("consul.address", "localhost:8500")
	v.SetDefault("consul.health_check_interval", 10*time.Second)
	v.SetDefault("postgres.url", "postgres://streamvault:password@localhost:5432/streamvault_recommendations")
	v.SetDefault("postgres.max_conns", 10)
	v.SetDefault("postgres.min_conns", 2)
	v.SetDefault("postgres.run_migrations", true)
	v.SetDefault("redis.url", "redis://localhost:6379/4")
	v.SetDefault("rabbitmq.url", "amqp://localhost:5672/")
	v.SetDefault("rabbitmq.catalog_queue", "recommendation.catalog")
	v.SetDefault("rabbitmq.watch_queue", "recommendation.watch")
	v.SetDefault("rabbitmq.ratings_queue", "recommendation.ratings")
	v.SetDefault("rabbitmq.prefetch", 10)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/recommendation")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("RECOMMENDATION")
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
	if cfg.Postgres.URL == "" {
		return fmt.Errorf("postgres.url must not be empty")
	}
	if cfg.Redis.URL == "" {
		return fmt.Errorf("redis.url must not be empty")
	}
	return nil
}
