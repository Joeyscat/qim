package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joeyscat/qim/wire/rpcc"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const app = "qim_t"

var log, _ = zap.NewDevelopment()

var messageService = NewMessageServiceWithSRV("http", &resty.SRVRecord{
	Domain:  "message",
	Service: "royal",
}, log)

func TestMessage(t *testing.T) {
	m := rpcc.Message{
		Type: 1,
		Body: "hello",
	}
	dest := fmt.Sprintf("u%d", time.Now().Unix())
	_, err := messageService.InsertUser(app, &rpcc.InsertMessageReq{
		Sender:   "u1",
		Dest:     dest,
		SendTime: time.Now().UnixNano(),
		Message:  &m,
	})
	assert.NoError(t, err)

	resp, err := messageService.GetMessageIndex(app, &rpcc.GetOfflineMessageIndexReq{
		Account: dest,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.GetList()))

	index := resp.GetList()[0]
	assert.Equal(t, "u1", index.GetAccountB())

	resp2, err := messageService.GetMessageContent(app, &rpcc.GetOfflineMessageContentReq{
		MessageIds: []int64{index.GetMessageId()},
	})
	assert.NoError(t, err)

	assert.Equal(t, 1, len(resp2.GetList()))
	content := resp2.GetList()[0]
	assert.Equal(t, m.GetBody(), content.GetBody())
	assert.Equal(t, m.GetType(), content.GetType())
	assert.Equal(t, index.GetMessageId(), content.GetId())

	resp, err = messageService.GetMessageIndex(app, &rpcc.GetOfflineMessageIndexReq{
		Account:   dest,
		MessageId: index.GetMessageId(),
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(resp.GetList()))

	resp, err = messageService.GetMessageIndex(app, &rpcc.GetOfflineMessageIndexReq{
		Account: dest,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(resp.GetList()))
}
