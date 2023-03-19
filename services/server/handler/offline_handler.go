package handler

import (
	"errors"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/services/server/service"
	"github.com/joeyscat/qim/wire/pkt"
	"github.com/joeyscat/qim/wire/rpcc"
)

type OfflineHandler struct {
	msgService service.Message
}

func NewOfflineHandler(message service.Message) *OfflineHandler {
	return &OfflineHandler{
		msgService: message,
	}
}

func (h *OfflineHandler) DoSyncIndex(ctx qim.Context) {
	var req pkt.MessageIndex
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	resp, err := h.msgService.GetMessageIndex(ctx.Session().GetApp(), &rpcc.GetOfflineMessageIndexReq{
		Account:   ctx.Session().GetAccount(),
		MessageId: req.GetMessageId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	var list = make([]*pkt.MessageIndex, 0, len(resp.GetList()))
	for i, value := range resp.GetList() {
		list[i] = &pkt.MessageIndex{
			Account:   value.GetAccountB(),
			Direction: value.GetDirection(),
			Group:     value.GetGroup(),
			MessageId: value.GetMessageId(),
			SendTime:  value.GetSendTime(),
		}
	}
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageIndexResp{
		Indexes: list,
	})
}

func (h *OfflineHandler) DoSyncContent(ctx qim.Context) {
	var req pkt.MessageContentReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	if len(req.GetMessageIds()) == 0 {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, errors.New("empty message ids"))
		return
	}

	resp, err := h.msgService.GetMessageContent(ctx.Session().GetApp(), &rpcc.GetOfflineMessageContentReq{
		MessageIds: req.GetMessageIds(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	var list = make([]*pkt.MessageContent, 0, len(resp.GetList()))
	for i, value := range resp.GetList() {
		list[i] = &pkt.MessageContent{
			MessageId: value.GetId(),
			Type:      value.GetType(),
			Body:      value.GetBody(),
			Extra:     value.GetExtra(),
		}
	}
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageContentResp{
		Contents: list,
	})
}
