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
	MinIO    MinIOConfig    `mapstructure:"minio"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Upload   UploadConfig   `mapstructure:"upload"`
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

type MinIOConfig struct {
	Endpoint         string `mapstructure:"endpoint"`
	AccessKey        string `mapstructure:"access_key"`
	SecretKey        string `mapstructure:"secret_key"`
	UseSSL           bool   `mapstructure:"use_ssl"`
	RawBucket        string `mapstructure:"raw_bucket"`
	EncodedBucket    string `mapstructure:"encoded_bucket"`
	ThumbnailsBucket string `mapstructure:"thumbnails_bucket"`
}

type RedisConfig struct {
	URL           string        `mapstructure:"url"`
	ProgressTTL   time.Duration `mapstructure:"progress_ttl"`
	ConcurrentTTL time.Duration `mapstructure:"concurrent_ttl"`
}

type RabbitMQConfig struct {
	URL              string `mapstructure:"url"`
	Exchange         string `mapstructure:"exchange"`
	PublishRoutingKey string `mapstructure:"publish_routing_key"`
	ResultQueue      string `mapstructure:"result_queue"`
	Prefetch         int    `mapstructure:"prefetch"`
	WatchExchange    string `mapstructure:"watch_events_exchange"`
	WatchRoutingKey  string `mapstructure:"watch_events_routing_key"`
}

type UploadConfig struct {
	MaxFileSize  int64    `mapstructure:"max_file_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.http_port", 5003)
	v.SetDefault("server.grpc_port", 50051)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 60*time.Second)
	v.SetDefault("consul.address", "localhost:8500")
	v.SetDefault("consul.health_check_interval", 10*time.Second)
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "")
	v.SetDefault("minio.secret_key", "")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("minio.raw_bucket", "streamvault-raw")
	v.SetDefault("minio.encoded_bucket", "streamvault-encoded")
	v.SetDefault("minio.thumbnails_bucket", "streamvault-thumbnails")
	v.SetDefault("redis.url", "redis://localhost:6379/0")
	v.SetDefault("redis.progress_ttl", 2160*time.Hour)
	v.SetDefault("redis.concurrent_ttl", 5*time.Minute)
	v.SetDefault("rabbitmq.url", "amqp://localhost:5672/")
	v.SetDefault("rabbitmq.exchange", "encoding")
	v.SetDefault("rabbitmq.publish_routing_key", "job.new")
	v.SetDefault("rabbitmq.result_queue", "encoding.results.streaming")
	v.SetDefault("rabbitmq.prefetch", 5)
	v.SetDefault("rabbitmq.watch_events_exchange", "watch.events")
	v.SetDefault("rabbitmq.watch_events_routing_key", "watch.completed")
	v.SetDefault("upload.max_file_size", 10737418240)
	v.SetDefault("upload.allowed_types", []string{"video/mp4", "video/quicktime", "video/x-msvideo", "video/x-matroska"})
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/streaming")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("STREAMING")
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
	if cfg.MinIO.Endpoint == "" {
		return fmt.Errorf("minio.endpoint must not be empty")
	}
	if cfg.Redis.URL == "" {
		return fmt.Errorf("redis.url must not be empty")
	}
	return nil
}
