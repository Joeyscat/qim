package main

import (
	"context"
	"flag"
	"log"

	"github.com/joeyscat/qim/examples/mock"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "qim",
		Version: version,
		Short:   "tools",
	}

	ctx := context.Background()

	// mock
	root.AddCommand(mock.NewServerCmd(ctx))
	root.AddCommand(mock.NewClientCmd(ctx))

	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}
