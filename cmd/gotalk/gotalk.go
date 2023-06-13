package main

import (
	"context"
	"flag"
	"github.com/chang144/gotalk/internal/him/services/chatServer"
	"github.com/chang144/gotalk/internal/him/services/gateway"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "him",
		Version: version,
		Short:   "IM Cloud",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCmd(ctx, version))
	root.AddCommand(chatServer.NewServerStartCmd(ctx, version))

	if err := root.Execute(); err != nil {
		return
	}
}
