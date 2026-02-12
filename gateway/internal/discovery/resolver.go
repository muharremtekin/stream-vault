package discovery

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/streamvault/gateway/internal/config"
)

// Resolver translates logical service names into concrete network addresses.
// It first checks static configuration, then falls back to Consul-based
// service discovery. When multiple instances are available it uses round-robin
// selection to distribute requests.
type Resolver struct {
	consul   *ConsulClient
	static   map[string]string // service name -> static address from config
	counters sync.Map          // service name -> *uint64 (round-robin counter)
}

// NewResolver creates a Resolver. The consul parameter may be nil if
// service discovery is not available, in which case only static entries
// from the config are used.
func NewResolver(consulClient *ConsulClient, services map[string]config.ServiceEntry) *Resolver {
	static := make(map[string]string, len(services))
	for name, entry := range services {
		if entry.URL != "" {
			static[name] = entry.URL
		}
	}

	return &Resolver{
		consul: consulClient,
		static: static,
	}
}

// Resolve returns a single address for the named service. If a static
// address is configured it is returned directly. Otherwise Consul is
// queried and one of the healthy instances is selected via round-robin.
func (r *Resolver) Resolve(serviceName string) (string, error) {
	// Static configuration takes precedence.
	if addr, ok := r.static[serviceName]; ok {
		return addr, nil
	}

	// Fall back to Consul discovery.
	if r.consul == nil {
		return "", fmt.Errorf("no address configured for service %q and consul is not available", serviceName)
	}

	addresses, err := r.consul.GetHealthyInstances(serviceName)
	if err != nil {
		return "", fmt.Errorf("resolving service %q: %w", serviceName, err)
	}

	addr := r.roundRobin(serviceName, addresses)
	log.Debug().Str("service", serviceName).Str("address", addr).Msg("resolved service address")
	return addr, nil
}

// roundRobin selects the next address from the list using an atomic counter
// per service name.
func (r *Resolver) roundRobin(serviceName string, addresses []string) string {
	if len(addresses) == 1 {
		return addresses[0]
	}

	counterVal, _ := r.counters.LoadOrStore(serviceName, new(uint64))
	counter := counterVal.(*uint64)

	idx := atomic.AddUint64(counter, 1)
	return addresses[idx%uint64(len(addresses))]
}
