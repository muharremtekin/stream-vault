package discovery

import (
	"fmt"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

type ConsulClient struct {
	client    *consul.Client
	serviceID string
}

func NewConsulClient(address string) (*ConsulClient, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = address

	client, err := consul.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating consul client: %w", err)
	}

	_, err = client.Agent().Self()
	if err != nil {
		log.Warn().Err(err).Str("address", address).Msg("consul agent not reachable")
	}

	return &ConsulClient{client: client}, nil
}

func (c *ConsulClient) Register(serviceName string, port int, interval time.Duration) error {
	c.serviceID = serviceName

	reg := &consul.AgentServiceRegistration{
		ID:      serviceName,
		Name:    serviceName,
		Address: serviceName,
		Port:    port,
		Tags:    []string{"streaming", "api", "v1"},
		Check: &consul.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health/ready", serviceName, port),
			Interval:                       interval.String(),
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "60s",
		},
	}

	if err := c.client.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("registering service %q with consul: %w", serviceName, err)
	}

	log.Info().Str("id", serviceName).Msg("registered with consul")
	return nil
}

func (c *ConsulClient) Deregister() error {
	if c.serviceID == "" {
		return nil
	}
	if err := c.client.Agent().ServiceDeregister(c.serviceID); err != nil {
		return fmt.Errorf("deregistering service %q from consul: %w", c.serviceID, err)
	}
	log.Info().Str("id", c.serviceID).Msg("deregistered from consul")
	return nil
}
