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
	ID        string
	Name      string
	Address   string
	Port      uint16
	Protocol  string
	Namespace string
	Tags      []string
	Meta      map[string]string
}

func NewEntry(id, name, protocol, address string, port uint16) *DefaultService {
	return &DefaultService{
		ID:       id,
		Name:     name,
		Address:  address,
		Port:     port,
		Protocol: protocol,
	}
}

// DialURL implements qim.ServiceRegistration
func (s *DefaultService) DialURL() string {
	if s.Protocol == "tcp" {
		return fmt.Sprintf("%s:%d", s.Address, s.Port)
	}
	return fmt.Sprintf("%s://%s:%d", s.Protocol, s.Address, s.Port)
}

// GetNamespace implements qim.ServiceRegistration
func (s *DefaultService) GetNamespace() string {
	return s.Namespace
}

// GetProtocol implements qim.ServiceRegistration
func (s *DefaultService) GetProtocol() string {
	return s.Protocol
}

// GetTags implements qim.ServiceRegistration
func (s *DefaultService) GetTags() []string {
	return s.Tags
}

// PublicAddress implements qim.ServiceRegistration
func (s *DefaultService) PublicAddress() string {
	return s.Address
}

// PublicPort implements qim.ServiceRegistration
func (s *DefaultService) PublicPort() uint16 {
	return s.Port
}

// String implements qim.ServiceRegistration
func (s *DefaultService) String() string {
	return fmt.Sprintf("ID: %s, Name: %s, Address: %s, Port: %d, Ns: %s, Tags: %v, Meta: %v",
		s.ID, s.Name, s.Address, s.Port, s.Namespace, s.Tags, s.Meta)
}

// GetMeta implements qim.Service
func (s *DefaultService) GetMeta() map[string]string {
	return s.Meta
}

// ServiceID implements qim.Service
func (s *DefaultService) ServiceID() string {
	return s.ID
}

// ServiceName implements qim.Service
func (s *DefaultService) ServiceName() string {
	return s.Name
}

var _ qim.Service = (*DefaultService)(nil)
var _ qim.ServiceRegistration = (*DefaultService)(nil)
