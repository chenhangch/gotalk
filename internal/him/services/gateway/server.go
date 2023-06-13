package gateway

import (
	"context"
	"fmt"

	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/container"
	"github.com/chang144/gotalk/internal/him/naming"
	"github.com/chang144/gotalk/internal/him/naming/consul"

	"time"

	"github.com/chang144/gotalk/internal/him/services/gateway/conf"
	"github.com/chang144/gotalk/internal/him/services/gateway/serv"
	"github.com/chang144/gotalk/internal/him/tcp"
	websocket "github.com/chang144/gotalk/internal/him/websocket"
	"github.com/chang144/gotalk/internal/him/wire"
	"github.com/klintcheng/kim/logger"
	"github.com/spf13/cobra"
)

// ServerStartOptions 服务器启动选项
type ServerStartOptions struct {
	config   string
	protocol string
}

// NewServerStartCmd 创建一个新的http服务命令
func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start a new gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "config", "conf.yaml", "Config file")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of ws or tcp")

	return cmd
}

// RunServerStart 启动服务
func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.InitGateWayConfig(opts.config)
	if err != nil {
		return err
	}
	logger.Init(logger.Settings{
		Level:    "info",
		Filename: "./data/gateway.log",
	})

	handler := &serv.Handler{
		ServiceId: config.ServiceId,
		AppSecret: config.AppSecret,
	}

	var srv him.Server
	service := &naming.RegisterService{
		Id:       config.ServiceId,
		Name:     config.ServiceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: opts.protocol,
		Tags:     config.Tags,
		Meta: map[string]string{
			consul.KeyHealthURL: fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.MonitorPort),
		},
	}

	// 根据protocol建立对应的连接
	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service)
	} else {
		srv = tcp.NewServer(config.Listen, service)
	}

	srv.SetReadWait(time.Minute * 2)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	// container 初始化
	_ = container.InitServer(srv, wire.SNChat, wire.SNLogin)
	container.EnableMonitor(fmt.Sprintf(":%d", config.MonitorPort))

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}
	container.SetServiceNaming(ns)
	// set a dialer
	container.SetDialer(serv.NewTcpDialer(config.ServiceId))

	return container.Start()
}
