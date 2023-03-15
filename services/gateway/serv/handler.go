package serv

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/container"
	"github.com/joeyscat/qim/wire"
	"github.com/joeyscat/qim/wire/pkt"
	"github.com/joeyscat/qim/wire/token"
	"go.uber.org/zap"
)

const (
	MetaKeyApp     = "app"
	MetaKeyAccount = "account"
)

type Handler struct {
	ServiceID string
	AppSecret string
	Lg        *zap.Logger
}

var _ qim.Acceptor = (*Handler)(nil)
var _ qim.MessageListener = (*Handler)(nil)
var _ qim.StateListener = (*Handler)(nil)

// Accept implements qim.Acceptor
func (h *Handler) Accept(conn qim.Conn, timeout time.Duration) (string, qim.Meta, error) {
	// read the login packet
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", nil, err
	}

	buf := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		h.Lg.Error("read packet error", zap.Error(err))
		return "", nil, err
	}

	// req must be a login packet
	if req.Command != wire.CommandLoginSignIn {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_InvalidCommand
		_ = conn.WriteFrame(qim.OpBinary, pkt.Marshal(resp))
		return "", nil, errors.New("must be a SignIn command")
	}

	// decode the login packet
	var login pkt.LoginReq
	err = req.ReadBody(&login)
	if err != nil {
		return "", nil, err
	}
	secret := h.AppSecret
	if secret == "" {
		secret = token.DefaultSecret
	}

	// parse token
	tk, err := token.Parse(secret, login.Token)
	if err != nil {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_Unauthorized
		_ = conn.WriteFrame(qim.OpBinary, pkt.Marshal(resp))
		return "", nil, err
	}

	// generate a globally unique ChannelID.
	id := generateChannelID(h.ServiceID, tk.Account)
	h.Lg.Info("accept channel", zap.Any("token", tk), zap.String("channelID", id))

	req.ChannelId = id
	req.WriteBody(&pkt.Session{
		Account:   tk.Account,
		ChannelId: id,
		GateId:    h.ServiceID,
		App:       tk.App,
		RemoteIp:  getIP(conn.RemoteAddr().String()),
	})
	req.AddStringMeta(MetaKeyApp, tk.App)
	req.AddStringMeta(MetaKeyAccount, tk.Account)

	err = container.Forward(wire.SNLogin, req)
	if err != nil {
		h.Lg.Error("container.Forward error", zap.Error(err))
		return "", nil, err
	}

	return id, qim.Meta{
		MetaKeyApp:     tk.App,
		MetaKeyAccount: tk.Account,
	}, nil
}

// Receive implements qim.MessageListener
func (h *Handler) Receive(agent qim.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.Read(buf)
	if err != nil {
		h.Lg.Error("read packet error", zap.Error(err))
		return
	}

	if basicPkt, ok := packet.(*pkt.BasicPkt); ok {
		if basicPkt.Code == pkt.CodePing {
			_ = agent.Push(pkt.Marshal(&pkt.BasicPkt{Code: pkt.CodePong}))
		}
		return
	}

	if logicPkt, ok := packet.(*pkt.LogicPkt); ok {
		logicPkt.ChannelId = agent.ID()

		messageInTotal.WithLabelValues(h.ServiceID, wire.SNTGateway, logicPkt.GetCommand()).Inc()
		messageInFlowBytes.WithLabelValues(h.ServiceID, wire.SNTGateway, logicPkt.GetCommand()).Add(float64(len(payload)))

		if agent.GetMeta() != nil {
			logicPkt.AddStringMeta(MetaKeyApp, agent.GetMeta()[MetaKeyApp])
			logicPkt.AddStringMeta(MetaKeyAccount, agent.GetMeta()[MetaKeyAccount])
		}

		err = container.Forward(logicPkt.ServiceName(), logicPkt)
		if err != nil {
			h.Lg.Error("container.Forward error", zap.Error(err),
				zap.String("id", agent.ID()),
				zap.String("command", logicPkt.GetCommand()),
				zap.String("dest", logicPkt.GetDest()))
		}
	}
}

// Disconnect implements qim.StateListener
func (h *Handler) Disconnect(channelID string) error {
	h.Lg.Info("disconnect", zap.String("channelID", channelID))

	logout := pkt.New(wire.CommandLoginSignOut, pkt.WithChannel(channelID))
	err := container.Forward(wire.SNLogin, logout)
	if err != nil {
		h.Lg.Error("logout error", zap.Error(err))
		return err
	}

	return nil
}

var ipExp = regexp.MustCompile(string("\\:[0-9]+$"))

func getIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	return ipExp.ReplaceAllString(remoteAddr, "")
}

func generateChannelID(serviceID, account string) string {
	return fmt.Sprintf("%s_%s_%d", serviceID, account, wire.Seq.Next())
}
