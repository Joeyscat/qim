package mock

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/websocket"
	"go.uber.org/zap"
)

type ServerDemo struct {
	lg *zap.Logger
}

func (s *ServerDemo) Start(id, protocol, addr string) {
	go func() {
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	var srv qim.Server
	service := naming.NewEntry(id, "", protocol, "", 1)
	if protocol == "ws" {
		srv = websocket.NewServer(addr, service)
	} else if protocol == "tcp" {
		log.Fatal("unimplement")
	}

	handler := &ServerHandler{
		lg: s.lg,
	}

	srv.SetReadwait(time.Minute)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}

type ServerHandler struct {
	lg *zap.Logger
}

// Accept implements qim.Acceptor
func (h *ServerHandler) Accept(conn qim.Conn, timeout time.Duration) (string, qim.Meta, error) {
	// 1. 读取：客户端发送的鉴权数据包
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", nil, err
	}
	//2. 解析：数据包内容就是userId
	userID := string(frame.GetPayload())
	// 3. 鉴权：这里只是为了示例做一个fake验证，非空
	if userID == "" {
		return "", nil, errors.New("user id is invalid")
	}
	h.lg.Info("logined", zap.String("userID", userID))
	return userID, nil, nil
}

// Receive implements qim.MessageListener
func (h *ServerHandler) Receive(agent qim.Agent, payload []byte) {
	h.lg.Info("received", zap.String("payload", string(payload)))
	_ = agent.Push([]byte("ok"))
}

// Disconnect implements qim.StateListener
func (h *ServerHandler) Disconnect(channelID string) error {
	h.lg.Info("disconnect", zap.String("id", channelID))
	return nil
}

var _ qim.Acceptor = (*ServerHandler)(nil)
var _ qim.MessageListener = (*ServerHandler)(nil)
var _ qim.StateListener = (*ServerHandler)(nil)
