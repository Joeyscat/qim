package websocket

import (
	"bufio"
	"net"

	"github.com/gobwas/ws"
	"github.com/joeyscat/qim"
)

type Upgrader struct {
}

var _ qim.Upgrader = (*Upgrader)(nil)

func (u *Upgrader) Name() string {
	return "websocket.Server"
}

func (u *Upgrader) Upgrade(rawconn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (qim.Conn, error) {

	_, err := ws.Upgrade(rawconn)
	if err != nil {
		return nil, err
	}
	conn := NewConnWithRW(rawconn, rd, wr)
	return conn, nil
}

func NewServer(listen string, service qim.ServiceRegistration, options ...qim.ServerOption) qim.Server {
	return qim.NewServer(listen, service, new(Upgrader), options...)
}
