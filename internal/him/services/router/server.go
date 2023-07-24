package router

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"net/http"
)

const DefaultPath = "../../internal/him/services/router/conf.yaml"

type ServerStartOption struct {
	config string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOption{}

	cmd := &cobra.Command{
		Use:   "router",
		Short: "start a router",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", DefaultPath, "config file")

	return cmd
}

func RunServerStart(ctx context.Context, optS *ServerStartOption, version string) error {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	return r.Run(":8080")
}
