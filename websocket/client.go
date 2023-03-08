package websocket

import (
	"errors"
	"log"
	"net"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
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
	conn    net.Conn
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
		logger, err = zap.NewDevelopment(zap.Fields(zap.String("id", id)))
	} else {
		logger, err = zap.NewProduction(zap.Fields(zap.String("id", id)))
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
		_ = wsutil.WriteClientMessage(c.conn, ws.OpClose, nil)

		c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

// Connect implements qim.Client
func (c *Client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return errors.New("invalid client state")
	}

	// 1. 拨号及握手
	conn, err := c.DialAndHandshake(qim.DialerContext{
		ID:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: qim.DefaultLoginwait,
	})
	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if conn == nil {
		return errors.New("connection is nil")
	}
	c.conn = conn

	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbeatloop(conn)
			if err != nil {
				c.lg.Error("heartbeatloop stopped -- ", zap.Error(err))
			}
		}()
	}

	return nil
}

// Read implements qim.Client
// Read a frame, this function is not safety for concurrent
func (c *Client) Read() (qim.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connecion is nil")
	}
	if c.options.Heartbeat > 0 {
		// 心跳控制: heartbeatloop()负责发送ping，这里设置readwait
		// 如果服务端正常返回pong，这里会一直刷新readDeadline
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.Readwait))
	}
	frame, err := ws.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}
	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("remote side close the channel")
	}

	return &Frame{
		raw: frame,
	}, nil
}

// Send implements qim.Client
func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return errors.New("connection is nil")
	}
	c.Lock()
	defer c.Unlock()
	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.Writewait))
	if err != nil {
		return err
	}
	// 客户端消息需要使用MASK
	return wsutil.WriteClientMessage(c.conn, ws.OpBinary, payload)
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
	c.Lock()
	defer c.Unlock()
	// 通过SetWriteDeadline可以感知到发送端的异常
	err := conn.SetWriteDeadline(time.Now().Add(c.options.Writewait))
	if err != nil {
		return err
	}
	c.lg.Debug("send ping to server")
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}
