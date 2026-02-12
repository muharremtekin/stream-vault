package discovery

import (
	"fmt"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

// ConsulClient wraps the HashiCorp Consul API client and provides methods
// for discovering healthy service instances.
type ConsulClient struct {
	client *consul.Client

	mu    sync.RWMutex
	cache map[string]*serviceCache
}

// serviceCache stores the last known healthy instances and the time they
// were fetched, allowing short-lived caching to reduce Consul load.
type serviceCache struct {
	addresses []string
	fetchedAt time.Time
	ttl       time.Duration
}

// NewConsulClient creates a new Consul client connected to the given address.
func NewConsulClient(address string) (*ConsulClient, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = address

	client, err := consul.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating consul client: %w", err)
	}

	// Verify connectivity.
	_, err = client.Agent().Self()
	if err != nil {
		log.Warn().Err(err).Str("address", address).Msg("consul agent not reachable, continuing without discovery")
	}

	return &ConsulClient{
		client: client,
		cache:  make(map[string]*serviceCache),
	}, nil
}

// GetHealthyInstances returns the addresses (host:port) of all healthy
// instances of the named service. Results are cached for a short duration.
func (cc *ConsulClient) GetHealthyInstances(serviceName string) ([]string, error) {
	cc.mu.RLock()
	if cached, ok := cc.cache[serviceName]; ok {
		if time.Since(cached.fetchedAt) < cached.ttl {
			addrs := cached.addresses
			cc.mu.RUnlock()
			return addrs, nil
		}
	}
	cc.mu.RUnlock()

	entries, _, err := cc.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("querying consul for service %q: %w", serviceName, err)
	}

	addresses := make([]string, 0, len(entries))
	for _, entry := range entries {
		addr := entry.Service.Address
		if addr == "" {
			addr = entry.Node.Address
		}
		port := entry.Service.Port
		addresses = append(addresses, fmt.Sprintf("%s:%d", addr, port))
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service %q", serviceName)
	}

	cc.mu.Lock()
	cc.cache[serviceName] = &serviceCache{
		addresses: addresses,
		fetchedAt: time.Now(),
		ttl:       5 * time.Second,
	}
	cc.mu.Unlock()

	log.Debug().Str("service", serviceName).Int("instances", len(addresses)).Msg("refreshed service instances from consul")

	return addresses, nil
}

// RegisterService registers the gateway itself with Consul so other services
// can discover it. The health check uses the gateway's /health endpoint.
func (cc *ConsulClient) RegisterService(id, name, address string, port int, healthURL string, interval time.Duration) error {
	reg := &consul.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Check: &consul.AgentServiceCheck{
			HTTP:                           healthURL,
			Interval:                       interval.String(),
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := cc.client.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("registering service %q with consul: %w", name, err)
	}

	log.Info().Str("id", id).Str("name", name).Msg("registered with consul")
	return nil
}

// DeregisterService removes the gateway from Consul's service registry.
func (cc *ConsulClient) DeregisterService(id string) error {
	if err := cc.client.Agent().ServiceDeregister(id); err != nil {
		return fmt.Errorf("deregistering service %q from consul: %w", id, err)
	}
	return nil
}
