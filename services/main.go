package main

import (
	"context"
	"flag"
	"log"

	"github.com/joeyscat/qim/services/gateway"
	"github.com/joeyscat/qim/services/router"
	"github.com/joeyscat/qim/services/server"
	"github.com/joeyscat/qim/services/service"
	"github.com/spf13/cobra"
)

const version = "v0.1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "qim",
		Version: version,
		Short:   "QIM is a simple IM system",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCmd(ctx, version))
	root.AddCommand(server.NewServerStartCmd(ctx, version))
	root.AddCommand(service.NewServerStartCmd(ctx, version))
	root.AddCommand(router.NewServerStartCmd(ctx, version))

	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}
