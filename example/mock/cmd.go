package mock

import (
	"context"
	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

type StartOptions struct {
	addr     string
	protocol string
}

func NewClientCmd(ctx context.Context) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "mock_cli",
		Short: "Start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runcli(ctx, opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.addr, "addr", "a", "ws://localhost:8000", "logicServer address")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol ws or tcp")
	return cmd
}

func runcli(ctx context.Context, opts *StartOptions) error {
	cli := ClientDemo{}
	cli.Start(ksuid.New().String(), opts.protocol, opts.addr)
	return nil
}

func NewServerCmd(ctx context.Context) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "mock_server",
		Short: "Start logicServer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runsrv(ctx, opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.addr, "addr", "a", ":8000", "listen address")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol ws or tcp")
	return cmd
}

func runsrv(ctx context.Context, opts *StartOptions) error {
	srv := &ServerDemo{}
	srv.Start("srv1", opts.protocol, opts.addr)
	return nil
}
