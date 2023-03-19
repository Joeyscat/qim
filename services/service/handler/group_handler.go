package handler

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/joeyscat/qim/services/service/database"
	"github.com/joeyscat/qim/wire/rpc"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func (h *ServiceHandler) GroupCreate(c iris.Context) {
	app := c.Params().Get("app")
	var req rpc.CreateGroupReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	req.App = app
	groupID, err := h.groupCreate(&req)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.CreateGroupResp{
		GroupId: groupID.Base36(),
	})
}

func (h *ServiceHandler) groupCreate(req *rpc.CreateGroupReq) (snowflake.ID, error) {
	groupID := h.IDgen.Next()
	g := &database.Group{
		Model:       database.Model{ID: groupID.Int64()},
		App:         req.GetApp(),
		Avatar:      req.GetAvatar(),
		Group:       groupID.Base36(),
		Introdution: req.GetIntroduction(),
		Name:        req.GetName(),
		Owner:       req.GetOwner(),
	}

	members := make([]database.GroupMember, len(req.GetMembers()))
	for i, user := range req.GetMembers() {
		members[i] = database.GroupMember{
			Model:   database.Model{ID: h.IDgen.Next().Int64()},
			Account: user,
			Group:   groupID.Base36(),
		}
	}

	err := h.BaseDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			return err
		}

		if err := tx.Create(&members).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func (h *ServiceHandler) GroupJoin(c iris.Context) {
	var req rpc.JoinGroupReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	gm := &database.GroupMember{
		Model:   database.Model{ID: h.IDgen.Next().Int64()},
		Account: req.GetAccount(),
		Group:   req.GetGroupId(),
	}
	err := h.BaseDB.Create(gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupQuit(c iris.Context) {
	var req rpc.QuitGroupReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	gm := &database.GroupMember{
		Account: req.GetAccount(),
		Group:   req.GetGroupId(),
	}
	err := h.BaseDB.Delete(&database.GroupMember{}, gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupMembers(c iris.Context) {
	group := c.Params().Get("id")
	if group == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("group id is empty"))
		return
	}

	var members []database.GroupMember
	err := h.BaseDB.Order("Updated_At asc").Find(&members, database.GroupMember{Group: group}).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	var users = make([]*rpc.Member, len(members))
	for i, m := range members {
		users[i] = &rpc.Member{
			Account:  m.Account,
			Alias:    m.Alias,
			JoinTime: m.CreatedAt.Unix(),
		}
	}
	_, _ = c.Negotiate(&rpc.GroupMembersResp{
		Users: users,
	})
}

func (h *ServiceHandler) GroupGet(c iris.Context) {
	groupID := c.Params().Get("id")
	if groupID == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("group id is empty"))
		return
	}

	id, err := h.IDgen.ParseBase36(groupID)
	if err != nil {
		c.StopWithError(iris.StatusBadRequest, fmt.Errorf("group id is invalid: %w", err))
		return
	}

	var group database.Group
	err = h.BaseDB.First(&group, id.Int64()).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.GetGroupResp{
		Id:           groupID,
		Name:         group.Name,
		Avatar:       group.Avatar,
		Introduction: group.Introdution,
		Owner:        group.Owner,
		CreatedAt:    group.CreatedAt.Unix(),
	})
}
