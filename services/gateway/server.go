package gateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/container"
	"github.com/joeyscat/qim/logger"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/naming/etcd"
	"github.com/joeyscat/qim/services/gateway/conf"
	"github.com/joeyscat/qim/services/gateway/serv"
	"github.com/joeyscat/qim/tcp"
	"github.com/joeyscat/qim/websocket"
	"github.com/joeyscat/qim/wire"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type ServerStartOptions struct {
	config   string
	protocol string
	route    string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./gateway/conf.yaml", "config file")
	cmd.PersistentFlags().StringVarP(&opts.route, "route", "r", "./gateway/route.json", "route file")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of ws or tcp")

	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}

	err = logger.Init(logger.Settings{
		Filename: "./data/gateway.log",
	})
	if err != nil {
		log.Fatal(err)
	}

	logger.L.Debug("load config finished", zap.String("config", config.String()))

	handler := serv.NewHander(config.ServiceID, config.AppSecret, logger.L.With(zap.String("module", "gateway.handler")))

	meta := make(map[string]string)
	meta["domain"] = config.Domain

	var srv qim.Server
	service := &naming.DefaultService{
		ID:       config.ServiceID,
		Name:     config.ServiceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: opts.protocol,
		Tags:     config.Tags,
		Meta:     meta,
	}
	srvOpts := []qim.ServerOption{
		qim.WithConnectionGPool(config.ConnectionGPool),
		qim.WithMessageGPool(config.MessageGPool),
	}

	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service, srvOpts...)
	} else if opts.protocol == "tcp" {
		srv = tcp.NewServer(config.Listen, service, srvOpts...)
	} else {
		return fmt.Errorf("unsupport protocol: %s", opts.protocol)
	}

	srv.SetReadwait(time.Minute * 2)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err = container.Init(srv, logger.L.With(zap.String("module", "gateway.container")), wire.SNChat, wire.SNLogin)
	if err != nil {
		log.Fatal(err)
	}
	container.EnableMonitor(fmt.Sprintf(":%d", config.MonitorPort))

	ns, err := etcd.NewNaming(strings.Split(config.EtcdEndpoints, ","), logger.L.With(zap.String("module", "gateway.naming")))
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)
	container.SetDialer(serv.NewDialer(config.ServiceID))
	selector, err := serv.NewRouteSelector(opts.route, logger.L.With(zap.String("module", "gateway.selector")))
	if err != nil {
		return err
	}
	container.SetSelector(selector)
	return container.Start()
}
