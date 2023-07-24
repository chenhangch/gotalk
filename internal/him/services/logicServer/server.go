package logicServer

import (
	"context"
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/container"
	"github.com/chang144/gotalk/internal/him/naming"
	"github.com/chang144/gotalk/internal/him/naming/consul"
	"github.com/chang144/gotalk/internal/him/services/logicServer/conf"
	"github.com/chang144/gotalk/internal/him/services/logicServer/handler"
	"github.com/chang144/gotalk/internal/him/services/logicServer/serv"
	"github.com/chang144/gotalk/internal/him/storage"
	"github.com/chang144/gotalk/internal/him/tcp"
	"github.com/chang144/gotalk/internal/him/wire"
	"github.com/spf13/cobra"
)

const DefaultPath = "../../internal/him/services/logicServer/conf.yaml"

// ServerStarOptions TODO: 这里是chat以及login逻辑服务器
type ServerStarOptions struct {
	config      string
	serviceName string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStarOptions{}

	cmd := &cobra.Command{
		Use:   "logicServer",
		Short: "Start a chat server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", DefaultPath, "Config file")
	cmd.PersistentFlags().StringVarP(&opts.config, "serverName", "s", "chat", "define a service name, option is login or chat")

	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStarOptions, version string) error {
	config, err := conf.InitLogicConfig(opts.config)
	if err != nil {
		return err
	}

	r := him.NewRouter()
	// login
	loginHandler := handler.NewLoginHandler()
	r.AddHandles(wire.CommandLoginSignIn, loginHandler.DoSysLogin)
	r.AddHandles(wire.CommandLoginSignOut, loginHandler.DoSysLogout)

	rdb, err := storage.InitRedis(config.RedisAddr, "")
	if err != nil {
		return err
	}
	cache := storage.NewRedisStorage(rdb)
	h := serv.NewLogicHandler(r, cache)

	rService := &naming.RegisterService{
		Id:       config.ServerId,
		Name:     opts.serviceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: string(wire.ProtocolTCP),
		Tags:     config.Tags,
	}
	tSrv := tcp.NewServer(config.Listen, rService)

	tSrv.SetReadWait(him.DefaultReadWait)
	tSrv.SetAcceptor(h)
	tSrv.SetMessageListener(h)
	tSrv.SetStateListener(h)

	if err := container.Init(tSrv); err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulRUL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)

	return container.Start()
}
