package server

import (
	"context"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/core"
	"github.com/dzjyyds666/vortex/v2"
	"github.com/labstack/echo/v4"
)

type StorageServer struct {
	ctx        context.Context
	v          *vortex.Vortex
	coreServer *core.StorageCoreServer
}

func NewStorageServer(ctx context.Context, cfg *core.Config, dsServer *ds.DatabaseServer) *StorageServer {

	s3Server := core.NewS3Server(ctx, cfg)
	fileIndexServer := core.NewFileIndexServer(ctx, cfg, dsServer, s3Server)

	server := &StorageServer{
		ctx:        ctx,
		coreServer: core.NewStorageCoreServer(ctx, cfg, fileIndexServer),
	}
	routers := PrepareRouters(server) // 创建路由
	v := vortex.BootStrap(
		ctx,
		vortex.WithPort(ptr.ToString(cfg.Port)),
		vortex.WithRouters(routers),
	)
	server.v = v

	return server
}

// 启动服务
func (s *StorageServer) Start() {
	s.v.Start()
}

// 停止服务
func (s *StorageServer) ShutDown(ctx context.Context) error {
	return nil
}

// HandleRouterTest 测试路由
func (s *StorageServer) HandleRouterTest(ctx *vortex.Context) error {
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success, echo.Map{"message": "hello world"})
}
