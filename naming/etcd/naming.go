package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/naming"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdNaming struct {
	cli *clientv3.Client
	mu  *sync.RWMutex
	// key: serviceID, value: service
	registry map[string]qim.ServiceRegistration
	// key: serviceName, value: callback function
	watchCallback map[string]func(services []qim.ServiceRegistration)
	// key: serviceName, value: cancelFunc for watch
	watchCancelFunc map[string]context.CancelFunc
	lg              *zap.Logger
}

func NewNaming(endpoints []string, lg *zap.Logger) (naming.Naming, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	e := &etcdNaming{
		cli:             cli,
		mu:              &sync.RWMutex{},
		registry:        make(map[string]qim.ServiceRegistration),
		watchCallback:   make(map[string]func([]qim.ServiceRegistration)),
		watchCancelFunc: make(map[string]context.CancelFunc),
		lg:              lg,
	}
	return e, nil
}

var _ naming.Naming = (*etcdNaming)(nil)

// Find implements naming.Naming
func (e *etcdNaming) Find(serviceName string, tags ...string) ([]qim.ServiceRegistration, error) {
	resp, err := e.cli.Get(context.Background(), ServicesPrefix+serviceName+"/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var services []qim.ServiceRegistration
	for _, kv := range resp.Kvs {
		var s naming.DefaultService
		if err := json.Unmarshal(kv.Value, &s); err != nil {
			return nil, err
		}
		if e.matchTags(s.GetTags(), tags) {
			services = append(services, &s)
		}
	}
	return services, nil
}

// Register implements naming.Naming
func (e *etcdNaming) Register(service qim.ServiceRegistration) error {
	// put kv
	// keepalive
	// TODO keep leaseID
	e.lg.Info("register service", zap.String("service", service.String()))
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.registry[service.ServiceID()]; ok {
		return fmt.Errorf("service already registered: %s", service.ServiceID())
	}

	data, err := json.Marshal(service)
	if err != nil {
		return err
	}
	key := ServicesPrefix + service.ServiceName() + "/" + service.ServiceID()
	leaseResp, err := e.cli.Grant(context.Background(), 5)
	if err != nil {
		return err
	}
	_, err = e.cli.Put(context.Background(), key, string(data), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	_, err = e.cli.KeepAlive(context.TODO(), leaseResp.ID)
	if err != nil {
		e.lg.Error("KeepAlive error", zap.Error(err))
		return err
	}

	e.registry[service.ServiceID()] = service
	return nil
}

// Deregister implements naming.Naming
func (e *etcdNaming) Deregister(serviceID string) error {
	// TODO revoke lease
	// delete kv

	e.mu.Lock()
	defer e.mu.Unlock()
	for k, service := range e.registry {
		if k == serviceID {
			delete(e.registry, k)
			key := ServicesPrefix + service.ServiceName() + "/" + service.ServiceID()
			_, err := e.cli.Delete(context.Background(), key)
			return err
		}
	}
	return errors.New("ServiceID  Not  Found")
}

// Subscribe implements naming.Naming
func (e *etcdNaming) Subscribe(serviceName string, callback func(services []qim.ServiceRegistration)) error {
	// cancel old watch
	// create new watch
	// keep the cancelFunc

	e.mu.Lock()
	defer e.mu.Unlock()

	if cancel, ok := e.watchCancelFunc[serviceName]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.watchCancelFunc[serviceName] = cancel

	go func() {
		watchKey := ServicesPrefix + serviceName + "/"
		rch := e.cli.Watch(ctx, watchKey, clientv3.WithPrefix())
		for wresp := range rch {
			for _, ev := range wresp.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					fallthrough
				case clientv3.EventTypePut:
					e.mu.Lock()
					cb, ok := e.watchCallback[serviceName]
					e.mu.Unlock()
					if ok {
						cb(e.getServices(serviceName))
					}
				default:
					e.lg.Info("unknown watch event", zap.Any("event", ev))
				}
			}
		}
		e.lg.Info("subscribe over", zap.String("serviceName", serviceName))
	}()

	e.watchCallback[serviceName] = callback
	callback(e.getServices(serviceName))

	return nil
}

// Unsubscribe implements naming.Naming
func (e *etcdNaming) Unsubscribe(serviceName string) error {
	e.mu.Lock()
	delete(e.watchCallback, serviceName)
	cancel, ok := e.watchCancelFunc[serviceName]
	e.mu.Unlock()

	if ok {
		cancel()
	}

	return nil
}

const (
	ServicesPrefix = "services/"
)

func (e *etcdNaming) getServices(serviceName string) []qim.ServiceRegistration {
	var services []qim.ServiceRegistration
	for _, v := range e.registry {
		if v.ServiceName() == serviceName {
			services = append(services, v)
		}
	}
	return services
}

func (e *etcdNaming) matchTags(srcTags, matchTags []string) bool {
	if len(matchTags) == 0 {
		return true
	}
	for _, mt := range matchTags {
		if !slices.Contains(srcTags, mt) {
			return false
		}
	}
	return true
}
