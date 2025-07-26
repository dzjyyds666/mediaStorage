package server

import (
	"context"

	"github.com/dzjyyds666/vortex/v2"
)

type MediaServer struct {
	ctx context.Context
	v   *vortex.Vortex
}

func NewMediaServer(ctx context.Context) *MediaServer {

	handler := NewStorageHandler()     // 创建handler
	routers := PrepareRouters(handler) // 创建路由
	v := vortex.BootStrap(
		ctx,
		vortex.WithPort("9000"),
		vortex.WithRouters(routers),
	)

	return &MediaServer{ctx: ctx, v: v}
}

func (s *MediaServer) Start() {

}
