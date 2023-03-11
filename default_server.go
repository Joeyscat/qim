package qim

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/pool/pbufio"
	"github.com/gobwas/ws"
	"github.com/panjf2000/ants/v2"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Upgrader interface {
	Name() string
	Upgrade(rawconn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (Conn, error)
}

type ServerOptions struct {
	Loginwait       time.Duration
	Readwait        time.Duration
	Writewait       time.Duration
	MessageGPool    int
	ConnectionGPool int
}

type ServerOption func(*ServerOptions)

func WithMessageGPool(val int) ServerOption {
	return func(opts *ServerOptions) {
		opts.MessageGPool = val
	}
}

func WithConnectionGPool(val int) ServerOption {
	return func(opt *ServerOptions) {
		opt.ConnectionGPool = val
	}
}

// DefaultServer is a websocket implemnetation of qim.Server
type DefaultServer struct {
	Upgrader
	listen string
	ServiceRegistration
	ChannelMap
	Acceptor
	MessageListener
	StateListener
	once    sync.Once
	options *ServerOptions
	quit    int32

	lg *zap.Logger
}

// Push implements Server
func (s *DefaultServer) Push(channelID string, payload []byte) error {
	ch, ok := s.ChannelMap.Get(channelID)
	if !ok {
		return fmt.Errorf("channel not found: %s", channelID)
	}
	return ch.Push(payload)
}

// SetAcceptor implements Server
func (s *DefaultServer) SetAcceptor(acceptor Acceptor) {
	s.Acceptor = acceptor
}

// SetChannelMap implements Server
func (s *DefaultServer) SetChannelMap(channelMap ChannelMap) {
	s.ChannelMap = channelMap
}

// SetMessageListener implements Server
func (s *DefaultServer) SetMessageListener(messageListener MessageListener) {
	s.MessageListener = messageListener
}

// SetReadwait implements Server
func (s *DefaultServer) SetReadwait(readwait time.Duration) {
	s.options.Readwait = readwait
}

// SetStateListener implements Server
func (s *DefaultServer) SetStateListener(stateListener StateListener) {
	s.StateListener = stateListener
}

// Shutdown implements Server
func (s *DefaultServer) Shutdown(ctx context.Context) error {
	s.once.Do(func() {
		defer func() {
			s.lg.Info("shutdown")
		}()

		if atomic.CompareAndSwapInt32(&s.quit, 0, 1) {
			return
		}

		// close channels
		channels := s.ChannelMap.All()
		for _, ch := range channels {
			ch.Close()

			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})

	return nil
}

// Start implements Server
func (s *DefaultServer) Start() error {
	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.ChannelMap == nil {
		s.ChannelMap = NewChannels(100)
	}
	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}

	mgpool, _ := ants.NewPool(s.options.MessageGPool, ants.WithPreAlloc(true))
	defer func() {
		mgpool.Release()
	}()

	log := s.lg.With(zap.String("listen", s.listen), zap.String("func", "Start"))
	log.Info("started")

	for {
		rawconn, err := lst.Accept()
		if err != nil {
			if rawconn != nil {
				rawconn.Close()
			}
			log.Warn(err.Error())
			continue
		}

		go s.connHandler(rawconn, mgpool)

		if atomic.LoadInt32(&s.quit) == 1 {
			break
		}
	}

	log.Info("quit")
	return nil
}

func (s *DefaultServer) connHandler(rawconn net.Conn, gpool *ants.Pool) {
	rd := pbufio.GetReader(rawconn, ws.DefaultServerReadBufferSize)
	wr := pbufio.GetWriter(rawconn, ws.DefaultServerWriteBufferSize)
	defer func() {
		pbufio.PutReader(rd)
		pbufio.PutWriter(wr)
	}()

	conn, err := s.Upgrade(rawconn, rd, wr)
	if err != nil {
		s.lg.Error("Upgrade error", zap.Error(err))
		rawconn.Close()
		return
	}

	id, meta, err := s.Accept(conn, s.options.Loginwait)
	if err != nil {
		_ = conn.WriteFrame(OpClose, []byte(err.Error()))
		conn.Close()
		return
	}
	if _, ok := s.Get(id); ok {
		_ = conn.WriteFrame(OpClose, []byte("channelId is repeated"))
		conn.Close()
		return
	}
	if meta == nil {
		meta = Meta{}
	}

	channel := NewChannel(id, meta, conn, gpool, s.lg)
	channel.SetReadwait(s.options.Readwait)
	channel.SetWritewait(s.options.Writewait)
	s.Add(channel)

	gaugeWithLable := channelTotalGauge.WithLabelValues(s.ServiceID(), s.ServiceName())
	gaugeWithLable.Inc()
	defer gaugeWithLable.Dec()

	s.lg.Info("accept channel", zap.String("channelID", channel.ID()),
		zap.String("remoteAddr", channel.RemoteAddr().String()))

	err = channel.Readloop(s.MessageListener)
	if err != nil {
		// TODO Info or Warn?
		s.lg.Info(err.Error())
	}
	s.Remove(channel.ID())
	_ = s.Disconnect(channel.ID())
	channel.Close()
}

var _ Server = (*DefaultServer)(nil)

func NewServer(
	listen string,
	service ServiceRegistration,
	upgrader Upgrader,
	options ...ServerOption,
) *DefaultServer {
	defaultOpts := &ServerOptions{
		Loginwait:       DefaultLoginwait,
		Readwait:        DefaultReadwait,
		Writewait:       DefaultWritewait,
		MessageGPool:    DefaultMessageReadPool,
		ConnectionGPool: DefaultConnectionPool,
	}
	for _, opt := range options {
		opt(defaultOpts)
	}
	s := &DefaultServer{
		listen:              listen,
		ServiceRegistration: service,
		options:             defaultOpts,
		Upgrader:            upgrader,
		quit:                0,
	}

	var err error
	if os.Getenv("DEBUG") == "true" {
		s.lg, err = zap.NewDevelopment(zap.Fields(
			zap.String("module", upgrader.Name()),
			zap.String("id", service.ServiceID()),
		))
	} else {
		s.lg, err = zap.NewProduction(zap.Fields(
			zap.String("module", upgrader.Name()),
			zap.String("id", service.ServiceID()),
		))
	}
	if err != nil {
		log.Fatalln(err)
	}

	return s
}

type defaultAcceptor struct {
}

// Accept implements Acceptor
func (*defaultAcceptor) Accept(conn Conn, timeout time.Duration) (string, Meta, error) {
	return ksuid.New().String(), Meta{}, nil
}

var _ Acceptor = (*defaultAcceptor)(nil)
