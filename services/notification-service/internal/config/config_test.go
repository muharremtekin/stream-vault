package config

import (
	"strings"
	"testing"
)

func validConfig() *Config {
	return &Config{
		Server:  ServerConfig{HTTPPort: 5008, GRPCPort: 50055},
		MongoDB: MongoDBConfig{URI: "mongodb://localhost:27017", Database: "streamvault_notifications"},
		JWT:     JWTConfig{Secret: strings.Repeat("x", MinJWTSecretBytes)},
	}
}

func TestValidateRejectsEmptyJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = " "

	if err := validate(cfg); err == nil {
		t.Fatal("expected empty JWT secret to fail validation")
	}
}

func TestValidateRejectsShortJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = strings.Repeat("x", MinJWTSecretBytes-1)

	if err := validate(cfg); err == nil {
		t.Fatal("expected short JWT secret to fail validation")
	}
}

func TestValidateAcceptsJWTSecretAtMinimumLength(t *testing.T) {
	if err := validate(validConfig()); err != nil {
		t.Fatalf("expected valid JWT secret: %v", err)
	}
}
