package discovery

// ServiceRegistry defines the interface for service registration and discovery
type ServiceRegistry interface {
	// Register registers a service with the registry
	Register(reg ServiceRegistration) error

	// Deregister removes a service from the registry
	Deregister(serviceID string) error

	// GetService returns the address of a service by name
	GetService(name string) (string, error)

	// GetAllServices returns all instances of a service by name
	GetAllServices(name string) ([]string, error)

	// RegisterSelf registers the current service with the registry
	RegisterSelf(serviceName string) error

	// DeregisterSelf deregisters the current service from the registry
	DeregisterSelf(serviceName string) error
}

// ServiceDiscovery provides a factory for creating service registries
type ServiceDiscovery struct {
	registry ServiceRegistry
}

// NewServiceDiscovery creates a new ServiceDiscovery with the given registry
func NewServiceDiscovery(registry ServiceRegistry) *ServiceDiscovery {
	return &ServiceDiscovery{
		registry: registry,
	}
}

// GetServiceRegistry returns the underlying service registry
func (sd *ServiceDiscovery) GetServiceRegistry() ServiceRegistry {
	return sd.registry
}

// RegisterService registers a service with the registry
func (sd *ServiceDiscovery) RegisterService(reg ServiceRegistration) error {
	return sd.registry.Register(reg)
}

// DeregisterService removes a service from the registry
func (sd *ServiceDiscovery) DeregisterService(serviceID string) error {
	return sd.registry.Deregister(serviceID)
}

// GetServiceAddress returns the address of a service by name
func (sd *ServiceDiscovery) GetServiceAddress(name string) (string, error) {
	return sd.registry.GetService(name)
}

// GetAllServiceAddresses returns all instances of a service by name
func (sd *ServiceDiscovery) GetAllServiceAddresses(name string) ([]string, error) {
	return sd.registry.GetAllServices(name)
}

// RegisterSelf registers the current service with the registry
func (sd *ServiceDiscovery) RegisterSelf(serviceName string) error {
	return sd.registry.RegisterSelf(serviceName)
}

// DeregisterSelf deregisters the current service from the registry
func (sd *ServiceDiscovery) DeregisterSelf(serviceName string) error {
	return sd.registry.DeregisterSelf(serviceName)
}
