package handler

import (
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/wire/pkt"
	"go.uber.org/zap"
)

type loginHandler struct {
	lg *zap.Logger
}

func NewLoginHandler(lg *zap.Logger) *loginHandler {
	return &loginHandler{lg: lg}
}

func (h *loginHandler) DoSysLogin(ctx qim.Context) {
	var session pkt.Session
	if err := ctx.ReadBody(&session); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	h.lg.Info("do login", zap.String("session", session.String()))

	// check if this account is already logged in
	old, err := ctx.GetLocation(session.GetAccount(), "")
	if err != nil && err != qim.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	if old != nil {
		// kick out
		_ = ctx.Dispatch(&pkt.KickoutNotify{ChannelId: old.ChannelID}, old)
	}

	// save session
	err = ctx.Add(&session)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	var resp = &pkt.LoginResp{
		ChannelId: session.GetChannelId(),
		Account:   session.GetAccount(),
	}
	_ = ctx.Resp(pkt.Status_Success, resp)
}

func (h *loginHandler) DoSysLogout(ctx qim.Context) {
	h.lg.Info("do logout", zap.String("channelID", ctx.Session().GetChannelId()), zap.String("account", ctx.Session().GetAccount()))

	err := ctx.Delete(ctx.Session().GetAccount(), ctx.Session().GetChannelId())
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}
