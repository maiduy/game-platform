package discovery

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulServiceRegistry provides service registration and discovery using Consul
type ConsulServiceRegistry struct {
	client *api.Client
	config *api.Config
}

// ServiceRegistration contains the information needed to register a service
type ServiceRegistration struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
}

// NewConsulServiceRegistry creates a new ConsulServiceRegistry
func NewConsulServiceRegistry(address string) (*ConsulServiceRegistry, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	return &ConsulServiceRegistry{
		client: client,
		config: config,
	}, nil
}

// Register registers a service with Consul
func (c *ConsulServiceRegistry) Register(reg ServiceRegistration) error {
	agent := c.client.Agent()

	// Create service registration
	service := &api.AgentServiceRegistration{
		ID:      reg.ID,
		Name:    reg.Name,
		Address: reg.Address,
		Port:    reg.Port,
		Tags:    reg.Tags,
		Meta:    reg.Meta,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", reg.Address, reg.Port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return agent.ServiceRegister(service)
}

// Deregister removes a service from Consul
func (c *ConsulServiceRegistry) Deregister(serviceID string) error {
	return c.client.Agent().ServiceDeregister(serviceID)
}

// GetService returns the address of a service by name
func (c *ConsulServiceRegistry) GetService(name string) (string, error) {
	// Query for service
	services, _, err := c.client.Health().Service(name, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to query service: %w", err)
	}

	// No healthy services found
	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances of service %s found", name)
	}

	// Return first healthy service
	service := services[0].Service
	return fmt.Sprintf("%s:%d", service.Address, service.Port), nil
}

// GetAllServices returns all instances of a service by name
func (c *ConsulServiceRegistry) GetAllServices(name string) ([]string, error) {
	// Query for service
	services, _, err := c.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query service: %w", err)
	}

	// No healthy services found
	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances of service %s found", name)
	}

	// Return all healthy services
	var addresses []string
	for _, service := range services {
		addresses = append(addresses, fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port))
	}

	return addresses, nil
}

// RegisterSelf registers the current service with Consul
func (c *ConsulServiceRegistry) RegisterSelf(serviceName string) error {
	// Get host information
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get port from environment or use default
	portStr := os.Getenv("GO_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8080 // Default port
	}

	// Get service address from environment or use hostname
	address := os.Getenv("SERVICE_ADDRESS")
	if address == "" {
		address = hostname
	}

	// Create service registration
	reg := ServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", serviceName, hostname, port),
		Name:    serviceName,
		Address: address,
		Port:    port,
		Tags:    []string{"go", "microservice"},
		Meta: map[string]string{
			"version": os.Getenv("SERVICE_VERSION"),
		},
	}

	// Register service
	err = c.Register(reg)
	if err != nil {
		return err
	}

	// Register shutdown hook
	go func() {
		// Wait for interrupt signal
		<-time.After(50 * time.Millisecond)
		log.Printf("Service %s registered with Consul", serviceName)
	}()

	return nil
}

// DeregisterSelf deregisters the current service from Consul
func (c *ConsulServiceRegistry) DeregisterSelf(serviceName string) error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	portStr := os.Getenv("GO_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8080
	}

	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, hostname, port)
	return c.Deregister(serviceID)
}
