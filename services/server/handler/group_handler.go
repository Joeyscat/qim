package handler

import (
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/services/server/service"
	"github.com/joeyscat/qim/wire/pkt"
	"github.com/joeyscat/qim/wire/rpcc"
)

type GroupHandler struct {
	groupService service.Group
}

func NewGroupHandler(groupService service.Group) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

func (h *GroupHandler) DoCreate(ctx qim.Context) {
	var req pkt.GroupCreateReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	resp, err := h.groupService.Create(ctx.Session().GetApp(), &rpcc.CreateGroupReq{
		Name:         req.GetName(),
		Avatar:       req.GetAvatar(),
		Introduction: req.GetIntroduction(),
		Owner:        req.GetOwner(),
		Members:      req.GetMembers(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	locs, err := ctx.GetLocations(req.GetMembers()...)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	if len(locs) > 0 {
		if err = ctx.Dispatch(&pkt.GroupCreateNotify{
			GroupId: resp.GetGroupId(),
			Members: req.GetMembers(),
		}, locs...); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupCreateResp{
		GroupId: resp.GetGroupId(),
	})
}

func (h *GroupHandler) DoJoin(ctx qim.Context) {
	var req pkt.GroupJoinReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	err := h.groupService.Join(ctx.Session().GetApp(), &rpcc.JoinGroupReq{
		Account: req.GetAccount(),
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}

func (h *GroupHandler) DoQuit(ctx qim.Context) {
	var req pkt.GroupQuitReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	err := h.groupService.Quit(ctx.Session().GetApp(), &rpcc.QuitGroupReq{
		Account: req.GetAccount(),
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}

func (h *GroupHandler) DoDetail(ctx qim.Context) {
	var req pkt.GroupGetReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	resp, err := h.groupService.Detail(ctx.Session().GetApp(), &rpcc.GetGroupReq{
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	membersResp, err := h.groupService.Members(ctx.Session().GetApp(), &rpcc.GroupMembersReq{
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var members = make([]*pkt.Member, len(membersResp.GetUsers()))
	for i, m := range membersResp.GetUsers() {
		members[i] = &pkt.Member{
			Account:  m.GetAccount(),
			Alias:    m.GetAlias(),
			Avatar:   m.GetAvatar(),
			JoinTime: m.GetJoinTime(),
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupGetResp{
		Id:           resp.GetId(),
		Name:         resp.GetName(),
		Introduction: resp.GetIntroduction(),
		Avatar:       resp.GetAvatar(),
		Owner:        resp.GetOwner(),
		Members:      members,
	})
}
