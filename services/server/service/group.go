package service

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joeyscat/qim/wire/rpc"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Group interface {
	Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error)
	Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error)
	Join(app string, req *rpc.JoinGroupReq) error
	Quit(app string, req *rpc.QuitGroupReq) error
	Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error)
}

type GroupHttp struct {
	url string
	cli *resty.Client
	srv *resty.SRVRecord
	lg  *zap.Logger
}

func NewGroupService(url string, lg *zap.Logger) Group {
	client := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	client.SetHeader("Content-Type", "application/x-protobuf")
	client.SetHeader("Accept", "application/x-protobuf")
	client.SetScheme("http")
	return &GroupHttp{
		url: url,
		cli: client,
		lg:  lg,
	}
}

func NewGroupServiceWithSRV(scheme string, srv *resty.SRVRecord, lg *zap.Logger) Group {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeader("Content-Type", "application/x-protobuf")
	cli.SetHeader("Accept", "application/x-protobuf")
	cli.SetScheme(scheme)
	return &GroupHttp{
		url: "",
		cli: cli,
		srv: srv,
		lg:  lg,
	}
}

// Create implements Group
func (g *GroupHttp) Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group", g.url, app)
	body, _ := proto.Marshal(req)

	response, err := g.Req().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHttp.Create - http status code: %d", response.StatusCode())
	}

	var resp rpc.CreateGroupResp
	_ = proto.Unmarshal(response.Body(), &resp)
	g.lg.Debug("GroupHttp.Create", zap.String("resp", resp.String()))

	return &resp, nil
}

// Detail implements Group
func (g *GroupHttp) Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group", g.url, app)

	response, err := g.Req().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHttp.Detail - http status code: %d", response.StatusCode())
	}

	var resp rpc.GetGroupResp
	_ = proto.Unmarshal(response.Body(), &resp)
	g.lg.Debug("GroupHttp.Detail", zap.String("resp", resp.String()))

	return &resp, nil
}

// Join implements Group
func (g *GroupHttp) Join(app string, req *rpc.JoinGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)

	response, err := g.Req().SetBody(body).Post(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHttp.Join - http status code: %d", response.StatusCode())
	}

	return nil
}

// Members implements Group
func (g *GroupHttp) Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error) {
	path := fmt.Sprintf("%s/api/%s/group/members/%s", g.url, app, req.GetGroupId())

	response, err := g.Req().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHttp.Members - http status code: %d", response.StatusCode())
	}

	var resp rpc.GroupMembersResp
	_ = proto.Unmarshal(response.Body(), &resp)
	g.lg.Debug("GroupHttp.Members", zap.String("resp", resp.String()))

	return &resp, nil
}

// Quit implements Group
func (g *GroupHttp) Quit(app string, req *rpc.QuitGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)

	response, err := g.Req().SetBody(body).Delete(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHttp.Quit - http status code: %d", response.StatusCode())
	}

	return nil
}

func (g *GroupHttp) Req() *resty.Request {
	if g.srv == nil {
		return g.cli.R()
	}
	return g.cli.R().SetSRV(g.srv)
}
