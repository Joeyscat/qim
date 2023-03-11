package naming

import (
	"errors"

	"github.com/joeyscat/qim"
)

var (
	ErrNotFound = errors.New("service not found")
)

type Naming interface {
	Find(serviceName string, tags ...string) ([]qim.ServiceRegistration, error)
	Subscribe(serviceName string, callback func(services []qim.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(service qim.ServiceRegistration) error
	Deregister(serviceID string) error
}
