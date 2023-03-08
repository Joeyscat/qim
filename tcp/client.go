package tcp

import (
	"errors"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joeyscat/qim"
	"go.uber.org/zap"
)

type ClientOptions struct {
	Heartbeat time.Duration
	Readwait  time.Duration
	Writewait time.Duration
}

type Client struct {
	sync.Mutex
	qim.Dialer
	once    sync.Once
	id      string
	name    string
	conn    qim.Conn
	state   int32
	options ClientOptions
	meta    map[string]string
	lg      *zap.Logger
}

func NewClient(id, name string, opts ClientOptions) qim.Client {
	return NewClientWithProps(id, name, make(map[string]string), opts)
}

func NewClientWithProps(id, name string, meta map[string]string, opts ClientOptions) qim.Client {
	if opts.Writewait == 0 {
		opts.Writewait = qim.DefaultWritewait
	}
	if opts.Readwait == 0 {
		opts.Readwait = qim.DefaultReadwait
	}

	var err error
	var logger *zap.Logger
	if os.Getenv("DEBUG") == "true" {
		logger, err = zap.NewDevelopment(zap.Fields(
			zap.String("module", "tcp.client"), zap.String("id", id)))
	} else {
		logger, err = zap.NewProduction(zap.Fields(
			zap.String("module", "tcp.client"), zap.String("id", id)))
	}
	if err != nil {
		log.Fatalln(err)
	}

	cli := &Client{
		id:      id,
		name:    name,
		options: opts,
		meta:    meta,
		lg:      logger,
	}
	return cli
}

// GetMeta implements qim.Client
func (c *Client) GetMeta() map[string]string {
	return c.meta
}

// ServiceID implements qim.Client
func (c *Client) ServiceID() string {
	return c.id
}

// ServiceName implements qim.Client
func (c *Client) ServiceName() string {
	return c.name
}

// Close implements qim.Client
func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}
		_ = c.conn.WriteFrame(qim.OpClose, nil)
		_ = c.conn.Flush()

		_ = c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

// Connect implements qim.Client
func (c *Client) Connect(addr string) error {
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return errors.New("invalid client state")
	}

	rawconn, err := c.DialAndHandshake(qim.DialerContext{
		ID:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: qim.DefaultLoginwait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if rawconn == nil {
		return errors.New("connection is nil")
	}
	c.conn = NewConn(rawconn)

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatloop(rawconn)
			if err != nil {
				c.lg.Error("heartbeatloop stopped -- ", zap.Error(err))
			}
		}()
	}

	return nil
}

// Read implements qim.Client
func (c *Client) Read() (qim.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connecion is nil")
	}
	if c.options.Heartbeat > 0 {
		// 心跳控制: heartbeatloop()负责发送ping，这里设置readwait
		// 如果服务端正常返回pong，这里会一直刷新readDeadline
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.Readwait))
	}
	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == qim.OpClose {
		return nil, errors.New("remote side close the channel")
	}

	return frame, nil
}

// Send implements qim.Client
func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return errors.New("connection is nil")
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.WriteFrame(qim.OpBinary, payload)
	if err != nil {
		return err
	}
	return c.conn.Flush()
}

// SetDialer implements qim.Client
func (c *Client) SetDialer(dialer qim.Dialer) {
	c.Dialer = dialer
}

func (c *Client) heartbeatloop(conn net.Conn) error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		if err := c.ping(conn); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ping(conn net.Conn) error {
	c.lg.Debug("send ping to server")

	err := c.conn.WriteFrame(qim.OpPing, nil)
	if err != nil {
		return err
	}
	return c.conn.Flush()
}
