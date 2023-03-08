package mock

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/tcp"
	"github.com/joeyscat/qim/websocket"
	"go.uber.org/zap"
)

type ClientDemo struct {
	lg *zap.Logger
}

func (c *ClientDemo) Start(userID, prorocol, addr string) {

	var cli qim.Client

	// 1. 初始化客户端
	if prorocol == "ws" {
		cli = websocket.NewClient(userID, "client", websocket.ClientOptions{})
		cli.SetDialer(&WebSocketDialer{})
	} else if prorocol == "tcp" {
		cli = tcp.NewClient(userID, "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{lg: c.lg.With(zap.String("module", "tcp.dialer"))})
	}

	// 2. 建立连接
	err := cli.Connect(addr)
	if err != nil {
		c.lg.Error(err.Error())
		return
	}
	c.lg.Info("connect finished")
	count := 10
	go func() {
		// 3. 发送消息，退出
		for i := 0; i < count; i++ {
			err := cli.Send([]byte(fmt.Sprintf("hello_%d", i)))
			if err != nil {
				c.lg.Error(err.Error())
				return
			}
			time.Sleep(time.Microsecond * 10)
		}
	}()

	// 4. 接收消息
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			c.lg.Warn(err.Error())
			break
		}
		if frame.GetOpCode() != qim.OpBinary {
			continue
		}
		recv++
		c.lg.Info("receive message", zap.String("payload", string(frame.GetPayload())))
		if recv == count { // 接收完消息
			break
		}
	}
	// 退出
	cli.Close()
}

type WebSocketDialer struct {
}

// DialAndHandshake implements qim.Dialer
func (*WebSocketDialer) DialAndHandshake(ctx qim.DialerContext) (net.Conn, error) {
	ctxWithTimeout, cancel := context.WithTimeout(context.TODO(), ctx.Timeout)
	defer cancel()

	// 1. 调用ws.Dial拨号
	conn, _, _, err := ws.Dial(ctxWithTimeout, ctx.Address)
	if err != nil {
		return nil, err
	}
	// 2. 发送用户认证信息，示例就是userID
	err = wsutil.WriteClientBinary(conn, []byte(ctx.ID))
	if err != nil {
		return nil, err
	}

	// 3. return conn
	return conn, nil
}

var _ qim.Dialer = (*WebSocketDialer)(nil)

type TCPDialer struct {
	lg *zap.Logger
}

// DialAndHandshake implements qim.Dialer
func (d *TCPDialer) DialAndHandshake(ctx qim.DialerContext) (net.Conn, error) {
	d.lg.Info("start tcp dial", zap.String("address", ctx.Address))
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}

	err = tcp.WriteFrame(conn, qim.OpBinary, []byte(ctx.ID))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

var _ qim.Dialer = (*TCPDialer)(nil)
