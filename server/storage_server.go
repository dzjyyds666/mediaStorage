package server

import (
	"context"
	"encoding/json"

	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/core"
	"github.com/dzjyyds666/vortex/v2"
)

type StorageServer struct {
	ctx        context.Context
	v          *vortex.Vortex
	coreServer *core.StorageCoreServer
}

func NewStorageServer(ctx context.Context, cfg *core.Config, dsServer *ds.DatabaseServer) *StorageServer {

	s3Server := core.NewS3Server(ctx, cfg)
	boxServer := core.NewBoxServer(ctx, cfg, dsServer)
	depotServer := core.NewDepotServer(ctx, cfg, dsServer, boxServer)
	fileIndexServer := core.NewFileIndexServer(ctx, cfg, dsServer, s3Server, boxServer, depotServer)

	server := &StorageServer{
		ctx:        ctx,
		coreServer: core.NewStorageCoreServer(ctx, cfg, fileIndexServer, boxServer, depotServer),
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

// 申请上传
func (s *StorageServer) HandleApplyUpload(ctx *vortex.Context) error {
	var init core.MediaFileInfo
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&init); err != nil {
		logx.Errorf("HandleApplyUpload|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.ParamsInvaild, nil)
	}

	// 开始申请文件信息
	s.coreServer.ApplyUpload(ctx.GetContext(), &init)

	return nil
}
