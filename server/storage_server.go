package server

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dzjyyds666/Allspark-go/conv"
	"github.com/dzjyyds666/Allspark-go/ds"
	"github.com/dzjyyds666/Allspark-go/logx"
	"github.com/dzjyyds666/Allspark-go/ptr"
	"github.com/dzjyyds666/mediaStorage/core"
	"github.com/dzjyyds666/mediaStorage/proto"
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
	var init core.InitUpload
	decoder := json.NewDecoder(ctx.Request().Body)
	if err := decoder.Decode(&init); err != nil {
		logx.Errorf("HandleApplyUpload|ParamsError|decoder err: %v", err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}

	// 开始申请文件信息
	fid, err := s.coreServer.ApplyUpload(ctx.GetContext(), &init)
	if err != nil {
		logx.Errorf("HandleApplyUpload|ApplyUpload|fid: %s|err: %v", fid, err)
		if errors.Is(err, proto.ErrorEnums.ErrFileExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.FileExist), nil)
		}
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.InternalError.WithSubCode(proto.SubStatusCodes.InternalError), nil)
	}
	logx.Infof("HandleApplyUpload|ApplyUpload|fid: %s|fileInfo: %s", fid, conv.ToJsonWithoutError(init))
	return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.FileExist), echo.Map{
		"init_info": init,
	})
}

// 文件直接上传
func (s *StorageServer) HandleSingleUpload(ctx *vortex.Context) error {
	fid := ctx.Param("fid")
	boxId := ctx.QueryParam("box_id")
	file, err := ctx.FormFile("file")
	if err != nil {
		logx.Errorf("HandleSingleUpload|FormFile|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}
	fileOpen, err := file.Open()
	if err != nil {
		logx.Errorf("HandleSingleUpload|Open|fid: %s|err: %v", fid, err)
		return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BadRequest), nil)
	}
	defer fileOpen.Close()

	err = s.coreServer.SingleUpload(ctx.GetContext(), boxId, fid, fileOpen)
	if nil != err {
		logx.Errorf("HandleSingleUpload|SingleUpload|fid: %s|err: %v", fid, err)
		if errors.Is(err, proto.ErrorEnums.ErrBoxNotExist) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.BoxNotExist), nil)
		} else if errors.Is(err, proto.ErrorEnums.ErrNoPrepareFileInfo) {
			return vortex.HttpJsonResponse(ctx, vortex.Statuses.Success.WithSubCode(proto.SubStatusCodes.NoPrepareFileInfo), nil)
		}
	}

	return nil
}
