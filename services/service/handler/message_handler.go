package handler

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joeyscat/qim/services/service/database"
	"github.com/joeyscat/qim/wire"
	"github.com/joeyscat/qim/wire/rpc"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

type ServiceHandler struct {
	BaseDB    *gorm.DB
	MessageDB *gorm.DB
	Cache     *redis.Client
	IDgen     *database.IDGenerator
}

func (h *ServiceHandler) InsertUserMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	messageID, err := h.insertUserMessage(&req)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.InsertMessageResp{
		MessageId: messageID,
	})
}

func (h *ServiceHandler) insertUserMessage(req *rpc.InsertMessageReq) (int64, error) {
	messageID := h.IDgen.Next().Int64()

	// diffusion write
	idxs := make([]database.MessageIndex, 2)
	idxs[0] = database.MessageIndex{
		ID:        h.IDgen.Next().Int64(),
		MessageID: messageID,
		AccountA:  req.GetDest(),
		AccountB:  req.GetSender(),
		Direction: 0,
		SendTime:  req.GetSendTime(),
	}
	idxs[1] = database.MessageIndex{
		ID:        h.IDgen.Next().Int64(),
		MessageID: messageID,
		AccountA:  req.GetSender(),
		AccountB:  req.GetDest(),
		Direction: 1,
		SendTime:  req.GetSendTime(),
	}

	messageContent := database.MessageContent{
		ID:       messageID,
		Type:     byte(req.GetMessage().GetType()),
		Body:     req.GetMessage().GetBody(),
		Extra:    req.GetMessage().GetExtra(),
		SendTime: req.GetSendTime(),
	}

	err := h.MessageDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}
		if err := tx.Create(&idxs).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return messageID, nil
}

func (h *ServiceHandler) InsertGroupMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	messageID, err := h.insertGroupMessage(&req)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.InsertMessageResp{
		MessageId: messageID,
	})
}

func (h *ServiceHandler) insertGroupMessage(req *rpc.InsertMessageReq) (int64, error) {
	messageID := h.IDgen.Next().Int64()

	var members []database.GroupMember
	err := h.BaseDB.Where(&database.GroupMember{Group: req.Dest}).Find(&members).Error
	if err != nil {
		return 0, err
	}

	// diffusion write
	idxs := make([]database.MessageIndex, len(members))
	for i, member := range members {
		idxs[i] = database.MessageIndex{
			ID:        h.IDgen.Next().Int64(),
			MessageID: messageID,
			AccountA:  member.Account,
			AccountB:  req.GetSender(),
			Direction: 0,
			Group:     member.Group,
			SendTime:  req.GetSendTime(),
		}
		if member.Account == req.GetSender() {
			idxs[i].Direction = 1
		}
	}

	messageContent := database.MessageContent{
		ID:       messageID,
		Type:     byte(req.GetMessage().GetType()),
		Body:     req.GetMessage().GetBody(),
		Extra:    req.GetMessage().GetExtra(),
		SendTime: req.GetSendTime(),
	}

	err = h.MessageDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}
		if err := tx.Create(&idxs).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return messageID, nil
}

func (h *ServiceHandler) MessageAck(c iris.Context) {
	var req rpc.AckMessageReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	err := setMessageAck(h.Cache, req.GetAccount(), req.GetMessageId())
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func setMessageAck(cache *redis.Client, account string, messageID int64) error {
	if messageID == 0 {
		return nil
	}

	key := database.KeyMessageAckIndex(account)
	return cache.Set(cache.Context(), key, messageID, wire.OfflineReadIndexExpiresIn).Err()
}

func (h *ServiceHandler) GetOfflineMessageIndex(c iris.Context) {
	var req rpc.GetOfflineMessageIndexReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	start, err := h.getSendTime(req.GetAccount(), req.GetMessageId())
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	var indexes []*rpc.MessageIndex
	tx := h.MessageDB.Model(&database.MessageIndex{}).Select("send_time", "account_b", "direction", "message_id", "group")
	err = tx.Where("account_a = ? AND send_time > ? and direction = ?", req.GetAccount(), start, 0).Order("send_time asc").
		Limit(wire.OfflineSyncIndexCount).Find(&indexes).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	err = setMessageAck(h.Cache, req.GetAccount(), req.GetMessageId())
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.GetOfflineMessageIndexResp{
		List: indexes,
	})
}

func (h *ServiceHandler) getSendTime(account string, messageID int64) (int64, error) {
	if messageID == 0 {
		key := database.KeyMessageAckIndex(account)
		messageID, _ = h.Cache.Get(h.Cache.Context(), key).Int64()
	}

	var start int64
	if messageID > 0 {
		var content database.MessageContent
		err := h.MessageDB.Select("send_time").First(&content, messageID).Error
		if err != nil {
			start = time.Now().AddDate(0, 0, -1).UnixNano()
		} else {
			start = content.SendTime
		}
	}

	earliestKeepTime := time.Now().AddDate(0, 0, -1*wire.OfflineMessageExpiresIn).UnixNano()
	if start == 0 || start < earliestKeepTime {
		start = earliestKeepTime
	}

	return start, nil
}

func (h *ServiceHandler) GetOfflineMessageContent(c iris.Context) {
	var req rpc.GetOfflineMessageContentReq
	if err := c.ReadJSON(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	mlen := len(req.GetMessageIds())
	if mlen > wire.MessageMaxCountPerPage {
		c.StopWithText(iris.StatusBadRequest, "too many message ids")
		return
	}

	var contents []*rpc.Message
	err := h.MessageDB.Model(&database.MessageContent{}).Where(req.GetMessageIds()).Find(&contents).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.GetOfflineMessageContentResp{
		List: contents,
	})
}
