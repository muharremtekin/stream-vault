package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOverridesStaticServiceURLsFromEnvironment(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	configContents := []byte(`
jwt:
  secret: test-secret
services:
  user-service:
    url: http://user-service:8080
  catalog-service:
    url: http://catalog-service:5100
  subscription-service:
    url: http://subscription-service:5007
`)
	if err := os.WriteFile(configPath, configContents, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("GATEWAY_USER_SERVICE_URL", "http://user-service:18080")
	t.Setenv("GATEWAY_CATALOG_SERVICE_URL", "http://catalog-service:15100")
	t.Setenv("GATEWAY_SUBSCRIPTION_SERVICE_URL", "http://subscription-service:15007")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	tests := map[string]string{
		"user-service":         "http://user-service:18080",
		"catalog-service":      "http://catalog-service:15100",
		"subscription-service": "http://subscription-service:15007",
	}
	for serviceName, expectedURL := range tests {
		if actualURL := cfg.Services[serviceName].URL; actualURL != expectedURL {
			t.Errorf("%s URL = %q, want %q", serviceName, actualURL, expectedURL)
		}
	}
}
