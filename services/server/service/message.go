package service

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joeyscat/qim/wire/rpcc"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Message interface {
	InsertUser(app string, req *rpcc.InsertMessageReq) (*rpcc.InsertMessageResp, error)
	InsertGroup(app string, req *rpcc.InsertMessageReq) (*rpcc.InsertMessageResp, error)
	SetAck(app string, req *rpcc.AckMessageReq) error
	GetMessageIndex(app string, req *rpcc.GetOfflineMessageIndexReq) (*rpcc.GetOfflineMessageIndexResp, error)
	GetMessageContent(app string, req *rpcc.GetOfflineMessageContentReq) (*rpcc.GetOfflineMessageContentResp, error)
}

type MessageHttp struct {
	url string
	cli *resty.Client
	srv *resty.SRVRecord
	lg  *zap.Logger
}

func NewMessageService(url string, lg *zap.Logger) Message {
	client := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	client.SetHeader("Content-Type", "application/x-protobuf")
	client.SetHeader("Accept", "application/x-protobuf")
	return &MessageHttp{
		url: url,
		cli: client,
		lg:  lg,
	}
}
func NewMessageServiceWithSRV(scheme string, srv *resty.SRVRecord, lg *zap.Logger) Message {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-Type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme(scheme)

	return &MessageHttp{
		url: "",
		cli: cli,
		srv: srv,
		lg:  lg,
	}
}

// GetMessageContent implements Message
func (m *MessageHttp) GetMessageContent(app string, req *rpcc.GetOfflineMessageContentReq) (*rpcc.GetOfflineMessageContentResp, error) {
	path := fmt.Sprintf("%s/api/%s/offline/content", m.url, app)

	body, _ := proto.Marshal(req)
	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.GetMessageContent - http status code: %d", response.StatusCode())
	}

	var resp rpcc.GetOfflineMessageContentResp
	_ = proto.Unmarshal(response.Body(), &resp)
	return &resp, nil
}

// GetMessageIndex implements Message
func (m *MessageHttp) GetMessageIndex(app string, req *rpcc.GetOfflineMessageIndexReq) (*rpcc.GetOfflineMessageIndexResp, error) {
	path := fmt.Sprintf("%s/api/%s/offline/index", m.url, app)
	body, _ := proto.Marshal(req)

	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.GetMessageIndex - http status code: %d", response.StatusCode())
	}

	var resp rpcc.GetOfflineMessageIndexResp
	_ = proto.Unmarshal(response.Body(), &resp)
	return &resp, nil
}

// InsertGroup implements Message
func (m *MessageHttp) InsertGroup(app string, req *rpcc.InsertMessageReq) (*rpcc.InsertMessageResp, error) {
	path := fmt.Sprintf("%s/api/%s/message/group", m.url, app)
	t1 := time.Now()
	body, _ := proto.Marshal(req)

	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.InsertGroup - http status code: %d", response.StatusCode())
	}

	var resp rpcc.InsertMessageResp
	_ = proto.Unmarshal(response.Body(), &resp)
	m.lg.Debug("MessageHttp.InsertGroup", zap.Duration("cost", time.Since(t1)), zap.String("resp", resp.String()))
	return &resp, nil
}

// InsertUser implements Message
func (m *MessageHttp) InsertUser(app string, req *rpcc.InsertMessageReq) (*rpcc.InsertMessageResp, error) {
	path := fmt.Sprintf("%s/api/%s/message/user", m.url, app)
	t1 := time.Now()
	body, _ := proto.Marshal(req)

	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("MessageHttp.InsertUser - http status code: %d", response.StatusCode())
	}

	var resp rpcc.InsertMessageResp
	_ = proto.Unmarshal(response.Body(), &resp)
	m.lg.Debug("MessageHttp.InsertUser", zap.Duration("cost", time.Since(t1)), zap.String("resp", resp.String()))
	return &resp, nil
}

// SetAck implements Message
func (m *MessageHttp) SetAck(app string, req *rpcc.AckMessageReq) error {
	path := fmt.Sprintf("%s/api/%s/message/ack", m.url, app)
	body, _ := proto.Marshal(req)

	response, err := m.Req().SetBody(body).Post(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("MessageHttp.SetAck - http status code: %d", response.StatusCode())
	}

	return nil
}

func (m MessageHttp) Req() *resty.Request {
	if m.srv == nil {
		return m.cli.R()
	}
	return m.cli.R().SetSRV(m.srv)
}
