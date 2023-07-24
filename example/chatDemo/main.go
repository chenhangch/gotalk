package main

import (
	"context"
	"flag"
	"github.com/chang144/gotalk/example/chatDemo/client"

	"github.com/chang144/gotalk/example/chatDemo/serv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := cobra.Command{
		Use:     "chat",
		Version: version,
		Short:   "chat logicServer",
	}
	ctx := context.Background()

	root.AddCommand(serv.NewServerStartCmd(ctx, version))
	root.AddCommand(client.NewCmd(ctx))

	if err := root.Execute(); err != nil {
		logrus.WithError(err).Fatal("Could not run command")
	}
}
