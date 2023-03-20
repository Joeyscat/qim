package service

import (
	"context"
	"hash/crc32"
	"log"
	"strings"

	"github.com/joeyscat/qim/logger"
	"github.com/joeyscat/qim/naming"
	"github.com/joeyscat/qim/naming/etcd"
	"github.com/joeyscat/qim/services/service/conf"
	"github.com/joeyscat/qim/services/service/database"
	"github.com/joeyscat/qim/services/service/handler"
	"github.com/joeyscat/qim/wire"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ServerStartOptions struct {
	config string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "royal",
		Short: "Start a rpc service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./service/conf.yaml", "config file")
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
	logger.L.Debug("load config finished", zap.String("config", config.String()))

	var (
		baseDB    *gorm.DB
		messageDB *gorm.DB
	)
	baseDB, err = database.InitDB(config.Driver, config.BaseDB)
	if err != nil {
		return err
	}
	messageDB, err = database.InitDB(config.Driver, config.MessageDB)
	if err != nil {
		return err
	}

	_ = baseDB.AutoMigrate(&database.Group{}, &database.GroupMember{})
	_ = baseDB.AutoMigrate(&database.MessageIndex{}, &database.MessageContent{})

	if config.NodeID == 0 {
		config.NodeID = int64(HashCode(config.ServiceID))
	}

	idgen, err := database.NewIDGenerator(config.NodeID)
	if err != nil {
		return err
	}

	rdb, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}

	ns, err := etcd.NewNaming(strings.Split(config.EtcdEndpoints, ","),
		logger.L.With(zap.String("module", "service.naming")))
	if err != nil {
		return err
	}
	_ = ns.Register(&naming.DefaultService{
		ID:       config.ServiceID,
		Name:     wire.SNService,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: "http",
		Tags:     config.Tags,
		Meta:     map[string]string{},
	})

	defer func() {
		_ = ns.Deregister(config.ServiceID)
	}()

	serviceHandler := handler.ServiceHandler{
		BaseDB:    baseDB,
		MessageDB: messageDB,
		IDgen:     idgen,
		Cache:     rdb,
	}

	ac := conf.MakeAccessLog()
	defer ac.Close()

	app := newApp(&serviceHandler)
	app.UseRouter(ac.Handler)
	app.UseRouter(setAllowedResponses)

	return app.Listen(config.Listen, iris.WithOptimizations)
}

func newApp(serviceHandler *handler.ServiceHandler) *iris.Application {
	app := iris.Default()

	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})
	messageAPI := app.Party("/api/:app/message")
	{
		messageAPI.Post("/user", serviceHandler.InsertUserMessage)
		messageAPI.Post("/group", serviceHandler.InsertGroupMessage)
		messageAPI.Get("/ack", serviceHandler.MessageAck)
	}

	groupAPI := app.Party("/api/:app/group")
	{
		groupAPI.Get("/:id", serviceHandler.GroupGet)
		groupAPI.Post("/", serviceHandler.GroupCreate)
		groupAPI.Post("/member", serviceHandler.GroupJoin)
		groupAPI.Delete("/member", serviceHandler.GroupQuit)
		groupAPI.Get("/members/:id", serviceHandler.GroupMembers)
	}

	offlineAPI := app.Party("/api/:app/offline")
	{
		offlineAPI.Use(iris.Compression)
		offlineAPI.Get("/index", serviceHandler.GetOfflineMessageIndex)
		offlineAPI.Get("/content", serviceHandler.GetOfflineMessageContent)
	}

	return app
}

func setAllowedResponses(ctx iris.Context) {
	ctx.Negotiation().JSON().Protobuf().MsgPack()

	ctx.Negotiation().Accept.JSON()

	ctx.Next()
}

func HashCode(key string) uint32 {
	hash32 := crc32.NewIEEE()
	hash32.Write([]byte(key))
	return hash32.Sum32() % 1000
}
