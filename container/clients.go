package container

import (
	"sync"

	"github.com/joeyscat/qim"
	"go.uber.org/zap"
)

type ClientMap interface {
	Add(client qim.Client)
	Remove(id string)
	Get(id string) (qim.Client, bool)
	Services(kvs ...string) []qim.Service
}

type ClientsImpl struct {
	clients *sync.Map
}

// Add implements ClientMap
func (ch *ClientsImpl) Add(client qim.Client) {
	if client.ServiceID() == "" {
		c.lg.Error("client id is required", zap.String("module", "ClientsImpl"))
	}

	ch.clients.Store(client.ServiceID(), client)
}

// Get implements ClientMap
func (ch *ClientsImpl) Get(id string) (qim.Client, bool) {
	if id == "" {
		c.lg.Error("client id is required", zap.String("module", "ClientsImpl"))
	}

	val, ok := ch.clients.Load(id)
	if !ok {
		return nil, false
	}
	return val.(qim.Client), true
}

// Remove implements ClientMap
func (ch *ClientsImpl) Remove(id string) {
	ch.clients.Delete(id)
}

// Services implements ClientMap
func (ch *ClientsImpl) Services(kvs ...string) []qim.Service {
	kvLen := len(kvs)
	if kvLen != 0 && kvLen != 2 {
		return nil
	}
	arr := make([]qim.Service, 0)
	ch.clients.Range(func(key, value any) bool {
		svc := value.(qim.Service)
		if kvLen > 0 && svc.GetMeta()[kvs[0]] != kvs[1] {
			return true
		}
		arr = append(arr, svc)
		return true
	})

	return arr
}

func NewClients(num int) ClientMap {
	return &ClientsImpl{
		clients: new(sync.Map),
	}
}
