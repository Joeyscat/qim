package router

import (
	"context"
	"log"
	"path"
	"strings"

	"github.com/joeyscat/qim/logger"
	"github.com/joeyscat/qim/naming/etcd"
	"github.com/joeyscat/qim/services/router/apis"
	"github.com/joeyscat/qim/services/router/conf"
	"github.com/joeyscat/qim/services/router/ipregion"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type ServerStartOptions struct {
	config string
	data   string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "router",
		Short: "Start a router",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./router/conf.yaml", "config file")
	cmd.PersistentFlags().StringVarP(&opts.data, "data", "d", "./router/data", "data path")

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

	mappings, err := conf.LoadMapping(path.Join(opts.data, "mapping.json"))
	if err != nil {
		return err
	}
	logger.L.Info("load mapping", zap.Any("mapping", mappings))

	regions, err := conf.LoadRegions(path.Join(opts.data, "region.json"))
	if err != nil {
		return err
	}
	logger.L.Info("load region", zap.Any("region", regions))

	region, err := ipregion.NewIP2Region(path.Join(opts.data, "ip2region.db"))
	if err != nil {
		return err
	}

	ns, err := etcd.NewNaming(strings.Split(config.EtcdEndpoints, ","), logger.L.With(zap.String("module", "router.naming")))
	if err != nil {
		return err
	}

	router := apis.RouterApi{
		Naming:   ns,
		IPRegion: region,
		Config: conf.Router{
			Mapping: mappings,
			Regions: regions,
		},
		Lg: logger.L.With(zap.String("module", "router")),
	}

	app := iris.Default()

	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})

	routerAPI := app.Party("/api/lookup")
	{
		routerAPI.Get("/:token", router.Lookup)
	}

	return app.Listen(config.Listen, iris.WithOptimizations)
}
