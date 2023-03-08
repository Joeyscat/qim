package naming

import (
	"fmt"

	"github.com/joeyscat/qim"
)

// "ID": "qa-dfirst-zfirst-tgateway-172.16.235.145-0-8000",
// "Service": "tgateway",
// "Tags": [
// "ZONE:qa-dfirst-zfirst",
// "TMC_REGION:SH",
// "TMC_DOMAIN:g002-qa.tutormeetplus.com"
// ],
// "Address": "172.16.235.145",
// "Port": 8000,

// implementation of qim.Service and qim.ServiceRegistration
type DefaultService struct {
	id        string
	name      string
	address   string
	port      uint8
	protocol  string
	namespace string
	tags      []string
	meta      map[string]string
}

func NewEntry(id, name, protocol, address string, port uint8) *DefaultService {
	return &DefaultService{
		id:       id,
		name:     name,
		address:  address,
		port:     port,
		protocol: protocol,
	}
}

// DialURL implements qim.ServiceRegistration
func (s *DefaultService) DialURL() string {
	if s.protocol == "tcp" {
		return fmt.Sprintf("%s:%d", s.address, s.port)
	}
	return fmt.Sprintf("%s://%s:%d", s.protocol, s.address, s.port)
}

// GetNamespace implements qim.ServiceRegistration
func (s *DefaultService) GetNamespace() string {
	return s.namespace
}

// GetProtocol implements qim.ServiceRegistration
func (s *DefaultService) GetProtocol() string {
	return s.protocol
}

// GetTags implements qim.ServiceRegistration
func (s *DefaultService) GetTags() []string {
	return s.tags
}

// PublicAddress implements qim.ServiceRegistration
func (s *DefaultService) PublicAddress() string {
	return s.address
}

// PublicPort implements qim.ServiceRegistration
func (s *DefaultService) PublicPort() uint8 {
	return s.port
}

// String implements qim.ServiceRegistration
func (s *DefaultService) String() string {
	return fmt.Sprintf("ID: %s, Name: %s, Address: %s, Port: %d, Ns: %s, Tags: %v, Meta: %v",
		s.id, s.name, s.address, s.port, s.namespace, s.tags, s.meta)
}

// GetMeta implements qim.Service
func (s *DefaultService) GetMeta() map[string]string {
	return s.meta
}

// ServiceID implements qim.Service
func (s *DefaultService) ServiceID() string {
	return s.id
}

// ServiceName implements qim.Service
func (s *DefaultService) ServiceName() string {
	return s.name
}

var _ qim.Service = (*DefaultService)(nil)
var _ qim.ServiceRegistration = (*DefaultService)(nil)
