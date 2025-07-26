package server

import (
	"context"

	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/core"
	"github.com/dzjyyds666/vortex/v2"
)

type MediaServer struct {
	ctx context.Context
	v   *vortex.Vortex
}

func NewMediaServer(ctx context.Context, cfg *core.Config) *MediaServer {

	handler := NewStorageHandler()     // 创建handler
	routers := PrepareRouters(handler) // 创建路由
	v := vortex.BootStrap(
		ctx,
		vortex.WithPort(ptr.ToString(cfg.Port)),
		vortex.WithRouters(routers),
	)

	return &MediaServer{ctx: ctx, v: v}
}

// 启动服务
func (s *MediaServer) Start() {
	s.v.Start()
}

// 停止服务
func (s *MediaServer) ShutDown(ctx context.Context) error {
	return nil
}
