package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/container"
	"github.com/joeyscat/qim/logger"
	"github.com/joeyscat/qim/middleware"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/naming/etcd"
	"github.com/joeyscat/qim/services/server/conf"
	"github.com/joeyscat/qim/services/server/handler"
	"github.com/joeyscat/qim/services/server/serv"
	"github.com/joeyscat/qim/services/server/service"
	"github.com/joeyscat/qim/storage"
	"github.com/joeyscat/qim/tcp"
	"github.com/joeyscat/qim/wire"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type ServerStartOptions struct {
	config     string
	serverName string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./server/conf.yaml", "config file")
	cmd.PersistentFlags().StringVarP(&opts.config, "serviceName", "s", "chat", "define a service name, option is login or chat")

	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}

	err = logger.Init(logger.Settings{
		Filename: "./data/server.log",
	})
	if err != nil {
		log.Fatal(err)
	}

	var groupService service.Group
	var messageService service.Message
	if strings.TrimSpace(config.RoyalURL) != "" {
		groupService = service.NewGroupService(config.RoyalURL)
		messageService = service.NewMessageService(config.RoyalURL)
	} else {
		// TODO
	}

	r := qim.NewRouter()
	r.Use(middleware.Recover())

	// login
	loginHandler := handler.NewLoginHandler(logger.L.With(zap.String("module", "login")))
	r.Handle(wire.CommandLoginSignIn, loginHandler.DoSysLogin)
	r.Handle(wire.CommandLoginSignOut, loginHandler.DoSysLogout)
	// talk
	chatHandler := handler.NewChatHandler(messageService, groupService)
	r.Handle(wire.CommandChatUserTalk, chatHandler.DoUserTalk)
	r.Handle(wire.CommandChatGroupTalk, chatHandler.DoGroupTalk)
	r.Handle(wire.CommandChatTalkAck, chatHandler.DoTalkAck)
	// group
	groupHandler := handler.NewGroupHandler(groupService)
	r.Handle(wire.CommandGroupCreate, groupHandler.DoCreate)
	r.Handle(wire.CommandGroupJoin, groupHandler.DoJoin)
	r.Handle(wire.CommandGroupQuit, groupHandler.DoQuit)
	r.Handle(wire.CommandGroupDetail, groupHandler.DoDetail)

	// TODO
	// offline

	rdb, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}
	cache := storage.NewRedisStorage(rdb)
	servhandler := serv.NewServHandler(r, cache,
		logger.L.With(zap.String("module", "service")))

	meta := make(map[string]string)
	meta["zone"] = config.Zone

	service := &naming.DefaultService{
		ID:       config.ServerID,
		Name:     opts.serverName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: string(wire.ProtocolTCP),
		Tags:     config.Tags,
		Meta:     meta,
	}
	srvOpts := []qim.ServerOption{
		qim.WithConnectionGPool(config.ConnectionGPool),
		qim.WithMessageGPool(config.MessageGPool),
	}

	srv := tcp.NewServer(config.Listen, service, srvOpts...)
	srv.SetReadwait(time.Minute * 2)
	srv.SetAcceptor(servhandler)
	srv.SetMessageListener(servhandler)
	srv.SetStateListener(servhandler)

	err = container.Init(srv)
	if err != nil {
		log.Fatal(err)
	}
	container.EnableMonitor(fmt.Sprintf(":%d", config.MonitorPort))

	ns, err := etcd.NewNaming(strings.Split(config.EtcdEndpoints, ","),
		logger.L.With(zap.String("module", "naming")))
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	return container.Start()
}
