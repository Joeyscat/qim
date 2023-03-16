package handler

import (
	"errors"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/services/server/service"
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
	panic("unimplemented")
}

func (h *ChatHandler) DoGroupTalk(ctx qim.Context) {
	panic("unimplemented")
}

func (h *ChatHandler) DoTalkAck(ctx qim.Context) {
	panic("unimplemented")
}
