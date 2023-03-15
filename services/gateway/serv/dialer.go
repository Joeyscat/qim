package serv

import (
	"net"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/tcp"
	"github.com/joeyscat/qim/wire/pkt"
	"google.golang.org/protobuf/proto"
)

type TcpDialer struct {
	ServiceID string
}

func NewDialer(serviceID string) qim.Dialer {
	return &TcpDialer{
		ServiceID: serviceID,
	}
}

// DialAndHandshake implements qim.Dialer
func (d *TcpDialer) DialAndHandshake(ctx qim.DialerContext) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	req := &pkt.InnerHandshakeReq{
		ServiceId: d.ServiceID,
	}

	bts, _ := proto.Marshal(req)
	err = tcp.WriteFrame(conn, qim.OpBinary, bts)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
