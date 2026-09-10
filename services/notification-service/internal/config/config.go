package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const MinJWTSecretBytes = 32

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Consul    ConsulConfig    `mapstructure:"consul"`
	MongoDB   MongoDBConfig   `mapstructure:"mongodb"`
	RabbitMQ  RabbitMQConfig  `mapstructure:"rabbitmq"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	OTel      OTelConfig      `mapstructure:"otel"`
}

// OTelConfig holds OpenTelemetry tracing settings.
type OTelConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
	Insecure    bool   `mapstructure:"insecure"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
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

type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type RabbitMQConfig struct {
	URL               string `mapstructure:"url"`
	SubscriptionQueue string `mapstructure:"subscription_queue"`
	EncodingQueue     string `mapstructure:"encoding_queue"`
	ContentQueue      string `mapstructure:"content_queue"`
	Prefetch          int    `mapstructure:"prefetch"`
}

type WebSocketConfig struct {
	ReadBufferSize        int           `mapstructure:"read_buffer_size"`
	WriteBufferSize       int           `mapstructure:"write_buffer_size"`
	PingInterval          time.Duration `mapstructure:"ping_interval"`
	PongTimeout           time.Duration `mapstructure:"pong_timeout"`
	MaxConnectionsPerUser int           `mapstructure:"max_connections_per_user"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.http_port", 5008)
	v.SetDefault("server.grpc_port", 50055)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("consul.address", "localhost:8500")
	v.SetDefault("consul.health_check_interval", 10*time.Second)
	v.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	v.SetDefault("mongodb.database", "streamvault_notifications")
	v.SetDefault("rabbitmq.url", "amqp://localhost:5672/")
	v.SetDefault("rabbitmq.subscription_queue", "notification.subscription")
	v.SetDefault("rabbitmq.encoding_queue", "notification.encoding")
	v.SetDefault("rabbitmq.content_queue", "notification.content")
	v.SetDefault("rabbitmq.prefetch", 10)
	v.SetDefault("websocket.read_buffer_size", 1024)
	v.SetDefault("websocket.write_buffer_size", 1024)
	v.SetDefault("websocket.ping_interval", 30*time.Second)
	v.SetDefault("websocket.pong_timeout", 10*time.Second)
	v.SetDefault("websocket.max_connections_per_user", 5)
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.issuer", "StreamVault.UserService")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("otel.enabled", true)
	v.SetDefault("otel.endpoint", "jaeger:4317")
	v.SetDefault("otel.service_name", "notification-service")
	v.SetDefault("otel.insecure", true)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/notification")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("NOTIFICATION")
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
	if cfg.MongoDB.URI == "" {
		return fmt.Errorf("mongodb.uri must not be empty")
	}
	if cfg.MongoDB.Database == "" {
		return fmt.Errorf("mongodb.database must not be empty")
	}
	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		return fmt.Errorf("jwt.secret must not be empty")
	}
	if len([]byte(cfg.JWT.Secret)) < MinJWTSecretBytes {
		return fmt.Errorf("jwt.secret must be at least %d bytes", MinJWTSecretBytes)
	}
	return nil
}
