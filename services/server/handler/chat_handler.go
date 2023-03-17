package handler

import (
	"errors"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/services/server/service"
	"github.com/joeyscat/qim/wire/pkt"
	"github.com/joeyscat/qim/wire/rpc"
)

var ErrNoDestination = errors.New("dest is empty")

type ChatHandler struct {
	msgService   service.Message
	groupService service.Group
}

func NewChatHandler(message service.Message, group service.Group) *ChatHandler {
	return &ChatHandler{
		msgService:   message,
		groupService: group,
	}
}

func (h *ChatHandler) DoUserTalk(ctx qim.Context) {
	// validate
	if ctx.Header().Dest == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrNoDestination)
		return
	}

	// decode the message
	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	// get the location of the receiver
	receiver := ctx.Header().GetDest()
	loc, err := ctx.GetLocation(receiver, "")
	if err != nil && err != qim.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// save offline message
	sendTime := time.Now().Local().UnixNano()
	resp, err := h.msgService.InsertUser(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     receiver,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// push the message to the receiver if online
	if loc != nil {
		if err = ctx.Dispatch(&pkt.MessagePush{
			MessageId: resp.GetMessageId(),
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, loc); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	// response to the sender
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResp{
		MessageId: resp.GetMessageId(),
		SendTime:  sendTime,
	})
}

func (h *ChatHandler) DoGroupTalk(ctx qim.Context) {
	if ctx.Header().Dest == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrNoDestination)
		return
	}

	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	group := ctx.Header().GetDest()
	sendTime := time.Now().Local().UnixNano()

	resp, err := h.msgService.InsertGroup(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     group,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	membersResponse, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{
		GroupId: group,
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var members = make([]string, len(membersResponse.GetUsers()))
	for i, user := range membersResponse.GetUsers() {
		members[i] = user.Account
	}

	locs, err := ctx.GetLocations(members...)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	if len(locs) > 0 {
		if err = ctx.Dispatch(&pkt.MessagePush{
			MessageId: resp.GetMessageId(),
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, locs...); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResp{
		MessageId: resp.MessageId,
		SendTime:  sendTime,
	})
}

func (h *ChatHandler) DoTalkAck(ctx qim.Context) {
	var req pkt.MessageAckReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	err := h.msgService.SetAck(ctx.Session().GetApp(), &rpc.AckMessageReq{
		Account:   ctx.Session().GetAccount(),
		MessageId: req.GetMessageId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}
