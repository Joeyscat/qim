package dialer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/wire"
	"github.com/joeyscat/qim/wire/pkt"
	"github.com/joeyscat/qim/wire/token"
)

type ClientDialer struct {
	AppSecret string
}

var _ qim.Dialer = (*ClientDialer)(nil)

// DialAndHandshake implements qim.Dialer
func (d *ClientDialer) DialAndHandshake(ctx qim.DialerContext) (net.Conn, error) {
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	if d.AppSecret == "" {
		d.AppSecret = token.DefaultSecret
	}

	tk, err := token.Generate(d.AppSecret, &token.Token{
		Account: ctx.ID,
		App:     "qim",
		Exp:     time.Now().AddDate(0, 0, 1).Unix(),
	})
	if err != nil {
		return nil, err
	}

	loginReq := pkt.New(wire.CommandLoginSignIn).WriteBody(&pkt.LoginReq{
		Token: tk,
	})
	err = wsutil.WriteClientBinary(conn, pkt.Marshal(loginReq))
	if err != nil {
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
	frame, err := ws.ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	ack, err := pkt.MustReadLogicPkt(bytes.NewBuffer(frame.Payload))
	if err != nil {
		return nil, err
	}

	if ack.Status != pkt.Status_Success {
		return nil, fmt.Errorf("login failed: %s", ack.Header.String())
	}
	var resp = new(pkt.LoginResp)
	_ = ack.ReadBody(resp)

	log.Println("login success", resp.GetChannelId())

	return conn, nil
}
