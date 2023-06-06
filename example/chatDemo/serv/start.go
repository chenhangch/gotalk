package serv

import (
	"context"

	"github.com/spf13/cobra"
)

type ServerStartOptions struct {
	id     string
	listen string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "start a chat chatServer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.id, "serverid", "i", "demo", "chatServer id")
	cmd.PersistentFlags().StringVarP(&opts.listen, "listen", "l", ":8000", "listen address to chat chatServer")

	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	server := NewServer(opts.id, opts.listen)
	defer server.Shutdown()
	return server.Start()
}
