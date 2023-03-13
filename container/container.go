package container

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/tcp"
	"github.com/joeyscat/qim/wire"
	"github.com/joeyscat/qim/wire/pkt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	stateUninitialized = iota
	stateInitialized
	stateStarted
	stateClosed
)

const (
	StateYoung = "young"
	StateAdult = "adult"
)

const (
	KeyServiceState = "service_state"
)

type Container struct {
	sync.RWMutex
	Naming     naming.Naming
	Srv        qim.Server
	state      uint32
	srvclients map[string]ClientMap
	seletor    Selector
	dialer     qim.Dialer
	deps       map[string]struct{}
	monitor    sync.Once
	lg         *zap.Logger // TODO
}

// Default Container
var c = &Container{
	state:   0,
	seletor: &HashSelector{},
	deps:    map[string]struct{}{},
	lg:      zap.NewExample(),
}

func Default() *Container {
	return c
}

// init container with a Server, and its deps
//
// For example, in the gateway, it depends on the login and chat services,
// and it will cal the function like this:
// _ = container.Init(srv, wire.SNChat, wire.SNLogin)
func Init(srv qim.Server, deps ...string) error {
	if !atomic.CompareAndSwapUint32(&c.state, stateUninitialized, stateInitialized) {
		return errors.New("already initialized")
	}
	c.Srv = srv

	for _, dep := range deps {
		if _, ok := c.deps[dep]; ok {
			continue
		}
		c.deps[dep] = struct{}{}
	}
	var err error
	zapFields := zap.Fields(zap.String("module", "container"))
	if os.Getenv("DEBUG") == "true" {
		c.lg, err = zap.NewDevelopment(zapFields)
	} else {
		c.lg, err = zap.NewProduction(zapFields)
	}
	if err != nil {
		return err
	}

	c.srvclients = make(map[string]ClientMap, len(deps))
	return nil
}

func SetDialer(dialer qim.Dialer) {
	c.dialer = dialer
}

func SetSelector(selector Selector) {
	c.seletor = selector
}

func SetServiceNaming(nm naming.Naming) {
	c.Naming = nm
}

func EnableMonitor(listen string) {
	c.monitor.Do(func() {
		go func() {
			http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("ok"))
			})
			http.Handle("/metrics", promhttp.Handler())
			_ = http.ListenAndServe(listen, nil)
		}()
	})
}

// start server
func Start() error {
	if c.Naming == nil {
		return errors.New("naming is nil")
	}
	if !atomic.CompareAndSwapUint32(&c.state, stateInitialized, stateStarted) {
		return fmt.Errorf("invalid state: %d", c.state)
	}

	go func(srv qim.Server) {
		err := srv.Start()
		if err != nil {
			c.lg.Error(err.Error())
		}
	}(c.Srv)

	// 1.
	for service := range c.deps {
		go func(service string) {
			err := connectToService(service)
			if err != nil {
				c.lg.Error(err.Error())
			}
		}(service)
	}

	// 2.
	if c.Srv.PublicAddress() != "" && c.Srv.PublicPort() != 0 {
		err := c.Naming.Register(c.Srv)
		if err != nil {
			c.lg.Error(err.Error())
		}
	}

	// 3.
	cx := make(chan os.Signal, 1)
	signal.Notify(cx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	c.lg.Info("shutdown", zap.Any("signal", <-cx))
	// 4.
	return shutdown()
}

// push message to server
func Push(server string, p *pkt.LogicPkt) error {
	p.AddStringMeta(wire.MetaDestServer, server)
	return c.Srv.Push(server, pkt.Marshal(p))
}

// forward message to service
func Forward(serviceName string, packet *pkt.LogicPkt) error {
	if packet != nil {
		return errors.New("packet is nil")
	}
	if packet.Command == "" {
		return errors.New("command is empty in packet")
	}
	if packet.ChannelId == "" {
		return errors.New("ChannelId is empty in packet")
	}
	return ForwardWithSelector(serviceName, packet, c.seletor)
}

func ForwardWithSelector(serviceName string, packet *pkt.LogicPkt, selector Selector) error {
	cli, err := lookup(serviceName, &packet.Header, selector)
	if err != nil {
		return err
	}
	// add a tag to packet
	packet.AddStringMeta(wire.MetaDestServer, c.Srv.ServiceID())

	c.lg.Debug("forward message", zap.String("to", cli.ServiceID()), zap.String("header", packet.Header.String()))
	return cli.Send(pkt.Marshal(packet))
}

func lookup(serviceName string, header *pkt.Header, selector Selector) (qim.Client, error) {
	clients, ok := c.srvclients[serviceName]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	srvs := clients.Services(KeyServiceState, StateAdult)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("no services found for %s", serviceName)
	}
	id := selector.Lookup(header, srvs)
	if cli, ok := clients.Get(id); ok {
		return cli, nil
	}
	return nil, errors.New("no client found")
}

func shutdown() error {
	if !atomic.CompareAndSwapUint32(&c.state, stateStarted, stateClosed) {
		return fmt.Errorf("invalid state: %d", c.state)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	err := c.Srv.Shutdown(ctx)
	if err != nil {
		c.lg.Error(err.Error())
	}

	err = c.Naming.Deregister(c.Srv.ServiceID())
	if err != nil {
		c.lg.Warn(err.Error())
	}

	for dep := range c.deps {
		_ = c.Naming.Unsubscribe(dep)
	}

	c.lg.Info("shutdown")
	return nil
}

func connectToService(serviceName string) error {
	log := c.lg.With(zap.String("func", "connectToService"))
	clients := NewClients(10)
	c.srvclients[serviceName] = clients
	// 1. Watch for new services
	delay := time.Second * 10
	err := c.Naming.Subscribe(serviceName, func(services []qim.ServiceRegistration) {
		for _, service := range services {
			if _, ok := clients.Get(service.ServiceID()); ok {
				continue
			}
			log.Info("Watch for a new service", zap.Any("service", service))
			service.GetMeta()[KeyServiceState] = StateYoung

			go func(service qim.ServiceRegistration) {
				time.Sleep(delay)
				service.GetMeta()[KeyServiceState] = StateAdult
			}(service)

			_, err := buildClient(clients, service)
			if err != nil {
				log.Warn(err.Error())
			}
		}
	})
	if err != nil {
		return err
	}
	// 2. get online services
	services, err := c.Naming.Find(serviceName)
	if err != nil {
		return err
	}
	log.Info("find service", zap.Any("services", services))
	for _, service := range services {
		// change service state to adult
		service.GetMeta()[KeyServiceState] = StateAdult
		_, err := buildClient(clients, service)
		if err != nil {
			log.Warn(err.Error())
		}
	}
	return nil
}

func buildClient(clients ClientMap, service qim.ServiceRegistration) (qim.Client, error) {
	c.Lock()
	defer c.Unlock()
	var (
		id   = service.ServiceID()
		name = service.ServiceName()
		meta = service.GetMeta()
	)
	// 1. return if client already exists
	if _, ok := clients.Get(id); ok {
		return nil, nil
	}
	// 2. use only tcp between services
	if service.GetProtocol() != string(wire.ProtocolTCP) {
		return nil, fmt.Errorf("unexpected service protocol: %s", service.GetProtocol())
	}
	// 3. create client and connect
	cli := tcp.NewClientWithProps(id, name, meta, tcp.ClientOptions{
		Heartbeat: qim.DefaultHeartbeat,
		Readwait:  qim.DefaultReadwait,
		Writewait: qim.DefaultWritewait,
	})
	if c.dialer == nil {
		return nil, errors.New("dialer is nil")
	}
	cli.SetDialer(c.dialer)
	err := cli.Connect(service.DialURL())
	if err != nil {
		return nil, err
	}
	// 4. read messages
	go func(cli qim.Client) {
		err := readloop(cli)
		if err != nil {
			c.lg.Debug(err.Error())
		}
		clients.Remove(id)
		cli.Close()
	}(cli)
	// 5.
	clients.Add(cli)
	return cli, nil
}

// Receive default listener
func readloop(cli qim.Client) error {
	log := c.lg.With(zap.String("func", "readloop"))
	log.Info("readloop starting", zap.String("serviceID", cli.ServiceID()), zap.String("serviceName", cli.ServiceName()))

	for {
		frame, err := cli.Read()
		if err != nil {
			log.Info(err.Error())
			return err
		}
		if frame.GetOpCode() != qim.OpBinary {
			continue
		}
		buf := bytes.NewBuffer(frame.GetPayload())

		packet, err := pkt.MustReadLogicPkt(buf)
		if err != nil {
			log.Info(err.Error())
			continue
		}
		err = pushMessage(packet)
		if err != nil {
			log.Info(err.Error())
		}
	}
}

// push the message to the channel through the gateway server
func pushMessage(packet *pkt.LogicPkt) error {
	server, _ := packet.GetMeta(wire.MetaDestServer)
	if server != c.Srv.ServiceID() {
		return fmt.Errorf("dest_server is incorrect, %s != %s", server, c.Srv.ServiceID())
	}
	channels, ok := packet.GetMeta(wire.MetaDestChannels)
	if !ok {
		return errors.New("dest_channels is nil")
	}

	channelIDs := strings.Split(channels.(string), ",")
	packet.DelMeta(wire.MetaDestServer)
	packet.DelMeta(wire.MetaDestChannels)
	payload := pkt.Marshal(packet)
	c.lg.Debug("pushing message", zap.Any("channels", channelIDs), zap.Any("packet", packet))

	for _, channel := range channelIDs {
		messageOutFlowBytes.WithLabelValues(packet.Command).Add(float64(len(payload)))
		err := c.Srv.Push(channel, payload)
		if err != nil {
			c.lg.Info(err.Error())
		}
	}
	return nil
}
